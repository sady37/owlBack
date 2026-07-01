# case-d5f7-0630-19021918 — 每 tick belief 时间线 (room fd00:0:3:111:3:300, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
19:02:13 D5F7.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:45 D5F7.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:48 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:02:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:17 D5F7.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:20 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:35 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:36 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:37 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:38 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:39 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:40 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:41 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:42 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:43 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:44 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:45 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:46 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:47 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:48 D5F7.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:49 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:50 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:51 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:52 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:53 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:54 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:55 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:56 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:57 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:58 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:03:59 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:00 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:01 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:02 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:03 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:04 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:05 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:06 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:07 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:08 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:09 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:10 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:11 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:12 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:13 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:14 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:15 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:16 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:17 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:18 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:19 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:20 D5F7.88  -             88      -    -        no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:21 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:22 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:23 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:24 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:25 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:26 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:27 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:28 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:29 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:30 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:31 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:32 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:33 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:34 -.-      -             -       -    -        (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:35 D5F7.E   -             -       0    -        np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:36 D5F7.E0  -             -       0    -        EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
19:04:36 D5F7.0   D5F700436288  stand   83   -        stand              trk  1.00 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
19:04:36 D5F7.0   D5F700436288  stand   95   -        stand              trk  1.00 Empty      1   0     0.00  0.03  0.35  0.00  0.54  0.02
19:04:37 D5F7.0   D5F700436288  stand   78   -        stand              trk  1.00 OpenFloor  1   1     0.00  0.04  0.42  0.01  0.41  0.02
19:04:38 D5F7.0   D5F700436288  stand   86   -        stand              trk  1.00 OpenFloor  1   2     0.00  0.04  0.48  0.01  0.30  0.02
19:04:39 D5F7.0   D5F700436288  stand   61   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.53  0.01  0.22  0.03
19:04:40 D5F7.0   D5F700436288  walk    89   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.56  0.01  0.16  0.03
19:04:41 D5F7.0   D5F700436288  walk    62   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.58  0.02  0.12  0.03
19:04:42 D5F7.0   D5F700436288  walk    68   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.59  0.02  0.10  0.03
19:04:43 D5F7.0   D5F700436288  walk    85   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.60  0.02  0.08  0.03
19:04:44 D5F7.0   D5F700436288  walk    58   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.07  0.03
19:04:45 D5F7.0   D5F700436288  walk    63   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.07  0.03
19:04:46 D5F7.0   D5F700436288  walk    72   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
19:04:47 D5F7.0   D5F700436288  walk    61   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
19:04:48 D5F7.0   D5F700436288  walk    63   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
19:04:49 D5F7.0   D5F700436288  walk    63   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
19:04:50 D5F7.0   D5F700436288  walk    73   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:04:51 D5F7.0   D5F700436288  sit     82   -        sit                trk  1.00 OpenFloor  1   0     0.00  0.05  0.58  0.02  0.05  0.03
19:04:52 D5F7.0   D5F700436288  sit     47   -        sit                trk  1.00 OpenFloor  1   0     0.01  0.07  0.37  0.02  0.06  0.04
19:04:53 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.06  0.25  0.03  0.08  0.03
19:04:54 D5F7.0   D5F700436288  sit     71   -        sit                trk  1.00 Sit        1   0     0.01  0.05  0.18  0.04  0.08  0.02
19:04:55 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.04  0.12  0.04  0.05  0.02
19:04:56 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.03  0.10  0.04  0.04  0.02
19:04:57 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.03  0.10  0.04  0.04  0.02
19:04:58 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.04  0.03  0.02
19:04:59 D5F7.0   D5F700436288  sit     83   -        sit                trk  1.00 Sit        1   0     0.00  0.02  0.17  0.03  0.03  0.02
19:05:00 D5F7.0   D5F700436288  sit     57   -        sit                trk  1.00 Sit        1   0     0.00  0.02  0.12  0.03  0.02  0.02
19:05:01 D5F7.0   D5F700436288  sit     75   -        sit                trk  1.00 Sit        1   0     0.00  0.02  0.10  0.03  0.03  0.02
19:05:02 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.00  0.02  0.10  0.03  0.02  0.02
19:05:03 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:04 D5F7.0   D5F700436288  sit     78   -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:05 D5F7.0   D5F700436288  sit     69   -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:06 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:07 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:08 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:09 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:10 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:11 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:12 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:13 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:14 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:15 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:16 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:17 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:18 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:19 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:20 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:21 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:22 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:23 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:24 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:25 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:26 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:27 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:28 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:29 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:30 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:31 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:32 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:33 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:34 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:35 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:36 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   28    0.01  0.02  0.09  0.03  0.02  0.02
19:05:37 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   29    0.01  0.02  0.09  0.03  0.02  0.02
19:05:38 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   30    0.01  0.02  0.09  0.03  0.02  0.02
19:05:39 D5F7.0   D5F700436288  sit     60   -        sit                trk  1.00 Sit        1   31    0.01  0.02  0.09  0.03  0.02  0.02
19:05:40 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   32    0.01  0.02  0.09  0.03  0.02  0.02
19:05:40 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 Sit        1   33    0.01  0.02  0.09  0.03  0.02  0.02
19:05:41 -.-      -             -       -    -        (no frame, held)   room -    Sit        1   33    0.01  0.02  0.09  0.03  0.02  0.02
19:05:42 D5F7.0   D5F700436288  sit     73   -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:43 D5F7.0   D5F700436288  sit     69   -        sit                trk  1.00 Sit        1   0     0.01  0.02  0.09  0.03  0.02  0.02
19:05:44 D5F7.0   D5F700436288  walk    89   -        walk               trk  1.00 Sit        1   0     0.00  0.03  0.25  0.05  0.03  0.03
19:05:45 D5F7.0   D5F700436288  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.04  0.38  0.05  0.04  0.03
19:05:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.47  0.04  0.05  0.03
19:05:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.53  0.04  0.05  0.03
19:05:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.56  0.03  0.05  0.03
19:05:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.58  0.03  0.05  0.03
19:05:49 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.58  0.03  0.05  0.03
19:05:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.59  0.02  0.05  0.03
19:05:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.05  0.03
19:05:51 D5F7.0   D5F700436288  stand   38   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.05  0.03
19:05:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:53 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:05:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:01 D5F7.0   D5F700436288  stand   90   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:05 D5F7.0   D5F700436288  walk    94   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:06 D5F7.0   D5F700436288  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:07 D5F7.0   D5F700436288  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:11 D5F7.0   D5F700436288  stand   90   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:12 D5F7.0   D5F700436288  stand   91   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:06:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:06:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.61  0.02  0.05  0.03
19:06:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.61  0.02  0.05  0.03
19:06:44 D5F7.0   D5F700436288  stand   105  -        stand              trk  1.00 OpenFloor  1   30    0.00  0.05  0.61  0.02  0.05  0.03
19:06:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.61  0.02  0.05  0.03
19:06:46 D5F7.0   D5F700436288  stand   89   -        stand              trk  1.00 OpenFloor  1   32    0.00  0.05  0.61  0.02  0.05  0.03
19:06:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   33    0.01  0.05  0.61  0.02  0.05  0.03
19:06:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   34    0.01  0.05  0.61  0.02  0.05  0.03
19:06:49 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   35    0.01  0.05  0.61  0.02  0.05  0.03
19:06:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   36    0.01  0.05  0.61  0.02  0.05  0.03
19:06:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   37    0.01  0.05  0.61  0.02  0.05  0.03
19:06:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   38    0.01  0.05  0.61  0.02  0.05  0.03
19:06:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   39    0.01  0.05  0.61  0.02  0.05  0.03
19:06:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   40    0.01  0.05  0.61  0.02  0.05  0.03
19:06:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   41    0.01  0.05  0.61  0.02  0.05  0.03
19:06:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   42    0.01  0.05  0.61  0.02  0.05  0.03
19:06:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   43    0.01  0.05  0.61  0.02  0.05  0.03
19:06:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   44    0.01  0.05  0.61  0.02  0.05  0.03
19:06:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   45    0.01  0.05  0.61  0.02  0.05  0.03
19:07:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   46    0.01  0.05  0.61  0.02  0.05  0.03
19:07:01 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   47    0.01  0.05  0.61  0.02  0.05  0.03
19:07:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   48    0.01  0.05  0.61  0.02  0.05  0.03
19:07:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   49    0.01  0.05  0.61  0.02  0.05  0.03
19:07:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   50    0.01  0.05  0.61  0.02  0.05  0.03
19:07:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   51    0.01  0.05  0.61  0.02  0.05  0.03
19:07:06 D5F7.0   D5F700436288  stand   94   -        stand              trk  1.00 OpenFloor  1   52    0.00  0.05  0.61  0.02  0.05  0.03
19:07:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   53    0.01  0.05  0.61  0.02  0.05  0.03
19:07:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   54    0.01  0.05  0.61  0.02  0.05  0.03
19:07:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   55    0.01  0.05  0.61  0.02  0.05  0.03
19:07:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   56    0.01  0.05  0.61  0.02  0.05  0.03
19:07:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   57    0.01  0.05  0.61  0.02  0.05  0.03
19:07:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   58    0.01  0.05  0.61  0.02  0.05  0.03
19:07:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   59    0.01  0.05  0.61  0.02  0.05  0.03
19:07:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   60    0.01  0.05  0.61  0.02  0.05  0.03
19:07:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   61    0.01  0.05  0.61  0.02  0.05  0.03
19:07:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   62    0.01  0.05  0.61  0.02  0.05  0.03
19:07:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   63    0.01  0.05  0.61  0.02  0.05  0.03
19:07:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   64    0.01  0.05  0.61  0.02  0.05  0.03
19:07:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   65    0.01  0.05  0.61  0.02  0.05  0.03
19:07:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   66    0.01  0.05  0.61  0.02  0.05  0.03
19:07:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   67    0.01  0.05  0.61  0.02  0.05  0.03
19:07:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   68    0.01  0.05  0.61  0.02  0.05  0.03
19:07:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   69    0.01  0.05  0.61  0.02  0.05  0.03
19:07:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   70    0.01  0.05  0.61  0.02  0.05  0.03
19:07:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   71    0.01  0.05  0.61  0.02  0.05  0.03
19:07:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   72    0.01  0.05  0.61  0.02  0.05  0.03
19:07:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   73    0.01  0.05  0.61  0.02  0.05  0.03
19:07:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   74    0.01  0.05  0.61  0.02  0.05  0.03
19:07:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   75    0.01  0.05  0.61  0.02  0.05  0.03
19:07:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   76    0.01  0.05  0.61  0.02  0.05  0.03
19:07:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   77    0.01  0.05  0.61  0.02  0.05  0.03
19:07:31 D5F7.E   -             -       0    -        np=2               room -    OpenFloor  1   77    0.01  0.05  0.61  0.02  0.05  0.03
19:07:31 D5F7.E1  -             -       0    -        EnterRoom(rdr)     room -    OpenFloor  1   77    0.01  0.05  0.61  0.02  0.05  0.03
19:07:31 D5F7.1   D5F710731840  stand   115  -        stand              trk  1.00 OpenFloor  2   78    0.00  0.01  0.58  0.00  0.39  0.02
19:07:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   78    0.01  0.03  0.76  0.01  0.03  0.02
19:07:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   78    0.00  0.03  0.82  0.01  0.02  0.02
19:07:32 D5F7.1   D5F710731840  stand   99   -        stand              trk  1.00 OpenFloor  2   78    0.00  0.01  0.92  0.00  0.05  0.01
19:07:33 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   79    0.00  0.01  0.96  0.00  0.01  0.01
19:07:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   79    0.00  0.02  0.84  0.00  0.01  0.02
19:07:34 D5F7.1   D5F710731840  stand   111  -        stand              trk  1.00 OpenFloor  2   80    0.00  0.01  0.97  0.00  0.00  0.01
19:07:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   80    0.00  0.02  0.84  0.00  0.01  0.02
19:07:35 D5F7.0   D5F700436288  stand   104  -        stand              trk  1.00 OpenFloor  2   81    0.00  0.02  0.85  0.00  0.01  0.02
19:07:35 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   81    0.00  0.01  0.94  0.00  0.00  0.01
19:07:36 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   82    0.00  0.01  0.97  0.00  0.00  0.01
19:07:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   82    0.00  0.02  0.85  0.00  0.01  0.02
19:07:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   83    0.00  0.02  0.85  0.00  0.01  0.02
19:07:37 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   83    0.00  0.01  0.97  0.00  0.00  0.01
19:07:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   84    0.00  0.02  0.85  0.00  0.01  0.02
19:07:38 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   84    0.00  0.01  0.97  0.00  0.00  0.01
19:07:39 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   85    0.00  0.01  0.97  0.00  0.00  0.01
19:07:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   85    0.00  0.02  0.85  0.00  0.01  0.02
19:07:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   86    0.00  0.02  0.85  0.00  0.01  0.02
19:07:40 D5F7.1   D5F710731840  stand   0    -        stand              trk  1.00 OpenFloor  2   86    0.00  0.01  0.97  0.00  0.00  0.01
19:07:41 D5F7.E1  D5F710731840  -       0    -        ExitRoom(rdr)      trk  1.00 OpenFloor  2   86    0.00  0.01  0.97  0.00  0.00  0.01
19:07:41 D5F7.1   -             stand   0    -        stand              room -    OpenFloor  1   87    0.00  0.02  0.85  0.00  0.01  0.02
19:07:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   87    0.00  0.02  0.85  0.00  0.01  0.02
19:07:42 D5F7.E   -             -       0    -        np=1               room -    OpenFloor  1   87    0.00  0.02  0.85  0.00  0.01  0.02
19:07:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   88    0.00  0.04  0.74  0.00  0.02  0.04
19:07:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   89    0.01  0.05  0.68  0.01  0.03  0.04
19:07:44 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   90    0.01  0.05  0.65  0.01  0.04  0.03
19:07:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   91    0.01  0.05  0.63  0.01  0.04  0.03
19:07:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   92    0.01  0.05  0.62  0.02  0.05  0.03
19:07:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   93    0.01  0.05  0.62  0.02  0.05  0.03
19:07:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   94    0.01  0.05  0.62  0.02  0.05  0.03
19:07:49 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   95    0.01  0.05  0.61  0.02  0.05  0.03
19:07:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   96    0.01  0.05  0.61  0.02  0.05  0.03
19:07:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   97    0.01  0.05  0.61  0.02  0.05  0.03
19:07:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   98    0.01  0.05  0.61  0.02  0.05  0.03
19:07:52 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   98    0.01  0.05  0.61  0.02  0.05  0.03
19:07:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   99    0.01  0.05  0.61  0.02  0.05  0.03
19:07:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   100   0.01  0.05  0.61  0.02  0.05  0.03
19:07:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   101   0.01  0.05  0.61  0.02  0.05  0.03
19:07:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   102   0.01  0.05  0.61  0.02  0.05  0.03
19:07:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   103   0.01  0.05  0.61  0.02  0.05  0.03
19:07:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   104   0.01  0.05  0.61  0.02  0.05  0.03
19:07:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   105   0.01  0.05  0.61  0.02  0.05  0.03
19:07:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   106   0.01  0.05  0.61  0.02  0.05  0.03
19:08:00 D5F7.0   D5F700436288  stand   71   -        stand              trk  1.00 OpenFloor  1   107   0.01  0.05  0.61  0.02  0.05  0.03
19:08:01 D5F7.0   D5F700436288  stand   93   -        stand              trk  1.00 OpenFloor  1   108   0.00  0.05  0.58  0.02  0.05  0.08
19:08:02 D5F7.0   D5F700436288  stand   99   -        stand              trk  1.00 OpenFloor  1   109   0.00  0.05  0.59  0.02  0.08  0.04
19:08:03 D5F7.0   D5F700436288  stand   68   -        stand              trk  1.00 OpenFloor  1   110   0.00  0.05  0.55  0.02  0.07  0.11
19:08:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   111   0.01  0.05  0.57  0.02  0.11  0.04
19:08:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   112   0.01  0.05  0.52  0.02  0.08  0.14
19:08:06 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   113   0.01  0.04  0.45  0.01  0.12  0.21
19:08:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   114   0.01  0.03  0.39  0.01  0.17  0.26
19:08:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   115   0.00  0.03  0.32  0.01  0.21  0.31
19:08:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Left       1   116   0.00  0.02  0.26  0.01  0.25  0.36
19:08:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Left       1   117   0.00  0.02  0.20  0.01  0.27  0.43
19:08:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Left       1   118   0.00  0.01  0.16  0.00  0.30  0.47
19:08:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Left       1   119   0.00  0.01  0.13  0.00  0.33  0.48
19:08:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Left       1   120   0.00  0.01  0.10  0.00  0.36  0.49
19:08:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Empty      1   0     0.00  0.02  0.17  0.00  0.68  0.07
19:08:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Empty      1   0     0.00  0.02  0.27  0.01  0.60  0.02
19:08:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Empty      1   0     0.00  0.03  0.36  0.01  0.47  0.02
19:08:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.04  0.44  0.01  0.35  0.02
19:08:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.04  0.49  0.01  0.25  0.03
19:08:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.53  0.02  0.19  0.03
19:08:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.56  0.02  0.14  0.03
19:08:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.58  0.02  0.11  0.03
19:08:22 D5F7.0   D5F700436288  stand   101  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.59  0.02  0.09  0.03
19:08:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.08  0.03
19:08:24 D5F7.0   D5F700436288  stand   78   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.60  0.02  0.07  0.03
19:08:25 D5F7.0   D5F700436288  stand   89   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
19:08:26 D5F7.0   D5F700436288  stand   77   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.06  0.03
19:08:27 D5F7.0   D5F700436288  stand   111  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.06  0.03
19:08:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.06  0.03
19:08:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:08:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:08:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:08:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.61  0.02  0.05  0.03
19:08:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.61  0.02  0.05  0.03
19:08:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   30    0.01  0.05  0.61  0.02  0.05  0.03
19:08:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.61  0.02  0.05  0.03
19:08:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   32    0.01  0.05  0.61  0.02  0.05  0.03
19:08:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   33    0.01  0.05  0.61  0.02  0.05  0.03
19:08:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   34    0.01  0.05  0.61  0.02  0.05  0.03
19:08:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   35    0.01  0.05  0.61  0.02  0.05  0.03
19:08:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   36    0.01  0.05  0.61  0.02  0.05  0.03
19:08:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   37    0.01  0.05  0.61  0.02  0.05  0.03
19:08:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   38    0.01  0.05  0.61  0.02  0.05  0.03
19:08:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   39    0.01  0.05  0.61  0.02  0.05  0.03
19:08:44 D5F7.0   D5F700436288  stand   98   -        stand              trk  1.00 OpenFloor  1   40    0.00  0.05  0.61  0.02  0.05  0.03
19:08:45 D5F7.0   D5F700436288  stand   99   -        stand              trk  1.00 OpenFloor  1   41    0.00  0.05  0.61  0.02  0.05  0.03
19:08:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   42    0.01  0.05  0.61  0.02  0.05  0.03
19:08:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   43    0.01  0.05  0.61  0.02  0.05  0.03
19:08:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   44    0.01  0.05  0.61  0.02  0.05  0.03
19:08:49 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   45    0.01  0.05  0.61  0.02  0.05  0.03
19:08:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   46    0.01  0.05  0.61  0.02  0.05  0.03
19:08:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   47    0.01  0.05  0.61  0.02  0.05  0.03
19:08:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   48    0.01  0.05  0.61  0.02  0.05  0.03
19:08:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   49    0.01  0.05  0.61  0.02  0.05  0.03
19:08:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   50    0.01  0.05  0.61  0.02  0.05  0.03
19:08:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   51    0.01  0.05  0.61  0.02  0.05  0.03
19:08:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   52    0.01  0.05  0.61  0.02  0.05  0.03
19:08:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   53    0.01  0.05  0.61  0.02  0.05  0.03
19:08:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   54    0.01  0.05  0.61  0.02  0.05  0.03
19:08:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   55    0.01  0.05  0.61  0.02  0.05  0.03
19:09:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   56    0.01  0.05  0.61  0.02  0.05  0.03
19:09:01 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   57    0.01  0.05  0.61  0.02  0.05  0.03
19:09:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   58    0.01  0.05  0.61  0.02  0.05  0.03
19:09:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   59    0.01  0.05  0.61  0.02  0.05  0.03
19:09:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   60    0.01  0.05  0.61  0.02  0.05  0.03
19:09:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   61    0.01  0.05  0.61  0.02  0.05  0.03
19:09:06 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   62    0.01  0.05  0.61  0.02  0.05  0.03
19:09:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   63    0.01  0.05  0.61  0.02  0.05  0.03
19:09:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   64    0.01  0.05  0.61  0.02  0.05  0.03
19:09:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   65    0.01  0.05  0.61  0.02  0.05  0.03
19:09:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   66    0.01  0.05  0.61  0.02  0.05  0.03
19:09:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   67    0.01  0.05  0.61  0.02  0.05  0.03
19:09:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   68    0.01  0.05  0.61  0.02  0.05  0.03
19:09:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   69    0.01  0.05  0.61  0.02  0.05  0.03
19:09:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   70    0.01  0.05  0.61  0.02  0.05  0.03
19:09:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   71    0.01  0.05  0.61  0.02  0.05  0.03
19:09:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   72    0.01  0.05  0.61  0.02  0.05  0.03
19:09:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   73    0.01  0.05  0.61  0.02  0.05  0.03
19:09:18 D5F7.0   D5F700436288  stand   101  -        stand              trk  1.00 OpenFloor  1   74    0.00  0.05  0.61  0.02  0.05  0.03
19:09:19 D5F7.0   D5F700436288  stand   115  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:28 D5F7.0   D5F700436288  stand   98   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:44 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:47 D5F7.0   D5F700436288  stand   101  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:49 D5F7.0   D5F700436288  stand   109  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:51 D5F7.0   D5F700436288  stand   101  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:52 D5F7.0   D5F700436288  stand   102  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:09:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:55 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:09:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:01 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:06 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:10:17 D5F7.0   D5F700436288  stand   98   -        stand              trk  1.00 OpenFloor  1   27    0.00  0.05  0.61  0.02  0.05  0.03
19:10:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.61  0.02  0.05  0.03
19:10:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.61  0.02  0.05  0.03
19:10:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   30    0.01  0.05  0.61  0.02  0.05  0.03
19:10:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.61  0.02  0.05  0.03
19:10:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   32    0.01  0.05  0.61  0.02  0.05  0.03
19:10:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   33    0.01  0.05  0.61  0.02  0.05  0.03
19:10:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   34    0.01  0.05  0.61  0.02  0.05  0.03
19:10:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   35    0.01  0.05  0.61  0.02  0.05  0.03
19:10:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   36    0.01  0.05  0.61  0.02  0.05  0.03
19:10:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   37    0.01  0.05  0.61  0.02  0.05  0.03
19:10:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   38    0.01  0.05  0.61  0.02  0.05  0.03
19:10:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   39    0.01  0.05  0.61  0.02  0.05  0.03
19:10:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   40    0.01  0.05  0.61  0.02  0.05  0.03
19:10:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   41    0.01  0.05  0.61  0.02  0.05  0.03
19:10:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   42    0.01  0.05  0.61  0.02  0.05  0.03
19:10:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   43    0.01  0.05  0.61  0.02  0.05  0.03
19:10:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   44    0.01  0.05  0.61  0.02  0.05  0.03
19:10:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   45    0.01  0.05  0.61  0.02  0.05  0.03
19:10:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   46    0.01  0.05  0.61  0.02  0.05  0.03
19:10:37 D5F7.0   D5F700436288  stand   98   -        stand              trk  1.00 OpenFloor  1   47    0.00  0.05  0.61  0.02  0.05  0.03
19:10:38 D5F7.0   D5F700436288  stand   81   -        stand              trk  1.00 OpenFloor  1   48    0.00  0.05  0.61  0.02  0.05  0.03
19:10:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   49    0.01  0.05  0.61  0.02  0.05  0.03
19:10:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   50    0.01  0.05  0.61  0.02  0.05  0.03
19:10:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   51    0.01  0.05  0.61  0.02  0.05  0.03
19:10:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   52    0.01  0.05  0.61  0.02  0.05  0.03
19:10:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   53    0.01  0.05  0.61  0.02  0.05  0.03
19:10:44 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   54    0.01  0.05  0.61  0.02  0.05  0.03
19:10:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   55    0.01  0.05  0.61  0.02  0.05  0.03
19:10:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   56    0.01  0.05  0.61  0.02  0.05  0.03
19:10:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   57    0.01  0.05  0.61  0.02  0.05  0.03
19:10:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   58    0.01  0.05  0.61  0.02  0.05  0.03
19:10:49 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   59    0.01  0.05  0.61  0.02  0.05  0.03
19:10:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   60    0.01  0.05  0.61  0.02  0.05  0.03
19:10:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   61    0.01  0.05  0.61  0.02  0.05  0.03
19:10:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   62    0.01  0.05  0.61  0.02  0.05  0.03
19:10:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   63    0.01  0.05  0.61  0.02  0.05  0.03
19:10:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   64    0.01  0.05  0.61  0.02  0.05  0.03
19:10:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   65    0.01  0.05  0.61  0.02  0.05  0.03
19:10:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   66    0.01  0.05  0.61  0.02  0.05  0.03
19:10:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   67    0.01  0.05  0.61  0.02  0.05  0.03
19:10:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   68    0.01  0.05  0.61  0.02  0.05  0.03
19:10:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   69    0.01  0.05  0.61  0.02  0.05  0.03
19:11:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   70    0.01  0.05  0.61  0.02  0.05  0.03
19:11:01 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   71    0.01  0.05  0.61  0.02  0.05  0.03
19:11:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   72    0.01  0.05  0.61  0.02  0.05  0.03
19:11:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   73    0.01  0.05  0.61  0.02  0.05  0.03
19:11:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   74    0.01  0.05  0.61  0.02  0.05  0.03
19:11:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   75    0.01  0.05  0.61  0.02  0.05  0.03
19:11:06 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   76    0.01  0.05  0.61  0.02  0.05  0.03
19:11:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   77    0.01  0.05  0.61  0.02  0.05  0.03
19:11:08 D5F7.0   D5F700436288  stand   65   -        stand              trk  1.00 OpenFloor  1   78    0.01  0.05  0.61  0.02  0.05  0.03
19:11:09 D5F7.0   D5F700436288  stand   63   -        stand              trk  1.00 OpenFloor  1   79    0.01  0.05  0.61  0.02  0.05  0.03
19:11:10 D5F7.0   D5F700436288  stand   73   -        stand              trk  1.00 OpenFloor  1   80    0.01  0.05  0.61  0.02  0.05  0.03
19:11:11 D5F7.0   D5F700436288  walk    88   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:12 D5F7.0   D5F700436288  walk    86   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:13 D5F7.0   D5F700436288  walk    93   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:14 D5F7.0   D5F700436288  walk    66   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:15 D5F7.0   D5F700436288  walk    93   -        walk               trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:19 D5F7.0   D5F700436288  stand   112  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:20 D5F7.0   D5F700436288  stand   66   -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:30 D5F7.0   D5F700436288  stand   104  -        stand              trk  1.00 OpenFloor  1   0     0.00  0.05  0.61  0.02  0.05  0.03
19:11:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:44 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:49 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.61  0.02  0.05  0.03
19:11:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.61  0.02  0.05  0.03
19:11:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   30    0.01  0.05  0.61  0.02  0.05  0.03
19:11:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.61  0.02  0.05  0.03
19:11:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:58 D5F7.E   -             -       0    -        np=2               room -    OpenFloor  1   0     0.01  0.05  0.61  0.02  0.05  0.03
19:11:58 D5F7.1   D5F711158099  stand   78   -        stand              trk  0.58 OpenFloor  2   0     0.00  0.02  0.26  0.00  0.69  0.03
19:11:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  2   0     0.01  0.03  0.76  0.01  0.03  0.02
19:11:58 D5F7.1   D5F711158099  stand   77   -        stand              trk  0.42 OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
19:11:58 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.82  0.01  0.02  0.02
19:11:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:11:59 D5F7.1   D5F711158099  stand   105  -        stand              trk  0.17 OpenFloor  1   0     0.00  0.02  0.70  0.00  0.18  0.02
19:12:00 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.79  0.00  0.07  0.02
19:12:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:12:01 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.06 OpenFloor  1   0     0.00  0.02  0.83  0.00  0.03  0.02
19:12:01 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:02 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.02  0.02
19:12:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:04 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:05 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:06 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:06 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:08 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:12:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:12:09 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:12:10 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:12:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:12:11 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:12:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:12:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:12:12 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:12:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:12:13 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:12:14 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:12:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:12:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:12:15 D5F7.1   D5F711158099  stand   97   -        stand              trk  0.08 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:12:16 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:12:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:12:17 D5F7.1   D5F711158099  stand   72   -        stand              trk  0.08 OpenFloor  1   23    0.00  0.02  0.85  0.00  0.01  0.02
19:12:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   23    0.00  0.02  0.85  0.00  0.01  0.02
19:12:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   24    0.00  0.02  0.85  0.00  0.01  0.02
19:12:18 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   24    0.00  0.02  0.85  0.00  0.01  0.02
19:12:19 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   25    0.00  0.02  0.85  0.00  0.01  0.02
19:12:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   25    0.00  0.02  0.85  0.00  0.01  0.02
19:12:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   26    0.00  0.02  0.85  0.00  0.01  0.02
19:12:20 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   26    0.00  0.02  0.85  0.00  0.01  0.02
19:12:21 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   27    0.00  0.02  0.85  0.00  0.01  0.02
19:12:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   27    0.00  0.02  0.85  0.00  0.01  0.02
19:12:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
19:12:22 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
19:12:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
19:12:23 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
19:12:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
19:12:24 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
19:12:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
19:12:25 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
19:12:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
19:12:26 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
19:12:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
19:12:27 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
19:12:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   34    0.00  0.02  0.85  0.00  0.01  0.02
19:12:28 D5F7.1   D5F711158099  stand   104  -        stand              trk  0.08 OpenFloor  1   34    0.00  0.02  0.85  0.00  0.01  0.02
19:12:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
19:12:29 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
19:12:30 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
19:12:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
19:12:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
19:12:31 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
19:12:32 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   38    0.00  0.02  0.85  0.00  0.01  0.02
19:12:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   38    0.00  0.02  0.85  0.00  0.01  0.02
19:12:33 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   39    0.00  0.02  0.85  0.00  0.01  0.02
19:12:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   39    0.00  0.02  0.85  0.00  0.01  0.02
19:12:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   40    0.00  0.02  0.85  0.00  0.01  0.02
19:12:34 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   40    0.00  0.02  0.85  0.00  0.01  0.02
19:12:35 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   41    0.00  0.02  0.85  0.00  0.01  0.02
19:12:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   41    0.00  0.02  0.85  0.00  0.01  0.02
19:12:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   42    0.00  0.02  0.85  0.00  0.01  0.02
19:12:36 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   42    0.00  0.02  0.85  0.00  0.01  0.02
19:12:37 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   43    0.00  0.02  0.85  0.00  0.01  0.02
19:12:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   43    0.00  0.02  0.85  0.00  0.01  0.02
19:12:38 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   44    0.00  0.02  0.85  0.00  0.01  0.02
19:12:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   44    0.00  0.02  0.85  0.00  0.01  0.02
19:12:39 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   45    0.00  0.02  0.85  0.00  0.01  0.02
19:12:39 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   45    0.00  0.02  0.85  0.00  0.01  0.02
19:12:40 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   46    0.00  0.02  0.85  0.00  0.01  0.02
19:12:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   46    0.00  0.02  0.85  0.00  0.01  0.02
19:12:41 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   47    0.00  0.02  0.85  0.00  0.01  0.02
19:12:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   47    0.00  0.02  0.85  0.00  0.01  0.02
19:12:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   48    0.00  0.02  0.85  0.00  0.01  0.02
19:12:42 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   48    0.00  0.02  0.85  0.00  0.01  0.02
19:12:43 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   49    0.00  0.02  0.85  0.00  0.01  0.02
19:12:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   49    0.00  0.02  0.85  0.00  0.01  0.02
19:12:44 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   50    0.00  0.02  0.85  0.00  0.01  0.02
19:12:44 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   50    0.00  0.02  0.85  0.00  0.01  0.02
19:12:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   51    0.00  0.02  0.85  0.00  0.01  0.02
19:12:45 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   51    0.00  0.02  0.85  0.00  0.01  0.02
19:12:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   52    0.00  0.02  0.85  0.00  0.01  0.02
19:12:46 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   52    0.00  0.02  0.85  0.00  0.01  0.02
19:12:47 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   53    0.00  0.02  0.85  0.00  0.01  0.02
19:12:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   53    0.00  0.02  0.85  0.00  0.01  0.02
19:12:48 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   54    0.00  0.02  0.85  0.00  0.01  0.02
19:12:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   54    0.00  0.02  0.85  0.00  0.01  0.02
19:12:49 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   55    0.00  0.02  0.85  0.00  0.01  0.02
19:12:49 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   55    0.00  0.02  0.85  0.00  0.01  0.02
19:12:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   56    0.00  0.03  0.81  0.00  0.02  0.02
19:12:50 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   56    0.00  0.03  0.81  0.00  0.02  0.02
19:12:51 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   57    0.00  0.02  0.84  0.00  0.01  0.02
19:12:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   57    0.00  0.02  0.84  0.00  0.01  0.02
19:12:52 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   58    0.00  0.02  0.84  0.00  0.01  0.02
19:12:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   58    0.00  0.02  0.84  0.00  0.01  0.02
19:12:53 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   59    0.00  0.02  0.85  0.00  0.01  0.02
19:12:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   59    0.00  0.02  0.85  0.00  0.01  0.02
19:12:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   60    0.00  0.02  0.85  0.00  0.01  0.02
19:12:54 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   60    0.00  0.02  0.85  0.00  0.01  0.02
19:12:55 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   61    0.00  0.02  0.85  0.00  0.01  0.02
19:12:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   61    0.00  0.02  0.85  0.00  0.01  0.02
19:12:56 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   62    0.00  0.02  0.85  0.00  0.01  0.02
19:12:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   62    0.00  0.02  0.85  0.00  0.01  0.02
19:12:57 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   63    0.00  0.02  0.85  0.00  0.01  0.02
19:12:57 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   63    0.00  0.02  0.85  0.00  0.01  0.02
19:12:58 D5F7.E   -             -       0    -        np=1               room -    OpenFloor  1   63    0.00  0.02  0.85  0.00  0.01  0.02
19:12:58 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   64    0.00  0.04  0.74  0.00  0.02  0.04
19:12:59 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  0   61    0.01  0.05  0.68  0.01  0.03  0.04
19:13:00 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  0   62    0.01  0.05  0.65  0.01  0.04  0.03
19:13:01 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  0   63    0.01  0.05  0.63  0.01  0.04  0.03
19:13:02 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  0   64    0.01  0.05  0.62  0.02  0.05  0.03
19:13:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   65    0.01  0.05  0.62  0.02  0.05  0.03
19:13:04 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   66    0.01  0.05  0.62  0.02  0.05  0.03
19:13:05 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   67    0.01  0.05  0.61  0.02  0.05  0.03
19:13:06 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   68    0.01  0.05  0.61  0.02  0.05  0.03
19:13:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   69    0.01  0.05  0.61  0.02  0.05  0.03
19:13:08 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   70    0.01  0.05  0.61  0.02  0.05  0.03
19:13:09 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   71    0.01  0.05  0.61  0.02  0.05  0.03
19:13:10 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   72    0.01  0.05  0.61  0.02  0.05  0.03
19:13:11 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   73    0.01  0.05  0.61  0.02  0.05  0.03
19:13:12 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   74    0.01  0.05  0.61  0.02  0.05  0.03
19:13:13 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   75    0.01  0.05  0.61  0.02  0.05  0.03
19:13:14 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   76    0.01  0.05  0.61  0.02  0.05  0.03
19:13:15 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   77    0.01  0.05  0.61  0.02  0.05  0.03
19:13:16 D5F7.1   D5F711158099  stand   93   -        stand              trk  0.08 BlindOpen  0   78    0.00  0.05  0.61  0.02  0.05  0.03
19:13:17 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 BlindOpen  0   79    0.01  0.05  0.61  0.02  0.05  0.03
19:13:18 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Empty      0   80    0.01  0.05  0.61  0.02  0.05  0.03
19:13:19 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Empty      0   81    0.01  0.05  0.61  0.02  0.05  0.03
19:13:20 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Empty      0   82    0.01  0.05  0.61  0.02  0.05  0.03
19:13:21 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Empty      0   83    0.01  0.05  0.61  0.02  0.05  0.03
19:13:22 D5F7.1   D5F711158099  walk    118  -        walk               trk  0.08 Empty      0   0     0.00  0.05  0.61  0.02  0.05  0.03
19:13:23 D5F7.1   D5F711158099  walk    103  -        walk               trk  0.08 Empty      0   0     0.00  0.05  0.61  0.02  0.05  0.03
19:13:24 D5F7.1   D5F711158099  walk    85   -        walk               trk  0.08 Empty      0   0     0.00  0.05  0.61  0.02  0.05  0.03
19:13:25 D5F7.1   D5F711158099  walk    0    -        walk               trk  0.08 Empty      0   0     0.00  0.05  0.61  0.02  0.05  0.03
19:13:26 D5F7.1   D5F711158099  walk    0    -        walk               trk  0.08 Empty      0   0     0.00  0.05  0.61  0.02  0.05  0.03
19:13:27 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.07  0.41  0.03  0.07  0.04
19:13:28 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.07  0.28  0.03  0.09  0.03
19:13:29 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.06  0.20  0.04  0.09  0.03
19:13:30 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.05  0.15  0.04  0.08  0.02
19:13:31 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.04  0.12  0.04  0.07  0.02
19:13:32 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.03  0.11  0.04  0.05  0.02
19:13:33 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.03  0.10  0.04  0.04  0.02
19:13:34 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.10  0.04  0.04  0.02
19:13:35 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.04  0.03  0.02
19:13:36 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.03  0.02
19:13:37 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.03  0.02
19:13:38 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:39 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:40 D5F7.1   D5F711158099  sit     74   -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:41 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:42 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:43 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:44 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:45 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:46 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 Empty      0   0     0.01  0.02  0.09  0.03  0.02  0.02
19:13:47 D5F7.E   -             -       0    -        np=2               room -    Empty      0   0     0.12  0.10  0.15  0.13  0.19  0.02
19:13:47 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.05  0.02  0.01  0.01
19:13:47 D5F7.0   D5F700436288  stand   71   -        stand              trk  1.00 OpenFloor  1   0     0.11  0.09  0.25  0.11  0.16  0.02
19:13:48 D5F7.0   D5F700436288  stand   65   -        stand              trk  1.00 OpenFloor  1   0     0.07  0.06  0.52  0.07  0.10  0.01
19:13:48 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.01  0.01
19:13:49 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:49 D5F7.0   D5F700436288  stand   76   -        stand              trk  1.00 OpenFloor  1   0     0.03  0.04  0.71  0.03  0.05  0.02
19:13:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.02  0.03  0.80  0.01  0.02  0.02
19:13:50 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:51 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.01  0.01
19:13:51 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.05  0.66  0.01  0.04  0.03
19:13:52 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.01  0.03  0.78  0.01  0.03  0.02
19:13:52 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:53 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.83  0.01  0.02  0.02
19:13:53 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:54 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:54 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:13:55 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:13:56 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:13:56 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:57 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:57 D5F7.0   D5F700436288  stand   80   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:13:58 D5F7.0   D5F700436288  stand   63   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:13:58 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:59 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:13:59 D5F7.0   D5F700436288  stand   50   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:00 D5F7.0   D5F700436288  stand   47   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:00 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:14:01 D5F7.1   D5F711158099  sit     0    -        sit                trk  0.08 OpenFloor  1   13    0.00  0.01  0.04  0.01  0.00  0.01
19:14:01 D5F7.0   D5F700436288  stand   55   -        stand              trk  1.00 OpenFloor  1   13    0.00  0.02  0.85  0.00  0.01  0.02
19:14:02 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Sit        1   14    0.00  0.01  0.35  0.03  0.01  0.02
19:14:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 Sit        1   14    0.00  0.02  0.85  0.00  0.01  0.02
19:14:02 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   15    0.00  0.02  0.63  0.02  0.01  0.02
19:14:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:14:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   16    0.00  0.02  0.77  0.01  0.01  0.02
19:14:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:14:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:14:04 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   17    0.00  0.02  0.82  0.01  0.01  0.02
19:14:05 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   18    0.00  0.02  0.84  0.00  0.01  0.02
19:14:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:14:06 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   18    0.00  0.02  0.84  0.00  0.01  0.02
19:14:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:08 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:09 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:09 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:10 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:11 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:12 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:13 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:14 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:14 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:15 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:15 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:16 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:17 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:18 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:19 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:20 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:21 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   14    0.00  0.02  0.85  0.00  0.01  0.02
19:14:21 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   14    0.00  0.02  0.85  0.00  0.01  0.02
19:14:22 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:14:22 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:14:23 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:14:23 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:14:24 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:14:24 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:14:25 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:14:25 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:14:26 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:14:26 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:14:27 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:14:27 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:14:28 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:14:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:14:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:14:29 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:14:30 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:30 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:31 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:31 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:32 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:32 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:33 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:33 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:34 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:34 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:35 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:35 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:36 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:36 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:37 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:37 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:38 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:38 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:39 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:39 D5F7.0   D5F700436288  stand   92   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:40 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:40 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:41 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:41 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:42 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:42 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:43 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:43 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:44 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:44 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:45 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:45 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:46 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:46 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:47 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:47 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:48 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:48 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:49 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
19:14:49 D5F7.0   D5F700436288  stand   83   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.03  0.81  0.00  0.02  0.02
19:14:50 D5F7.1   D5F711158099  stand   82   -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:14:50 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:14:51 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:14:51 D5F7.0   D5F700436288  stand   64   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.01  0.02
19:14:52 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:52 D5F7.0   D5F700436288  stand   76   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:53 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:53 D5F7.0   D5F700436288  stand   71   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:14:54 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   13    0.00  0.02  0.85  0.00  0.01  0.02
19:14:54 D5F7.0   D5F700436288  stand   56   -        stand              trk  1.00 OpenFloor  1   13    0.00  0.02  0.85  0.00  0.01  0.02
19:14:55 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   14    0.00  0.02  0.85  0.00  0.01  0.02
19:14:55 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   14    0.00  0.02  0.85  0.00  0.01  0.02
19:14:56 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:14:56 D5F7.0   D5F700436288  stand   63   -        stand              trk  1.00 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:14:57 D5F7.0   D5F700436288  stand   75   -        stand              trk  1.00 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:14:57 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:14:58 D5F7.0   D5F700436288  stand   80   -        stand              trk  1.00 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:14:58 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:14:59 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:14:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:15:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:15:00 D5F7.1   D5F711158099  stand   114  -        stand              trk  0.08 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:15:01 D5F7.0   D5F700436288  stand   45   -        stand              trk  1.00 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:15:01 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:15:02 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:15:02 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:15:03 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:15:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:15:04 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   23    0.00  0.02  0.85  0.00  0.01  0.02
19:15:04 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   23    0.00  0.02  0.85  0.00  0.01  0.02
19:15:05 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   24    0.00  0.02  0.85  0.00  0.01  0.02
19:15:05 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   24    0.00  0.02  0.85  0.00  0.01  0.02
19:15:06 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   25    0.00  0.02  0.85  0.00  0.01  0.02
19:15:06 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   25    0.00  0.02  0.85  0.00  0.01  0.02
19:15:07 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   26    0.00  0.02  0.85  0.00  0.01  0.02
19:15:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   26    0.00  0.02  0.85  0.00  0.01  0.02
19:15:08 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   27    0.00  0.02  0.85  0.00  0.01  0.02
19:15:08 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   27    0.00  0.02  0.85  0.00  0.01  0.02
19:15:09 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
19:15:09 D5F7.0   D5F700436288  stand   94   -        stand              trk  1.00 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
19:15:10 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
19:15:10 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
19:15:11 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
19:15:11 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
19:15:12 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
19:15:12 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
19:15:13 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
19:15:13 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
19:15:14 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
19:15:14 D5F7.0   D5F700436288  stand   100  -        stand              trk  1.00 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
19:15:15 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   34    0.00  0.02  0.85  0.00  0.01  0.02
19:15:15 D5F7.0   D5F700436288  stand   98   -        stand              trk  1.00 OpenFloor  1   34    0.00  0.02  0.85  0.00  0.01  0.02
19:15:16 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
19:15:16 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
19:15:17 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
19:15:17 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
19:15:18 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
19:15:18 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
19:15:19 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   38    0.00  0.02  0.85  0.00  0.01  0.02
19:15:19 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   38    0.00  0.02  0.85  0.00  0.01  0.02
19:15:20 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   39    0.00  0.02  0.85  0.00  0.01  0.02
19:15:20 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   39    0.00  0.02  0.85  0.00  0.01  0.02
19:15:21 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   40    0.00  0.02  0.85  0.00  0.01  0.02
19:15:21 D5F7.0   D5F700436288  stand   76   -        stand              trk  1.00 OpenFloor  1   40    0.00  0.02  0.85  0.00  0.01  0.02
19:15:22 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   41    0.00  0.02  0.85  0.00  0.01  0.02
19:15:22 D5F7.0   D5F700436288  stand   91   -        stand              trk  1.00 OpenFloor  1   41    0.00  0.02  0.85  0.00  0.01  0.02
19:15:23 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   42    0.00  0.02  0.85  0.00  0.01  0.02
19:15:23 D5F7.0   D5F700436288  stand   82   -        stand              trk  1.00 OpenFloor  1   42    0.00  0.02  0.85  0.00  0.01  0.02
19:15:24 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   43    0.00  0.02  0.85  0.00  0.01  0.02
19:15:24 D5F7.0   D5F700436288  stand   82   -        stand              trk  1.00 OpenFloor  1   43    0.00  0.02  0.85  0.00  0.01  0.02
19:15:25 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   44    0.00  0.02  0.85  0.00  0.01  0.02
19:15:25 D5F7.0   D5F700436288  stand   50   -        stand              trk  1.00 OpenFloor  1   44    0.00  0.02  0.85  0.00  0.01  0.02
19:15:26 D5F7.0   D5F700436288  stand   64   -        stand              trk  1.00 OpenFloor  1   45    0.00  0.02  0.85  0.00  0.01  0.02
19:15:26 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   45    0.00  0.02  0.85  0.00  0.01  0.02
19:15:27 D5F7.0   D5F700436288  stand   36   -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:27 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:28 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:28 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:29 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:29 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:30 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 OpenFloor  1   0     0.01  0.05  0.48  0.01  0.02  0.05
19:15:30 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:31 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Sit        1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:31 D5F7.0   D5F700436288  sit     57   -        sit                trk  1.00 Sit        1   0     0.01  0.04  0.22  0.01  0.04  0.03
19:15:32 D5F7.0   D5F700436288  sit     24   -        sit                trk  1.00 Sit        1   0     0.00  0.02  0.09  0.01  0.02  0.01
19:15:32 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 Sit        1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:33 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 OpenFloor  1   0     0.00  0.01  0.05  0.01  0.01  0.01
19:15:33 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:34 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:34 D5F7.0   D5F700436288  sit     51   -        sit                trk  1.00 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.01  0.01
19:15:35 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:35 D5F7.0   D5F700436288  sit     50   -        sit                trk  1.00 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:15:36 D5F7.0   D5F700436288  sit     59   -        sit                trk  1.00 OpenFloor  1   0     0.00  0.01  0.04  0.01  0.00  0.01
19:15:36 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:37 D5F7.0   D5F700436288  sit     60   -        sit                trk  1.00 OpenFloor  1   14    0.00  0.01  0.04  0.01  0.00  0.01
19:15:37 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   14    0.00  0.02  0.85  0.00  0.01  0.02
19:15:38 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   15    0.00  0.02  0.85  0.00  0.01  0.02
19:15:38 D5F7.0   D5F700436288  sit     73   -        sit                trk  1.00 OpenFloor  1   15    0.00  0.01  0.04  0.01  0.00  0.01
19:15:39 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   16    0.00  0.02  0.85  0.00  0.01  0.02
19:15:39 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 OpenFloor  1   16    0.00  0.01  0.04  0.01  0.00  0.01
19:15:40 D5F7.0   D5F700436288  sit     60   -        sit                trk  1.00 OpenFloor  1   17    0.00  0.01  0.04  0.01  0.00  0.01
19:15:40 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   17    0.00  0.02  0.85  0.00  0.01  0.02
19:15:41 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   18    0.00  0.02  0.85  0.00  0.01  0.02
19:15:41 D5F7.0   D5F700436288  sit     47   -        sit                trk  1.00 OpenFloor  1   18    0.00  0.01  0.04  0.01  0.00  0.01
19:15:42 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   19    0.00  0.02  0.85  0.00  0.01  0.02
19:15:42 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 OpenFloor  1   19    0.00  0.01  0.04  0.01  0.00  0.01
19:15:43 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   20    0.00  0.02  0.85  0.00  0.01  0.02
19:15:43 D5F7.0   D5F700436288  sit     105  -        sit                trk  1.00 OpenFloor  1   20    0.00  0.00  0.07  0.01  0.00  0.01
19:15:44 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   21    0.00  0.02  0.85  0.00  0.01  0.02
19:15:44 D5F7.0   D5F700436288  sit     78   -        sit                trk  1.00 OpenFloor  1   21    0.00  0.01  0.08  0.01  0.00  0.01
19:15:45 D5F7.0   D5F700436288  sit     79   -        sit                trk  1.00 OpenFloor  1   22    0.00  0.01  0.05  0.01  0.00  0.01
19:15:45 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   22    0.00  0.02  0.85  0.00  0.01  0.02
19:15:46 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   23    0.00  0.02  0.85  0.00  0.01  0.02
19:15:46 D5F7.0   D5F700436288  sit     0    -        sit                trk  1.00 OpenFloor  1   23    0.00  0.01  0.04  0.01  0.00  0.01
19:15:47 D5F7.0   D5F700436288  sit     87   -        sit                trk  1.00 OpenFloor  1   24    0.00  0.00  0.13  0.01  0.00  0.01
19:15:47 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   24    0.00  0.02  0.85  0.00  0.01  0.02
19:15:48 D5F7.0   D5F700436288  stand   90   -        stand              trk  1.00 OpenFloor  1   25    0.00  0.01  0.46  0.02  0.01  0.02
19:15:48 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   25    0.00  0.02  0.85  0.00  0.01  0.02
19:15:49 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   26    0.00  0.03  0.81  0.00  0.02  0.02
19:15:49 D5F7.0   D5F700436288  stand   68   -        stand              trk  1.00 OpenFloor  1   26    0.00  0.03  0.72  0.02  0.02  0.02
19:15:50 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   27    0.00  0.02  0.84  0.00  0.01  0.02
19:15:50 D5F7.0   D5F700436288  stand   104  -        stand              trk  1.00 OpenFloor  1   27    0.00  0.02  0.81  0.01  0.01  0.02
19:15:51 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   28    0.00  0.02  0.84  0.00  0.01  0.02
19:15:51 D5F7.0   D5F700436288  stand   65   -        stand              trk  1.00 OpenFloor  1   28    0.00  0.02  0.84  0.00  0.01  0.02
19:15:52 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
19:15:52 D5F7.0   D5F700436288  stand   91   -        stand              trk  1.00 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
19:15:53 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
19:15:53 D5F7.0   D5F700436288  stand   94   -        stand              trk  1.00 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
19:15:54 D5F7.0   D5F700436288  stand   51   -        stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
19:15:54 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
19:15:55 D5F7.0   D5F700436288  walk    59   -        walk               trk  1.00 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
19:15:55 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
19:15:56 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
19:15:56 D5F7.0   D5F700436288  walk    63   -        walk               trk  1.00 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
19:15:57 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:57 D5F7.0   D5F700436288  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:58 D5F7.0   D5F700436288  walk    0    -        walk               trk  1.00 OpenFloor  1   0     0.00  0.01  0.96  0.00  0.00  0.01
19:15:58 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:15:59 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
19:15:59 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:16:00 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
19:16:00 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:16:01 D5F7.0   D5F700436288  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
19:16:01 D5F7.1   D5F711158099  stand   0    -        stand              trk  0.08 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:16:02 D5F7.E0  D5F700436288  -       0    -        ExitRoom(rdr)      trk  1.00 OpenFloor  1   0     0.00  0.01  0.97  0.00  0.00  0.01
19:16:02 D5F7.0   -             stand   0    -        stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:16:02 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:16:03 D5F7.E   -             -       0    -        np=1               room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
19:16:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   0     0.00  0.04  0.74  0.00  0.02  0.04
19:16:04 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   13    0.01  0.05  0.68  0.01  0.03  0.04
19:16:05 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   14    0.01  0.05  0.65  0.01  0.04  0.03
19:16:06 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   15    0.01  0.05  0.63  0.01  0.04  0.03
19:16:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   16    0.01  0.05  0.62  0.02  0.05  0.03
19:16:07 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   17    0.01  0.05  0.62  0.02  0.05  0.03
19:16:08 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   18    0.01  0.05  0.62  0.02  0.05  0.03
19:16:09 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  1   18    0.01  0.05  0.62  0.02  0.05  0.03
19:16:10 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   19    0.01  0.05  0.61  0.02  0.05  0.03
19:16:11 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   20    0.01  0.05  0.61  0.02  0.05  0.03
19:16:11 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   21    0.01  0.05  0.61  0.02  0.05  0.03
19:16:12 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   22    0.01  0.05  0.61  0.02  0.05  0.03
19:16:13 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   23    0.01  0.05  0.61  0.02  0.05  0.03
19:16:14 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   24    0.01  0.05  0.61  0.02  0.05  0.03
19:16:15 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   25    0.01  0.05  0.61  0.02  0.05  0.03
19:16:16 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   26    0.01  0.05  0.61  0.02  0.05  0.03
19:16:17 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   27    0.01  0.05  0.61  0.02  0.05  0.04
19:16:18 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   28    0.01  0.05  0.60  0.02  0.06  0.05
19:16:19 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   29    0.01  0.05  0.58  0.02  0.06  0.06
19:16:20 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   30    0.01  0.05  0.56  0.02  0.08  0.08
19:16:21 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   31    0.01  0.05  0.54  0.02  0.09  0.10
19:16:22 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   32    0.01  0.04  0.51  0.02  0.11  0.13
19:16:23 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   33    0.01  0.04  0.47  0.01  0.14  0.17
19:16:24 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   34    0.01  0.04  0.42  0.01  0.16  0.22
19:16:25 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   35    0.01  0.03  0.35  0.01  0.19  0.28
19:16:26 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   36    0.00  0.03  0.29  0.01  0.23  0.35
19:16:27 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   37    0.00  0.02  0.21  0.01  0.25  0.43
19:16:28 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   38    0.00  0.01  0.14  0.00  0.26  0.53
19:16:29 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   39    0.00  0.01  0.08  0.00  0.25  0.63
19:16:30 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   40    0.00  0.00  0.04  0.00  0.22  0.72
19:16:31 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   41    0.00  0.00  0.02  0.00  0.19  0.79
19:16:32 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   42    0.00  0.00  0.01  0.00  0.17  0.81
19:16:33 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   43    0.00  0.00  0.01  0.00  0.17  0.82
19:16:34 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   44    0.00  0.00  0.01  0.00  0.17  0.82
19:16:35 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   45    0.00  0.00  0.00  0.00  0.17  0.82
19:16:36 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   46    0.00  0.00  0.00  0.00  0.17  0.82
19:16:37 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   47    0.00  0.00  0.00  0.00  0.17  0.82
19:16:38 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   48    0.00  0.00  0.00  0.00  0.17  0.82
19:16:39 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   49    0.00  0.00  0.00  0.00  0.17  0.82
19:16:40 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   50    0.00  0.00  0.00  0.00  0.17  0.82
19:16:41 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   51    0.00  0.00  0.00  0.00  0.17  0.82
19:16:42 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   52    0.00  0.00  0.00  0.00  0.17  0.82
19:16:43 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   53    0.00  0.00  0.00  0.00  0.17  0.82
19:16:44 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   54    0.00  0.00  0.00  0.00  0.17  0.82
19:16:45 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   55    0.00  0.00  0.00  0.00  0.17  0.82
19:16:46 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Left       1   56    0.00  0.00  0.00  0.00  0.17  0.82
19:16:47 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Empty      1   57    0.00  0.00  0.02  0.00  0.85  0.12
19:16:48 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Empty      1   58    0.00  0.02  0.21  0.00  0.70  0.01
19:16:49 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Empty      1   59    0.00  0.03  0.31  0.01  0.55  0.01
19:16:50 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 Empty      1   60    0.00  0.04  0.40  0.01  0.41  0.02
19:16:51 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   61    0.00  0.04  0.46  0.01  0.30  0.02
19:16:52 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   62    0.01  0.05  0.51  0.01  0.22  0.03
19:16:53 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   63    0.01  0.05  0.55  0.02  0.17  0.03
19:16:54 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   64    0.01  0.05  0.57  0.02  0.13  0.03
19:16:55 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   65    0.01  0.05  0.58  0.02  0.10  0.03
19:16:56 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   66    0.01  0.05  0.59  0.02  0.08  0.03
19:16:57 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   67    0.01  0.05  0.60  0.02  0.07  0.03
19:16:58 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   68    0.01  0.05  0.60  0.02  0.07  0.03
19:16:59 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   69    0.01  0.05  0.61  0.02  0.06  0.03
19:17:00 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   70    0.01  0.05  0.61  0.02  0.06  0.03
19:17:01 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   71    0.01  0.05  0.61  0.02  0.06  0.03
19:17:02 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   72    0.01  0.05  0.61  0.02  0.06  0.03
19:17:03 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   73    0.01  0.05  0.61  0.02  0.05  0.03
19:17:04 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   74    0.01  0.05  0.61  0.02  0.05  0.03
19:17:05 D5F7.1   D5F711158099  stand   0    -        stand              trk  1.00 OpenFloor  1   75    0.01  0.05  0.61  0.02  0.05  0.03
19:17:06 D5F7.E   -             -       0    -        np=0  ★0           room -    OpenFloor  1   75    0.01  0.05  0.61  0.02  0.05  0.03
19:17:06 D5F7.88  -             88      -    -        no-target(88)      room -    OpenFloor  1   76    0.01  0.05  0.61  0.02  0.05  0.03
19:17:07 D5F7.88  -             88      -    -        no-target(88)      room -    OpenFloor  0   0     0.01  0.08  0.44  0.03  0.08  0.05
19:17:08 D5F7.88  -             88      -    -        no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:09 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:10 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:11 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:12 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:13 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:14 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:15 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:16 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:17 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:18 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:19 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:20 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:21 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:22 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:23 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:24 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:25 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:26 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:27 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:28 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:29 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:30 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:31 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:32 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:33 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.34  0.04  0.11  0.04
19:17:34 D5F7.88  -             88      -    -        no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:35 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:36 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:37 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:38 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:39 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:40 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:41 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:42 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:43 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:44 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:45 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:46 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:47 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:48 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:49 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:50 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:51 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:52 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:53 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:54 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:55 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:56 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:57 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:58 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:17:59 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:00 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:01 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:02 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:03 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:04 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:05 -.-      -             -       -    -        (no frame, held)   room -    OpenFloor  0   0     0.02  0.09  0.27  0.05  0.14  0.03
19:18:06 D5F7.88  -             88      -    -        no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:07 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:08 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:09 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:10 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:11 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:12 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:13 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:14 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:15 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:16 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:17 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:18 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:19 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:20 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:21 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:22 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:23 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:24 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:25 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:26 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:27 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:28 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:29 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:30 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:31 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:32 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:33 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:34 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:35 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:36 -.-      -             -       -    -        (no frame, held)   room -    BlindOpen  0   0     0.03  0.10  0.21  0.07  0.16  0.03
19:18:37 D5F7.88  -             88      -    -        no-target(88)      room -    BlindOpen  0   0     0.03  0.10  0.20  0.08  0.17  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
19:02:13.948 D5F7.88   88     -    -      -      -     -    -    
19:02:45.410 D5F7.88   88     -    -      -      -     -    -    
19:03:17.298 D5F7.88   88     -    -      -      -     -    -    
19:03:48.902 D5F7.88   88     -    -      -      -     -    -    
19:04:20.781 D5F7.88   88     -    -      -      -     -    -    
19:04:36.288 D5F7.0    stand  0    -60    60     83    80        
19:04:36.575 D5F7.0    stand  0    -40    40     95    80   28   
19:04:37.585 D5F7.0    stand  0    -40    40     78    80   0    
19:04:38.604 D5F7.0    stand  0    -30    30     86    80   14   
19:04:39.505 D5F7.0    stand  0    0      20     61    80   31   
19:04:40.510 D5F7.0    walk   0    30     30     89    80   31   
19:04:41.521 D5F7.0    walk   0    60     20     62    80   31   
19:04:42.536 D5F7.0    walk   0    80     0      68    80   28   
19:04:43.548 D5F7.0    walk   0    70     0      85    80   10   
19:04:44.470 D5F7.0    walk   0    70     -10    58    80   10   
19:04:45.471 D5F7.0    walk   0    70     -10    63    80   0    
19:04:46.479 D5F7.0    walk   0    90     -10    72    80   20   
19:04:47.490 D5F7.0    walk   0    110    -10    61    80   20   
19:04:48.510 D5F7.0    walk   0    80     -10    63    80   30   
19:04:49.411 D5F7.0    walk   0    90     -20    63    80   14   
19:04:50.432 D5F7.0    walk   0    110    -30    73    80   22   
19:04:51.440 D5F7.0    sit    0    100    -10    82    80   22   
19:04:52.457 D5F7.0    sit    0    90     0      47    80   14   
19:04:53.471 D5F7.0    sit    0    130    -10    0     80   41   
19:04:54.383 D5F7.0    sit    0    110    0      71    80   22   
19:04:55.464 D5F7.0    sit    0    30     -10    0     80   80   
19:04:56.417 D5F7.0    sit    0    30     -10    0     80   0    
19:04:57.354 D5F7.0    sit    0    30     0      0     80   10   
19:04:58.363 D5F7.0    sit    0    20     -10    0     80   14   
19:04:59.373 D5F7.0    sit    0    30     -10    83    80   10   
19:05:00.388 D5F7.0    sit    0    40     -20    57    80   14   
19:05:01.402 D5F7.0    sit    0    40     -20    75    80   0    
19:05:02.306 D5F7.0    sit    0    0      -10    0     80   41   
19:05:03.325 D5F7.0    sit    0    0      -10    0     80   0    
19:05:04.334 D5F7.0    sit    0    10     0      78    80   14   
19:05:05.352 D5F7.0    sit    0    20     -10    69    80   14   
19:05:06.363 D5F7.0    sit    0    80     0      0     80   60   
19:05:07.265 D5F7.0    sit    0    30     -10    0     80   50   
19:05:08.282 D5F7.0    sit    0    30     -10    0     80   0    
19:05:09.293 D5F7.0    sit    0    40     -10    0     80   10   
19:05:10.314 D5F7.0    sit    0    40     -10    0     80   0    
19:05:11.324 D5F7.0    sit    0    40     -10    0     80   0    
19:05:12.225 D5F7.0    sit    0    40     -10    0     80   0    
19:05:13.238 D5F7.0    sit    0    40     -20    0     80   10   
19:05:14.281 D5F7.0    sit    0    40     -20    0     80   0    
19:05:15.300 D5F7.0    sit    0    40     -20    0     80   0    
19:05:16.193 D5F7.0    sit    0    40     -20    0     80   0    
19:05:17.208 D5F7.0    sit    0    40     -20    0     80   0    
19:05:18.234 D5F7.0    sit    0    40     -20    0     80   0    
19:05:19.249 D5F7.0    sit    0    40     -20    0     80   0    
19:05:20.247 D5F7.0    sit    0    40     -20    0     80   0    
19:05:21.167 D5F7.0    sit    0    40     -20    0     80   0    
19:05:22.173 D5F7.0    sit    0    40     -20    0     80   0    
19:05:23.184 D5F7.0    sit    0    40     -20    0     80   0    
19:05:24.199 D5F7.0    sit    0    40     -20    0     80   0    
19:05:25.217 D5F7.0    sit    0    40     -20    0     80   0    
19:05:26.119 D5F7.0    sit    0    40     -20    0     80   0    
19:05:27.143 D5F7.0    sit    0    40     -20    0     80   0    
19:05:28.148 D5F7.0    sit    0    40     -20    0     80   0    
19:05:29.157 D5F7.0    sit    0    30     -10    0     80   14   
19:05:30.174 D5F7.0    sit    0    20     -10    0     80   10   
19:05:31.083 D5F7.0    sit    0    20     -10    0     80   0    
19:05:32.091 D5F7.0    sit    0    20     -10    0     80   0    
19:05:33.116 D5F7.0    sit    0    20     -10    0     80   0    
19:05:34.128 D5F7.0    sit    0    20     -10    0     80   0    
19:05:35.127 D5F7.0    sit    0    30     0      0     80   14   
19:05:36.040 D5F7.0    sit    0    30     0      0     80   0    
19:05:37.074 D5F7.0    sit    0    30     0      0     80   0    
19:05:38.062 D5F7.0    sit    0    30     0      0     80   0    
19:05:39.077 D5F7.0    sit    0    40     -20    60    80   22   
19:05:40.093 D5F7.0    sit    0    40     -30    0     80   10   
19:05:40.994 D5F7.0    sit    0    40     0      0     80   30   
19:05:42.022 D5F7.0    sit    0    120    -10    73    80   80   
19:05:43.026 D5F7.0    sit    0    90     -10    69    80   30   
19:05:44.040 D5F7.0    walk   0    100    -10    89    80   10   
19:05:45.046 D5F7.0    walk   0    40     0      0     80   60   
19:05:45.960 D5F7.0    stand  0    40     0      0     80   0    
19:05:46.976 D5F7.0    stand  0    40     0      0     80   0    
19:05:47.980 D5F7.0    stand  0    40     0      0     80   0    
19:05:48.997 D5F7.0    stand  0    40     0      0     80   0    
19:05:50.015 D5F7.0    stand  0    40     0      0     80   0    
19:05:50.920 D5F7.0    stand  0    40     0      0     80   0    
19:05:51.939 D5F7.0    stand  0    40     0      38    80   0    
19:05:52.945 D5F7.0    stand  0    40     0      0     80   0    
19:05:54.044 D5F7.0    stand  0    40     0      0     80   0    
19:05:54.889 D5F7.0    stand  0    40     0      0     80   0    
19:05:55.897 D5F7.0    stand  0    40     0      0     80   0    
19:05:56.908 D5F7.0    stand  0    40     0      0     80   0    
19:05:57.922 D5F7.0    stand  0    50     0      0     80   10   
19:05:58.934 D5F7.0    stand  0    50     0      0     80   0    
19:05:59.841 D5F7.0    stand  0    40     0      0     80   10   
19:06:00.862 D5F7.0    stand  0    50     0      0     80   10   
19:06:01.915 D5F7.0    stand  0    90     -10    90    80   41   
19:06:02.821 D5F7.0    stand  0    160    -10    0     80   70   
19:06:03.837 D5F7.0    stand  0    160    -10    0     80   0    
19:06:04.845 D5F7.0    stand  0    120    -20    0     80   41   
19:06:05.865 D5F7.0    walk   0    70     -10    94    80   50   
19:06:06.873 D5F7.0    walk   0    40     0      0     80   31   
19:06:07.785 D5F7.0    walk   0    40     0      0     80   0    
19:06:08.794 D5F7.0    stand  0    40     0      0     80   0    
19:06:09.806 D5F7.0    stand  0    40     -10    0     80   10   
19:06:10.818 D5F7.0    stand  0    40     -10    0     80   0    
19:06:11.830 D5F7.0    stand  0    40     0      90    80   10   
19:06:12.746 D5F7.0    stand  0    110    -10    91    80   70   
19:06:13.752 D5F7.0    stand  0    40     -10    0     80   70   
19:06:14.766 D5F7.0    stand  0    40     -10    0     80   0    
19:06:15.790 D5F7.0    stand  0    40     -10    0     80   0    
19:06:16.821 D5F7.0    stand  0    40     -10    0     80   0    
19:06:17.699 D5F7.0    stand  0    40     -10    0     80   0    
19:06:18.720 D5F7.0    stand  0    40     0      0     80   10   
19:06:19.724 D5F7.0    stand  0    40     0      0     80   0    
19:06:20.740 D5F7.0    stand  0    40     0      0     80   0    
19:06:21.752 D5F7.0    stand  0    40     0      0     80   0    
19:06:22.660 D5F7.0    stand  0    40     0      0     80   0    
19:06:23.682 D5F7.0    stand  0    40     0      0     80   0    
19:06:24.688 D5F7.0    stand  0    40     0      0     80   0    
19:06:25.702 D5F7.0    stand  0    40     0      0     80   0    
19:06:26.712 D5F7.0    stand  0    40     0      0     80   0    
19:06:27.629 D5F7.0    stand  0    40     0      0     80   0    
19:06:28.637 D5F7.0    stand  0    50     -10    0     80   14   
19:06:29.647 D5F7.0    stand  0    50     -10    0     80   0    
19:06:30.662 D5F7.0    stand  0    50     -10    0     80   0    
19:06:31.677 D5F7.0    stand  0    50     -10    0     80   0    
19:06:32.581 D5F7.0    stand  0    50     -10    0     80   0    
19:06:33.600 D5F7.0    stand  0    40     0      0     80   14   
19:06:34.609 D5F7.0    stand  0    40     0      0     80   0    
19:06:35.619 D5F7.0    stand  0    40     0      0     80   0    
19:06:36.635 D5F7.0    stand  0    40     0      0     80   0    
19:06:37.540 D5F7.0    stand  0    40     0      0     80   0    
19:06:38.558 D5F7.0    stand  0    40     0      0     80   0    
19:06:39.570 D5F7.0    stand  0    40     0      0     80   0    
19:06:40.582 D5F7.0    stand  0    40     0      0     80   0    
19:06:41.595 D5F7.0    stand  0    40     0      0     80   0    
19:06:42.504 D5F7.0    stand  0    40     0      0     80   0    
19:06:43.515 D5F7.0    stand  0    40     0      0     80   0    
19:06:44.562 D5F7.0    stand  0    40     0      105   80   0    
19:06:45.549 D5F7.0    stand  0    40     0      0     80   0    
19:06:46.551 D5F7.0    stand  0    40     0      89    80   0    
19:06:47.458 D5F7.0    stand  0    40     0      0     80   0    
19:06:48.474 D5F7.0    stand  0    30     0      0     80   10   
19:06:49.444 D5F7.0    stand  0    30     0      0     80   0    
19:06:50.459 D5F7.0    stand  0    40     0      0     80   10   
19:06:51.476 D5F7.0    stand  0    40     0      0     80   0    
19:06:52.486 D5F7.0    stand  0    40     0      0     80   0    
19:06:53.700 D5F7.0    stand  0    40     0      0     80   0    
19:06:54.408 D5F7.0    stand  0    40     0      0     80   0    
19:06:55.417 D5F7.0    stand  0    40     0      0     80   0    
19:06:56.445 D5F7.0    stand  0    40     0      0     80   0    
19:06:57.445 D5F7.0    stand  0    40     0      0     80   0    
19:06:58.456 D5F7.0    stand  0    40     0      0     80   0    
19:06:59.366 D5F7.0    stand  0    40     0      0     80   0    
19:07:00.376 D5F7.0    stand  0    40     0      0     80   0    
19:07:01.390 D5F7.0    stand  0    40     0      0     80   0    
19:07:02.406 D5F7.0    stand  0    40     0      0     80   0    
19:07:03.416 D5F7.0    stand  0    40     0      0     80   0    
19:07:04.328 D5F7.0    stand  0    40     0      0     80   0    
19:07:05.344 D5F7.0    stand  0    40     0      0     80   0    
19:07:06.374 D5F7.0    stand  0    40     0      94    80   0    
19:07:07.401 D5F7.0    stand  0    60     -10    0     80   22   
19:07:08.293 D5F7.0    stand  0    40     -10    0     80   20   
19:07:09.309 D5F7.0    stand  0    40     -10    0     80   0    
19:07:10.330 D5F7.0    stand  0    40     -10    0     80   0    
19:07:11.340 D5F7.0    stand  0    40     -10    0     80   0    
19:07:12.344 D5F7.0    stand  0    40     0      0     80   10   
19:07:13.252 D5F7.0    stand  0    40     0      0     80   0    
19:07:14.268 D5F7.0    stand  0    40     0      0     80   0    
19:07:15.311 D5F7.0    stand  0    40     0      0     80   0    
19:07:16.290 D5F7.0    stand  0    40     0      0     80   0    
19:07:17.302 D5F7.0    stand  0    40     0      0     80   0    
19:07:18.222 D5F7.0    stand  0    40     0      0     80   0    
19:07:19.226 D5F7.0    stand  0    40     0      0     80   0    
19:07:20.238 D5F7.0    stand  0    40     0      0     80   0    
19:07:21.249 D5F7.0    stand  0    40     0      0     80   0    
19:07:22.272 D5F7.0    stand  0    40     0      0     80   0    
19:07:23.170 D5F7.0    stand  0    40     0      0     80   0    
19:07:24.191 D5F7.0    stand  0    40     0      0     80   0    
19:07:25.198 D5F7.0    stand  0    40     0      0     80   0    
19:07:26.214 D5F7.0    stand  0    40     0      0     80   0    
19:07:27.226 D5F7.0    stand  0    40     0      0     80   0    
19:07:28.132 D5F7.0    stand  0    40     0      0     80   0    
19:07:29.146 D5F7.0    stand  0    40     0      0     80   0    
19:07:30.158 D5F7.0    stand  0    40     -10    0     80   10   
19:07:31.172 D5F7.0    stand  0    40     0      0     80   10   
19:07:31.840 D5F7.1    stand  0    -90    70     115   80   147  
19:07:31.840 D5F7.0    stand  0    40     0      0     80   147  
19:07:32.106 D5F7.0    stand  0    40     0      0     80   0    
19:07:32.106 D5F7.1    stand  0    -90    70     99    80   147  
19:07:33.113 D5F7.1    stand  0    -90    70     0     80   0    
19:07:33.113 D5F7.0    stand  0    40     0      0     80   147  
19:07:34.141 D5F7.1    stand  0    -90    80     111   80   152  
19:07:34.141 D5F7.0    stand  0    40     0      0     80   152  
19:07:35.142 D5F7.0    stand  0    40     0      104   80   0    
19:07:35.142 D5F7.1    stand  0    -90    80     0     80   152  
19:07:36.158 D5F7.1    stand  0    -100   80     0     80   10   
19:07:36.158 D5F7.0    stand  0    40     0      0     80   161  
19:07:37.063 D5F7.0    stand  0    40     0      0     80   0    
19:07:37.063 D5F7.1    stand  0    -90    80     0     80   152  
19:07:38.075 D5F7.0    stand  0    40     0      0     80   152  
19:07:38.075 D5F7.1    stand  0    -90    80     0     80   152  
19:07:39.092 D5F7.1    stand  0    -90    80     0     80   0    
19:07:39.092 D5F7.0    stand  0    40     0      0     80   152  
19:07:40.104 D5F7.0    stand  0    40     0      0     80   0    
19:07:40.104 D5F7.1    stand  0    -90    80     0     80   152  
19:07:41.184 D5F7.1    stand  0    -90    80     0     80   0    
19:07:41.184 D5F7.0    stand  0    40     0      0     80   152  
19:07:42.067 D5F7.0    stand  0    40     0      0     80   0    
19:07:43.064 D5F7.0    stand  0    40     0      0     80   0    
19:07:44.077 D5F7.0    stand  0    40     0      0     80   0    
19:07:45.095 D5F7.0    stand  0    40     0      0     80   0    
19:07:46.009 D5F7.0    stand  0    40     0      0     80   0    
19:07:47.009 D5F7.0    stand  0    40     0      0     80   0    
19:07:48.018 D5F7.0    stand  0    40     0      0     80   0    
19:07:49.034 D5F7.0    stand  0    40     0      0     80   0    
19:07:50.039 D5F7.0    stand  0    40     0      0     80   0    
19:07:50.955 D5F7.0    stand  0    40     0      0     80   0    
19:07:51.970 D5F7.0    stand  0    40     0      0     80   0    
19:07:53.049 D5F7.0    stand  0    40     -20    0     80   20   
19:07:54.018 D5F7.0    stand  0    40     -20    0     80   0    
19:07:54.918 D5F7.0    stand  0    40     -20    0     80   0    
19:07:55.930 D5F7.0    stand  0    40     -10    0     80   10   
19:07:56.943 D5F7.0    stand  0    40     -10    0     80   0    
19:07:57.960 D5F7.0    stand  0    40     -10    0     80   0    
19:07:58.967 D5F7.0    stand  0    40     -10    0     80   0    
19:07:59.876 D5F7.0    stand  0    40     -10    0     80   0    
19:08:00.898 D5F7.0    stand  0    50     0      71    80   14   
19:08:01.906 D5F7.0    stand  0    10     -40    93    80   56   
19:08:02.951 D5F7.0    stand  0    20     -30    99    80   14   
19:08:03.847 D5F7.0    stand  0    40     -20    68    80   22   
19:08:04.869 D5F7.0    stand  0    60     -10    0     80   22   
19:08:05.879 D5F7.0    stand  0    60     -10    0     80   0    
19:08:06.888 D5F7.0    stand  0    60     -10    0     80   0    
19:08:07.902 D5F7.0    stand  0    60     -10    0     80   0    
19:08:08.802 D5F7.0    stand  0    60     -10    0     80   0    
19:08:09.828 D5F7.0    stand  0    60     -10    0     80   0    
19:08:10.831 D5F7.0    stand  0    60     -10    0     80   0    
19:08:11.846 D5F7.0    stand  0    60     -10    0     80   0    
19:08:12.865 D5F7.0    stand  0    60     -10    0     80   0    
19:08:13.769 D5F7.0    stand  0    60     -10    0     80   0    
19:08:14.779 D5F7.0    stand  0    80     0      0     80   22   
19:08:15.794 D5F7.0    stand  0    80     0      0     80   0    
19:08:16.805 D5F7.0    stand  0    80     0      0     80   0    
19:08:17.818 D5F7.0    stand  0    80     0      0     80   0    
19:08:18.725 D5F7.0    stand  0    80     0      0     80   0    
19:08:19.739 D5F7.0    stand  0    80     0      0     80   0    
19:08:20.761 D5F7.0    stand  0    80     0      0     80   0    
19:08:21.765 D5F7.0    stand  0    80     0      0     80   0    
19:08:22.779 D5F7.0    stand  0    70     0      101   80   10   
19:08:23.689 D5F7.0    stand  0    40     0      0     80   30   
19:08:24.707 D5F7.0    stand  0    50     0      78    80   10   
19:08:25.711 D5F7.0    stand  0    70     0      89    80   20   
19:08:26.725 D5F7.0    stand  0    90     0      77    80   20   
19:08:27.742 D5F7.0    stand  0    70     0      111   80   20   
19:08:28.647 D5F7.0    stand  0    40     0      0     80   30   
19:08:29.657 D5F7.0    stand  0    40     0      0     80   0    
19:08:30.670 D5F7.0    stand  0    40     0      0     80   0    
19:08:31.685 D5F7.0    stand  0    40     0      0     80   0    
19:08:32.709 D5F7.0    stand  0    50     0      0     80   10   
19:08:33.609 D5F7.0    stand  0    50     0      0     80   0    
19:08:34.618 D5F7.0    stand  0    50     0      0     80   0    
19:08:35.631 D5F7.0    stand  0    50     0      0     80   0    
19:08:36.655 D5F7.0    stand  0    50     0      0     80   0    
19:08:37.661 D5F7.0    stand  0    50     0      0     80   0    
19:08:38.570 D5F7.0    stand  0    50     0      0     80   0    
19:08:39.579 D5F7.0    stand  0    50     0      0     80   0    
19:08:40.597 D5F7.0    stand  0    50     0      0     80   0    
19:08:41.629 D5F7.0    stand  0    50     0      0     80   0    
19:08:42.541 D5F7.0    stand  0    50     0      0     80   0    
19:08:43.547 D5F7.0    stand  0    50     0      0     80   0    
19:08:44.559 D5F7.0    stand  0    40     0      98    80   10   
19:08:45.576 D5F7.0    stand  0    50     -10    99    80   14   
19:08:46.585 D5F7.0    stand  0    40     0      0     80   14   
19:08:47.494 D5F7.0    stand  0    40     0      0     80   0    
19:08:48.517 D5F7.0    stand  0    40     0      0     80   0    
19:08:49.522 D5F7.0    stand  0    40     0      0     80   0    
19:08:50.544 D5F7.0    stand  0    40     0      0     80   0    
19:08:51.551 D5F7.0    stand  0    40     0      0     80   0    
19:08:52.458 D5F7.0    stand  0    40     0      0     80   0    
19:08:53.713 D5F7.0    stand  0    40     0      0     80   0    
19:08:54.439 D5F7.0    stand  0    40     0      0     80   0    
19:08:55.453 D5F7.0    stand  0    40     0      0     80   0    
19:08:56.466 D5F7.0    stand  0    40     0      0     80   0    
19:08:57.490 D5F7.0    stand  0    40     0      0     80   0    
19:08:58.493 D5F7.0    stand  0    40     0      0     80   0    
19:08:59.398 D5F7.0    stand  0    40     0      0     80   0    
19:09:00.412 D5F7.0    stand  0    40     0      0     80   0    
19:09:01.435 D5F7.0    stand  0    40     0      0     80   0    
19:09:02.442 D5F7.0    stand  0    40     0      0     80   0    
19:09:03.451 D5F7.0    stand  0    40     0      0     80   0    
19:09:04.362 D5F7.0    stand  0    40     0      0     80   0    
19:09:05.378 D5F7.0    stand  0    40     0      0     80   0    
19:09:06.385 D5F7.0    stand  0    40     0      0     80   0    
19:09:07.396 D5F7.0    stand  0    40     0      0     80   0    
19:09:08.409 D5F7.0    stand  0    40     0      0     80   0    
19:09:09.316 D5F7.0    stand  0    40     0      0     80   0    
19:09:10.331 D5F7.0    stand  0    40     0      0     80   0    
19:09:11.347 D5F7.0    stand  0    40     0      0     80   0    
19:09:12.356 D5F7.0    stand  0    40     0      0     80   0    
19:09:13.374 D5F7.0    stand  0    40     0      0     80   0    
19:09:14.276 D5F7.0    stand  0    40     0      0     80   0    
19:09:15.294 D5F7.0    stand  0    40     0      0     80   0    
19:09:16.327 D5F7.0    stand  0    40     0      0     80   0    
19:09:17.324 D5F7.0    stand  0    40     0      0     80   0    
19:09:18.334 D5F7.0    stand  0    50     -10    101   80   14   
19:09:19.241 D5F7.0    stand  0    100    0      115   80   50   
19:09:20.253 D5F7.0    stand  0    50     0      0     80   50   
19:09:21.266 D5F7.0    stand  0    40     0      0     80   10   
19:09:22.282 D5F7.0    stand  0    40     0      0     80   0    
19:09:23.305 D5F7.0    stand  0    40     0      0     80   0    
19:09:24.204 D5F7.0    stand  0    40     0      0     80   0    
19:09:25.211 D5F7.0    stand  0    40     0      0     80   0    
19:09:26.231 D5F7.0    stand  0    40     0      0     80   0    
19:09:27.237 D5F7.0    stand  0    40     0      0     80   0    
19:09:28.249 D5F7.0    stand  0    40     0      98    80   0    
19:09:29.161 D5F7.0    stand  0    40     0      0     80   0    
19:09:30.149 D5F7.0    stand  0    40     0      0     80   0    
19:09:31.168 D5F7.0    stand  0    40     0      0     80   0    
19:09:32.184 D5F7.0    stand  0    40     0      0     80   0    
19:09:33.188 D5F7.0    stand  0    40     0      0     80   0    
19:09:34.201 D5F7.0    stand  0    40     0      0     80   0    
19:09:35.111 D5F7.0    stand  0    40     0      0     80   0    
19:09:36.127 D5F7.0    stand  0    40     0      0     80   0    
19:09:37.142 D5F7.0    stand  0    40     0      0     80   0    
19:09:38.152 D5F7.0    stand  0    40     0      0     80   0    
19:09:39.166 D5F7.0    stand  0    40     0      0     80   0    
19:09:40.069 D5F7.0    stand  0    40     0      0     80   0    
19:09:41.093 D5F7.0    stand  0    40     0      0     80   0    
19:09:42.095 D5F7.0    stand  0    40     0      0     80   0    
19:09:43.129 D5F7.0    stand  0    40     0      0     80   0    
19:09:44.130 D5F7.0    stand  0    40     0      0     80   0    
19:09:45.037 D5F7.0    stand  0    40     0      0     80   0    
19:09:46.091 D5F7.0    stand  0    40     0      0     80   0    
19:09:47.096 D5F7.0    stand  0    50     0      101   80   10   
19:09:48.105 D5F7.0    stand  0    140    0      0     80   90   
19:09:49.007 D5F7.0    stand  0    130    0      109   80   10   
19:09:50.025 D5F7.0    stand  0    80     0      0     80   50   
19:09:51.051 D5F7.0    stand  0    40     0      101   80   40   
19:09:52.120 D5F7.0    stand  0    40     0      102   80   0    
19:09:52.971 D5F7.0    stand  0    40     0      0     80   0    
19:09:53.980 D5F7.0    stand  0    40     0      0     80   0    
19:09:54.993 D5F7.0    stand  0    40     0      0     80   0    
19:09:56.006 D5F7.0    stand  0    40     0      0     80   0    
19:09:57.026 D5F7.0    stand  0    40     0      0     80   0    
19:09:57.928 D5F7.0    stand  0    40     0      0     80   0    
19:09:58.941 D5F7.0    stand  0    40     0      0     80   0    
19:09:59.955 D5F7.0    stand  0    40     0      0     80   0    
19:10:00.982 D5F7.0    stand  0    40     0      0     80   0    
19:10:01.979 D5F7.0    stand  0    40     0      0     80   0    
19:10:02.885 D5F7.0    stand  0    40     0      0     80   0    
19:10:03.904 D5F7.0    stand  0    40     0      0     80   0    
19:10:04.911 D5F7.0    stand  0    40     0      0     80   0    
19:10:05.927 D5F7.0    stand  0    40     0      0     80   0    
19:10:06.941 D5F7.0    stand  0    40     0      0     80   0    
19:10:07.849 D5F7.0    stand  0    40     0      0     80   0    
19:10:08.863 D5F7.0    stand  0    40     0      0     80   0    
19:10:09.881 D5F7.0    stand  0    40     0      0     80   0    
19:10:10.884 D5F7.0    stand  0    40     0      0     80   0    
19:10:11.910 D5F7.0    stand  0    40     0      0     80   0    
19:10:12.810 D5F7.0    stand  0    40     0      0     80   0    
19:10:13.819 D5F7.0    stand  0    40     0      0     80   0    
19:10:14.834 D5F7.0    stand  0    40     0      0     80   0    
19:10:15.847 D5F7.0    stand  0    40     0      0     80   0    
19:10:16.859 D5F7.0    stand  0    40     0      0     80   0    
19:10:17.768 D5F7.0    stand  0    40     0      98    80   0    
19:10:18.784 D5F7.0    stand  0    40     0      0     80   0    
19:10:19.797 D5F7.0    stand  0    40     0      0     80   0    
19:10:20.813 D5F7.0    stand  0    50     0      0     80   10   
19:10:21.830 D5F7.0    stand  0    50     0      0     80   0    
19:10:22.726 D5F7.0    stand  0    50     0      0     80   0    
19:10:23.737 D5F7.0    stand  0    50     0      0     80   0    
19:10:24.758 D5F7.0    stand  0    50     0      0     80   0    
19:10:25.773 D5F7.0    stand  0    50     0      0     80   0    
19:10:26.789 D5F7.0    stand  0    50     0      0     80   0    
19:10:27.685 D5F7.0    stand  0    50     0      0     80   0    
19:10:28.705 D5F7.0    stand  0    50     0      0     80   0    
19:10:29.718 D5F7.0    stand  0    50     0      0     80   0    
19:10:30.725 D5F7.0    stand  0    50     0      0     80   0    
19:10:31.738 D5F7.0    stand  0    50     0      0     80   0    
19:10:32.649 D5F7.0    stand  0    50     0      0     80   0    
19:10:33.710 D5F7.0    stand  0    50     0      0     80   0    
19:10:34.723 D5F7.0    stand  0    50     0      0     80   0    
19:10:35.620 D5F7.0    stand  0    50     0      0     80   0    
19:10:36.639 D5F7.0    stand  0    50     0      0     80   0    
19:10:37.650 D5F7.0    stand  0    40     0      98    80   10   
19:10:38.667 D5F7.0    stand  0    80     -20    81    80   44   
19:10:39.678 D5F7.0    stand  0    60     -10    0     80   22   
19:10:40.588 D5F7.0    stand  0    60     -10    0     80   0    
19:10:41.597 D5F7.0    stand  0    60     -10    0     80   0    
19:10:42.607 D5F7.0    stand  0    60     -10    0     80   0    
19:10:43.630 D5F7.0    stand  0    60     -10    0     80   0    
19:10:44.633 D5F7.0    stand  0    60     -10    0     80   0    
19:10:45.543 D5F7.0    stand  0    60     -10    0     80   0    
19:10:46.557 D5F7.0    stand  0    60     -10    0     80   0    
19:10:47.579 D5F7.0    stand  0    60     -10    0     80   0    
19:10:48.598 D5F7.0    stand  0    60     -10    0     80   0    
19:10:49.607 D5F7.0    stand  0    60     -10    0     80   0    
19:10:50.500 D5F7.0    stand  0    60     -10    0     80   0    
19:10:51.518 D5F7.0    stand  0    60     -10    0     80   0    
19:10:52.610 D5F7.0    stand  0    60     -10    0     80   0    
19:10:53.563 D5F7.0    stand  0    60     -10    0     80   0    
19:10:54.482 D5F7.0    stand  0    60     -10    0     80   0    
19:10:55.492 D5F7.0    stand  0    60     -10    0     80   0    
19:10:56.506 D5F7.0    stand  0    60     -10    0     80   0    
19:10:57.510 D5F7.0    stand  0    60     -10    0     80   0    
19:10:58.528 D5F7.0    stand  0    60     -10    0     80   0    
19:10:59.436 D5F7.0    stand  0    60     -10    0     80   0    
19:11:00.449 D5F7.0    stand  0    60     -10    0     80   0    
19:11:01.461 D5F7.0    stand  0    60     -10    0     80   0    
19:11:02.471 D5F7.0    stand  0    60     -10    0     80   0    
19:11:03.482 D5F7.0    stand  0    60     -10    0     80   0    
19:11:04.393 D5F7.0    stand  0    60     -10    0     80   0    
19:11:05.403 D5F7.0    stand  0    60     -10    0     80   0    
19:11:06.428 D5F7.0    stand  0    60     -10    0     80   0    
19:11:07.452 D5F7.0    stand  0    60     -10    0     80   0    
19:11:08.441 D5F7.0    stand  0    70     -10    65    80   10   
19:11:09.352 D5F7.0    stand  0    90     -20    63    80   22   
19:11:10.370 D5F7.0    stand  0    40     -30    73    80   50   
19:11:11.382 D5F7.0    walk   0    30     -50    88    80   22   
19:11:12.400 D5F7.0    walk   0    90     -10    86    80   72   
19:11:13.409 D5F7.0    walk   0    110    -20    93    80   22   
19:11:14.317 D5F7.0    walk   0    110    0      66    80   20   
19:11:15.325 D5F7.0    walk   0    80     0      93    80   30   
19:11:16.337 D5F7.0    stand  0    60     0      0     80   20   
19:11:17.350 D5F7.0    stand  0    60     0      0     80   0    
19:11:18.377 D5F7.0    stand  0    100    0      0     80   40   
19:11:19.270 D5F7.0    stand  0    100    0      112   80   0    
19:11:20.286 D5F7.0    stand  0    100    0      66    80   0    
19:11:21.301 D5F7.0    stand  0    50     0      0     80   50   
19:11:22.248 D5F7.0    stand  0    50     0      0     80   0    
19:11:23.260 D5F7.0    stand  0    40     0      0     80   10   
19:11:24.271 D5F7.0    stand  0    50     0      0     80   10   
19:11:25.285 D5F7.0    stand  0    50     0      0     80   0    
19:11:26.298 D5F7.0    stand  0    50     0      0     80   0    
19:11:27.205 D5F7.0    stand  0    50     0      0     80   0    
19:11:28.225 D5F7.0    stand  0    50     0      0     80   0    
19:11:29.238 D5F7.0    stand  0    50     0      0     80   0    
19:11:30.245 D5F7.0    stand  0    50     -10    104   80   10   
19:11:31.266 D5F7.0    stand  0    60     0      0     80   14   
19:11:32.168 D5F7.0    stand  0    40     0      0     80   20   
19:11:33.181 D5F7.0    stand  0    40     0      0     80   0    
19:11:34.195 D5F7.0    stand  0    40     0      0     80   0    
19:11:35.204 D5F7.0    stand  0    40     0      0     80   0    
19:11:36.217 D5F7.0    stand  0    40     -10    0     80   10   
19:11:37.128 D5F7.0    stand  0    40     -10    0     80   0    
19:11:38.161 D5F7.0    stand  0    40     -10    0     80   0    
19:11:39.172 D5F7.0    stand  0    40     -10    0     80   0    
19:11:40.192 D5F7.0    stand  0    40     -10    0     80   0    
19:11:41.101 D5F7.0    stand  0    40     -10    0     80   0    
19:11:42.111 D5F7.0    stand  0    40     -10    0     80   0    
19:11:43.130 D5F7.0    stand  0    40     -10    0     80   0    
19:11:44.137 D5F7.0    stand  0    50     0      0     80   14   
19:11:45.146 D5F7.0    stand  0    40     0      0     80   10   
19:11:46.073 D5F7.0    stand  0    40     0      0     80   0    
19:11:47.072 D5F7.0    stand  0    40     0      0     80   0    
19:11:48.099 D5F7.0    stand  0    40     0      0     80   0    
19:11:49.116 D5F7.0    stand  0    40     0      0     80   0    
19:11:50.121 D5F7.0    stand  0    40     0      0     80   0    
19:11:51.022 D5F7.0    stand  0    40     0      0     80   0    
19:11:52.117 D5F7.0    stand  0    40     0      0     80   0    
19:11:53.066 D5F7.0    stand  0    80     0      0     80   40   
19:11:54.001 D5F7.0    stand  0    150    -10    0     80   70   
19:11:55.002 D5F7.0    stand  0    150    -10    0     80   0    
19:11:56.019 D5F7.0    stand  0    150    -10    0     80   0    
19:11:57.031 D5F7.0    stand  0    150    -10    0     80   0    
19:11:58.099 D5F7.1    stand  255  40     0      78    80   110  
19:11:58.099 D5F7.0    stand  0    140    -20    0     80   101  
19:11:58.959 D5F7.1    stand  255  60     -10    77    80   80   
19:11:58.959 D5F7.0    stand  0    160    -10    0     80   100  
19:11:59.978 D5F7.0    stand  0    160    -30    0     80   20   
19:11:59.978 D5F7.1    stand  255  40     0      105   80   123  
19:12:00.991 D5F7.1    stand  255  50     -10    0     80   14   
19:12:00.991 D5F7.0    stand  0    160    -40    0     80   114  
19:12:01.997 D5F7.1    stand  255  50     -20    0     80   111  
19:12:01.997 D5F7.0    stand  0    160    -40    0     80   111  
19:12:03.012 D5F7.1    stand  255  50     -10    0     80   114  
19:12:03.012 D5F7.0    stand  0    160    -40    0     80   114  
19:12:03.918 D5F7.1    stand  255  40     0      0     80   126  
19:12:03.918 D5F7.0    stand  0    160    -40    0     80   126  
19:12:04.923 D5F7.1    stand  255  40     0      0     80   126  
19:12:04.923 D5F7.0    stand  0    160    -40    0     80   126  
19:12:05.940 D5F7.1    stand  255  40     0      0     80   126  
19:12:05.940 D5F7.0    stand  0    160    -40    0     80   126  
19:12:06.950 D5F7.0    stand  0    160    -40    0     80   0    
19:12:06.950 D5F7.1    stand  255  40     0      0     80   126  
19:12:07.971 D5F7.0    stand  0    160    -40    0     80   126  
19:12:07.971 D5F7.1    stand  255  40     0      0     80   126  
19:12:08.877 D5F7.0    stand  0    160    -30    0     80   123  
19:12:08.877 D5F7.1    stand  255  40     0      0     80   123  
19:12:09.874 D5F7.0    stand  0    160    -30    0     80   123  
19:12:09.874 D5F7.1    stand  255  40     0      0     80   123  
19:12:10.874 D5F7.1    stand  255  40     0      0     80   0    
19:12:10.874 D5F7.0    stand  0    160    -30    0     80   123  
19:12:11.895 D5F7.1    stand  255  40     0      0     80   123  
19:12:11.895 D5F7.0    stand  0    160    -30    0     80   123  
19:12:12.907 D5F7.0    stand  0    160    -30    0     80   0    
19:12:12.907 D5F7.1    stand  255  40     0      0     80   123  
19:12:13.919 D5F7.0    stand  0    160    -30    0     80   123  
19:12:13.919 D5F7.1    stand  255  40     -10    0     80   121  
19:12:14.827 D5F7.1    stand  255  40     -20    0     80   10   
19:12:14.827 D5F7.0    stand  0    160    -30    0     80   120  
19:12:15.847 D5F7.0    stand  0    160    -30    0     80   0    
19:12:15.847 D5F7.1    stand  255  40     -10    97    80   121  
19:12:16.860 D5F7.1    stand  255  40     0      0     80   10   
19:12:16.860 D5F7.0    stand  0    160    -30    0     80   123  
19:12:17.866 D5F7.1    stand  255  40     0      72    80   123  
19:12:17.866 D5F7.0    stand  0    160    -30    0     80   123  
19:12:18.878 D5F7.0    stand  0    160    -30    0     80   0    
19:12:18.878 D5F7.1    stand  255  40     0      0     80   123  
19:12:19.786 D5F7.1    stand  255  40     0      0     80   0    
19:12:19.786 D5F7.0    stand  0    160    -30    0     80   123  
19:12:20.797 D5F7.0    stand  0    160    -30    0     80   0    
19:12:20.797 D5F7.1    stand  255  40     0      0     80   123  
19:12:21.823 D5F7.1    stand  255  40     0      0     80   0    
19:12:21.823 D5F7.0    stand  0    160    -30    0     80   123  
19:12:22.828 D5F7.0    stand  0    160    -30    0     80   0    
19:12:22.828 D5F7.1    stand  255  40     0      0     80   123  
19:12:23.842 D5F7.0    stand  0    160    -30    0     80   123  
19:12:23.842 D5F7.1    stand  255  40     0      0     80   123  
19:12:24.765 D5F7.0    stand  0    160    -30    0     80   123  
19:12:24.765 D5F7.1    stand  255  40     0      0     80   123  
19:12:25.791 D5F7.0    stand  0    160    -30    0     80   123  
19:12:25.791 D5F7.1    stand  255  40     0      0     80   123  
19:12:26.805 D5F7.0    stand  0    160    -30    0     80   123  
19:12:26.805 D5F7.1    stand  255  50     0      0     80   114  
19:12:27.825 D5F7.0    stand  0    160    -30    0     80   114  
19:12:27.825 D5F7.1    stand  255  50     0      0     80   114  
19:12:28.718 D5F7.0    stand  0    160    -30    0     80   114  
19:12:28.718 D5F7.1    stand  255  40     0      104   80   123  
19:12:29.728 D5F7.0    stand  0    160    -30    0     80   123  
19:12:29.728 D5F7.1    stand  255  40     0      0     80   123  
19:12:30.740 D5F7.1    stand  255  40     0      0     80   0    
19:12:30.740 D5F7.0    stand  0    160    -30    0     80   123  
19:12:31.757 D5F7.0    stand  0    160    -30    0     80   0    
19:12:31.757 D5F7.1    stand  255  40     0      0     80   123  
19:12:32.772 D5F7.1    stand  255  40     0      0     80   0    
19:12:32.772 D5F7.0    stand  0    160    -30    0     80   123  
19:12:33.675 D5F7.1    stand  255  40     0      0     80   123  
19:12:33.675 D5F7.0    stand  0    160    -30    0     80   123  
19:12:34.689 D5F7.0    stand  0    160    -30    0     80   0    
19:12:34.689 D5F7.1    stand  255  40     0      0     80   123  
19:12:35.697 D5F7.1    stand  255  40     0      0     80   0    
19:12:35.697 D5F7.0    stand  0    160    -30    0     80   123  
19:12:36.719 D5F7.0    stand  0    160    -30    0     80   0    
19:12:36.719 D5F7.1    stand  255  40     0      0     80   123  
19:12:37.732 D5F7.1    stand  255  40     0      0     80   0    
19:12:37.732 D5F7.0    stand  0    160    -30    0     80   123  
19:12:38.640 D5F7.1    stand  255  40     0      0     80   123  
19:12:38.640 D5F7.0    stand  0    160    -30    0     80   123  
19:12:39.655 D5F7.0    stand  0    160    -30    0     80   0    
19:12:39.655 D5F7.1    stand  255  40     0      0     80   123  
19:12:40.679 D5F7.1    stand  255  40     0      0     80   0    
19:12:40.679 D5F7.0    stand  0    160    -30    0     80   123  
19:12:41.676 D5F7.1    stand  255  40     0      0     80   123  
19:12:41.676 D5F7.0    stand  0    160    -30    0     80   123  
19:12:42.690 D5F7.0    stand  0    160    -30    0     80   0    
19:12:42.690 D5F7.1    stand  255  40     0      0     80   123  
19:12:43.604 D5F7.1    stand  255  40     0      0     80   0    
19:12:43.604 D5F7.0    stand  0    160    -30    0     80   123  
19:12:44.605 D5F7.1    stand  255  40     0      0     80   123  
19:12:44.605 D5F7.0    stand  0    160    -30    0     80   123  
19:12:45.617 D5F7.0    stand  0    160    -30    0     80   0    
19:12:45.617 D5F7.1    stand  255  40     0      0     80   123  
19:12:46.634 D5F7.0    stand  0    160    -30    0     80   123  
19:12:46.634 D5F7.1    stand  255  40     0      0     80   123  
19:12:47.645 D5F7.1    stand  255  40     0      0     80   0    
19:12:47.645 D5F7.0    stand  0    160    -30    0     80   123  
19:12:48.608 D5F7.1    stand  255  40     0      0     80   123  
19:12:48.608 D5F7.0    stand  0    160    -30    0     80   123  
19:12:49.577 D5F7.1    stand  255  40     0      0     80   123  
19:12:49.577 D5F7.0    stand  0    160    -30    0     80   123  
19:12:50.665 D5F7.0    stand  0    160    -30    0     80   0    
19:12:50.665 D5F7.1    stand  255  40     0      0     80   123  
19:12:51.625 D5F7.1    stand  255  40     0      0     80   0    
19:12:51.625 D5F7.0    stand  0    160    -30    0     80   123  
19:12:52.528 D5F7.1    stand  255  40     0      0     80   123  
19:12:52.528 D5F7.0    stand  0    160    -30    0     80   123  
19:12:53.536 D5F7.1    stand  255  40     0      0     80   123  
19:12:53.536 D5F7.0    stand  0    160    -30    0     80   123  
19:12:54.548 D5F7.0    stand  0    160    -30    0     80   0    
19:12:54.548 D5F7.1    stand  255  40     0      0     80   123  
19:12:55.574 D5F7.1    stand  255  40     0      0     80   0    
19:12:55.574 D5F7.0    stand  0    160    -30    0     80   123  
19:12:56.587 D5F7.1    stand  255  40     0      0     80   123  
19:12:56.587 D5F7.0    stand  0    160    -30    0     80   123  
19:12:57.492 D5F7.0    stand  0    160    -30    0     80   0    
19:12:57.492 D5F7.1    stand  255  40     0      0     80   123  
19:12:58.548 D5F7.1    stand  255  40     0      0     80   0    
19:12:59.541 D5F7.1    stand  255  40     0      0     80   0    
19:13:00.544 D5F7.1    stand  255  40     0      0     80   0    
19:13:01.453 D5F7.1    stand  255  40     0      0     80   0    
19:13:02.460 D5F7.1    stand  255  40     0      0     80   0    
19:13:03.471 D5F7.1    stand  255  40     0      0     80   0    
19:13:04.495 D5F7.1    stand  255  40     0      0     80   0    
19:13:05.509 D5F7.1    stand  255  40     0      0     80   0    
19:13:06.420 D5F7.1    stand  255  40     0      0     80   0    
19:13:07.425 D5F7.1    stand  255  40     0      0     80   0    
19:13:08.439 D5F7.1    stand  255  40     0      0     80   0    
19:13:09.532 D5F7.1    stand  255  40     0      0     80   0    
19:13:10.376 D5F7.1    stand  255  40     0      0     80   0    
19:13:11.400 D5F7.1    stand  255  40     0      0     80   0    
19:13:12.403 D5F7.1    stand  255  40     0      0     80   0    
19:13:13.415 D5F7.1    stand  255  40     0      0     80   0    
19:13:14.429 D5F7.1    stand  255  50     0      0     80   10   
19:13:15.342 D5F7.1    stand  255  50     0      0     80   0    
19:13:16.362 D5F7.1    stand  255  80     -10    93    80   31   
19:13:17.371 D5F7.1    stand  255  40     0      0     80   41   
19:13:18.388 D5F7.1    stand  255  40     0      0     80   0    
19:13:19.396 D5F7.1    stand  255  40     0      0     80   0    
19:13:20.303 D5F7.1    stand  255  40     0      0     80   0    
19:13:21.310 D5F7.1    stand  255  40     0      0     80   0    
19:13:22.330 D5F7.1    walk   255  100    -10    118   80   60   
19:13:23.336 D5F7.1    walk   255  170    -10    103   80   70   
19:13:24.363 D5F7.1    walk   255  150    0      85    80   22   
19:13:25.258 D5F7.1    walk   255  110    0      0     80   40   
19:13:26.274 D5F7.1    walk   255  170    0      0     80   60   
19:13:27.291 D5F7.1    sit    255  160    0      0     80   10   
19:13:28.314 D5F7.1    sit    255  160    0      0     80   0    
19:13:29.316 D5F7.1    sit    255  160    0      0     80   0    
19:13:30.220 D5F7.1    sit    255  160    0      0     80   0    
19:13:31.239 D5F7.1    sit    255  160    0      0     80   0    
19:13:32.248 D5F7.1    sit    255  160    0      0     80   0    
19:13:33.259 D5F7.1    sit    255  160    0      0     80   0    
19:13:34.278 D5F7.1    sit    255  160    0      0     80   0    
19:13:35.187 D5F7.1    sit    255  160    0      0     80   0    
19:13:36.194 D5F7.1    sit    255  160    0      0     80   0    
19:13:37.207 D5F7.1    sit    255  160    0      0     80   0    
19:13:38.212 D5F7.1    sit    255  160    0      0     80   0    
19:13:39.239 D5F7.1    sit    255  160    0      0     80   0    
19:13:40.134 D5F7.1    sit    255  160    0      74    80   0    
19:13:41.159 D5F7.1    sit    255  160    -10    0     80   10   
19:13:42.168 D5F7.1    sit    255  160    -10    0     80   0    
19:13:43.184 D5F7.1    sit    255  160    -10    0     80   0    
19:13:44.199 D5F7.1    sit    255  160    -10    0     80   0    
19:13:45.101 D5F7.1    sit    255  160    -10    0     80   0    
19:13:46.118 D5F7.1    sit    255  160    -10    0     80   0    
19:13:47.187 D5F7.1    sit    255  160    0      0     80   10   
19:13:47.187 D5F7.0    stand  255  40     0      71    80   120  
19:13:48.158 D5F7.0    stand  255  40     0      65    80   0    
19:13:48.158 D5F7.1    sit    255  150    -10    0     80   110  
19:13:49.068 D5F7.1    sit    255  160    -10    0     80   10   
19:13:49.068 D5F7.0    stand  255  30     0      76    80   130  
19:13:50.085 D5F7.0    stand  255  40     0      0     80   10   
19:13:50.085 D5F7.1    sit    255  160    0      0     80   120  
19:13:51.623 D5F7.1    sit    255  160    0      0     80   0    
19:13:51.623 D5F7.0    stand  255  40     0      0     80   120  
19:13:52.131 D5F7.0    stand  255  40     0      0     80   0    
19:13:52.131 D5F7.1    sit    255  160    0      0     80   120  
19:13:53.036 D5F7.0    stand  255  40     0      0     80   120  
19:13:53.036 D5F7.1    sit    255  160    0      0     80   120  
19:13:54.046 D5F7.1    sit    255  160    0      0     80   0    
19:13:54.046 D5F7.0    stand  255  40     0      0     80   120  
19:13:55.064 D5F7.1    sit    255  160    0      0     80   120  
19:13:55.064 D5F7.0    stand  255  40     0      0     80   120  
19:13:56.077 D5F7.0    stand  255  40     0      0     80   0    
19:13:56.077 D5F7.1    sit    255  160    0      0     80   120  
19:13:57.096 D5F7.1    sit    255  160    -20    0     80   20   
19:13:57.096 D5F7.0    stand  255  40     0      80    80   121  
19:13:58.017 D5F7.0    stand  255  40     -10    63    80   10   
19:13:58.017 D5F7.1    sit    255  160    -20    0     80   120  
19:13:59.020 D5F7.1    sit    255  150    -20    0     80   10   
19:13:59.020 D5F7.0    stand  255  30     -10    50    80   120  
19:14:00.031 D5F7.0    stand  255  40     -10    47    80   10   
19:14:00.031 D5F7.1    sit    255  160    -20    0     80   120  
19:14:01.049 D5F7.1    sit    255  160    -10    0     80   10   
19:14:01.049 D5F7.0    stand  255  40     -10    55    80   120  
19:14:02.059 D5F7.1    stand  255  160    -50    0     80   126  
19:14:02.059 D5F7.0    stand  255  40     -10    0     80   126  
19:14:02.956 D5F7.1    stand  255  160    -50    0     80   126  
19:14:02.956 D5F7.0    stand  255  40     -10    0     80   126  
19:14:03.979 D5F7.1    stand  255  160    -50    0     80   126  
19:14:03.979 D5F7.0    stand  255  40     -10    0     80   126  
19:14:04.980 D5F7.0    stand  255  50     -10    0     80   10   
19:14:04.980 D5F7.1    stand  255  140    -30    0     80   92   
19:14:05.992 D5F7.1    stand  255  140    -20    0     80   10   
19:14:05.992 D5F7.0    stand  255  60     0      0     80   82   
19:14:07.010 D5F7.1    stand  255  140    -20    0     80   82   
19:14:07.010 D5F7.0    stand  255  120    -50    0     80   36   
19:14:07.919 D5F7.1    stand  255  140    -20    0     80   36   
19:14:07.919 D5F7.0    stand  255  160    -20    0     80   20   
19:14:08.932 D5F7.1    stand  255  140    -20    0     80   20   
19:14:08.932 D5F7.0    stand  255  160    -20    0     80   20   
19:14:09.946 D5F7.1    stand  255  140    -20    0     80   20   
19:14:09.946 D5F7.0    stand  255  160    -20    0     80   20   
19:14:10.961 D5F7.0    stand  255  150    -40    0     80   22   
19:14:10.961 D5F7.1    stand  255  140    -20    0     80   22   
19:14:11.965 D5F7.0    stand  255  130    -70    0     80   50   
19:14:11.965 D5F7.1    stand  255  140    -20    0     80   50   
19:14:12.878 D5F7.1    stand  255  140    -20    0     80   0    
19:14:12.878 D5F7.0    stand  255  130    -70    0     80   50   
19:14:13.887 D5F7.0    stand  255  130    -70    0     80   0    
19:14:13.887 D5F7.1    stand  255  140    -20    0     80   50   
19:14:14.913 D5F7.0    stand  255  130    -70    0     80   50   
19:14:14.913 D5F7.1    stand  255  140    -20    0     80   50   
19:14:15.916 D5F7.1    stand  255  140    -20    0     80   0    
19:14:15.916 D5F7.0    stand  255  120    -60    0     80   44   
19:14:16.930 D5F7.0    stand  255  120    -60    0     80   0    
19:14:16.930 D5F7.1    stand  255  140    -20    0     80   44   
19:14:17.836 D5F7.1    stand  255  140    -20    0     80   0    
19:14:17.836 D5F7.0    stand  255  130    -60    0     80   41   
19:14:18.849 D5F7.1    stand  255  140    -20    0     80   41   
19:14:18.849 D5F7.0    stand  255  130    -60    0     80   41   
19:14:19.864 D5F7.1    stand  255  140    -20    0     80   41   
19:14:19.864 D5F7.0    stand  255  130    -70    0     80   50   
19:14:20.872 D5F7.1    stand  255  140    -20    0     80   50   
19:14:20.872 D5F7.0    stand  255  130    -70    0     80   50   
19:14:21.897 D5F7.0    stand  255  130    -70    0     80   0    
19:14:21.897 D5F7.1    stand  255  140    -20    0     80   50   
19:14:22.804 D5F7.1    stand  255  140    -20    0     80   0    
19:14:22.804 D5F7.0    stand  255  130    -70    0     80   50   
19:14:23.808 D5F7.1    stand  255  140    -20    0     80   50   
19:14:23.808 D5F7.0    stand  255  130    -70    0     80   50   
19:14:24.821 D5F7.1    stand  255  140    -20    0     80   50   
19:14:24.821 D5F7.0    stand  255  130    -70    0     80   50   
19:14:25.838 D5F7.1    stand  255  140    -20    0     80   50   
19:14:25.838 D5F7.0    stand  255  130    -70    0     80   50   
19:14:26.846 D5F7.0    stand  255  130    -70    0     80   0    
19:14:26.846 D5F7.1    stand  255  140    -20    0     80   50   
19:14:27.753 D5F7.1    stand  255  140    -20    0     80   0    
19:14:27.753 D5F7.0    stand  255  130    -70    0     80   50   
19:14:28.766 D5F7.1    stand  255  140    -20    0     80   50   
19:14:28.766 D5F7.0    stand  255  130    -70    0     80   50   
19:14:29.782 D5F7.0    stand  255  100    -40    0     80   42   
19:14:29.782 D5F7.1    stand  255  140    -20    0     80   44   
19:14:30.795 D5F7.1    stand  255  140    -20    0     80   0    
19:14:30.795 D5F7.0    stand  255  160    -50    0     80   36   
19:14:31.817 D5F7.0    stand  255  150    -50    0     80   10   
19:14:31.817 D5F7.1    stand  255  140    -20    0     80   31   
19:14:32.719 D5F7.0    stand  255  150    -50    0     80   31   
19:14:32.719 D5F7.1    stand  255  140    -20    0     80   31   
19:14:33.736 D5F7.0    stand  255  150    -50    0     80   31   
19:14:33.736 D5F7.1    stand  255  140    -20    0     80   31   
19:14:34.749 D5F7.0    stand  255  140    -40    0     80   20   
19:14:34.749 D5F7.1    stand  255  140    -20    0     80   20   
19:14:35.765 D5F7.1    stand  255  140    -20    0     80   0    
19:14:35.765 D5F7.0    stand  255  140    -40    0     80   20   
19:14:36.772 D5F7.0    stand  255  100    -40    0     80   40   
19:14:36.772 D5F7.1    stand  255  140    -20    0     80   44   
19:14:37.675 D5F7.0    stand  255  100    -30    0     80   41   
19:14:37.675 D5F7.1    stand  255  140    -20    0     80   41   
19:14:38.692 D5F7.1    stand  255  140    -20    0     80   0    
19:14:38.692 D5F7.0    stand  255  100    -30    0     80   41   
19:14:39.704 D5F7.1    stand  255  160    -10    0     80   63   
19:14:39.704 D5F7.0    stand  255  70     -20    92    80   90   
19:14:40.729 D5F7.0    stand  255  40     0      0     80   36   
19:14:40.729 D5F7.1    stand  255  150    0      0     80   110  
19:14:41.751 D5F7.0    stand  255  40     0      0     80   110  
19:14:41.751 D5F7.1    stand  255  150    0      0     80   110  
19:14:42.636 D5F7.1    stand  255  150    0      0     80   0    
19:14:42.636 D5F7.0    stand  255  40     0      0     80   110  
19:14:43.646 D5F7.1    stand  255  150    0      0     80   110  
19:14:43.646 D5F7.0    stand  255  40     0      0     80   110  
19:14:44.663 D5F7.1    stand  255  150    0      0     80   110  
19:14:44.663 D5F7.0    stand  255  40     0      0     80   110  
19:14:45.672 D5F7.1    stand  255  150    0      0     80   110  
19:14:45.672 D5F7.0    stand  255  40     0      0     80   110  
19:14:46.690 D5F7.1    stand  255  150    0      0     80   110  
19:14:46.690 D5F7.0    stand  255  40     0      0     80   110  
19:14:47.605 D5F7.0    stand  255  40     0      0     80   0    
19:14:47.605 D5F7.1    stand  255  150    0      0     80   110  
19:14:48.619 D5F7.0    stand  255  40     0      0     80   110  
19:14:48.619 D5F7.1    stand  255  150    0      0     80   110  
19:14:49.706 D5F7.1    stand  255  150    -10    0     80   10   
19:14:49.706 D5F7.0    stand  255  40     0      83    80   110  
19:14:50.574 D5F7.1    stand  255  150    -10    82    80   110  
19:14:50.574 D5F7.0    stand  255  40     -10    0     80   110  
19:14:51.586 D5F7.1    stand  255  130    -30    0     80   92   
19:14:51.586 D5F7.0    stand  255  30     -20    64    80   100  
19:14:52.603 D5F7.1    stand  255  130    -50    0     80   104  
19:14:52.603 D5F7.0    stand  255  10     -30    76    80   121  
19:14:53.610 D5F7.1    stand  255  130    -50    0     80   121  
19:14:53.610 D5F7.0    stand  255  0      -40    71    80   130  
19:14:54.631 D5F7.1    stand  255  150    -30    0     80   150  
19:14:54.631 D5F7.0    stand  255  20     -10    56    80   131  
19:14:55.538 D5F7.1    stand  255  160    -20    0     80   140  
19:14:55.538 D5F7.0    stand  255  20     -30    0     80   140  
19:14:56.545 D5F7.1    stand  255  150    -20    0     80   130  
19:14:56.545 D5F7.0    stand  255  20     -30    63    80   130  
19:14:57.560 D5F7.0    stand  255  30     -10    75    80   22   
19:14:57.560 D5F7.1    stand  255  150    -10    0     80   120  
19:14:58.582 D5F7.0    stand  255  40     -10    80    80   110  
19:14:58.582 D5F7.1    stand  255  160    -10    0     80   120  
19:14:59.599 D5F7.1    stand  255  150    -10    0     80   10   
19:14:59.599 D5F7.0    stand  255  30     -10    0     80   120  
19:15:00.511 D5F7.0    stand  255  30     -10    0     80   0    
19:15:00.511 D5F7.1    stand  255  150    -20    114   80   120  
19:15:01.508 D5F7.0    stand  255  40     -10    45    80   110  
19:15:01.508 D5F7.1    stand  255  150    -40    0     80   114  
19:15:02.526 D5F7.1    stand  255  140    -50    0     80   14   
19:15:02.526 D5F7.0    stand  255  40     0      0     80   111  
19:15:03.530 D5F7.0    stand  255  50     -20    0     80   22   
19:15:03.530 D5F7.1    stand  255  140    -50    0     80   94   
19:15:04.553 D5F7.1    stand  255  140    -50    0     80   0    
19:15:04.553 D5F7.0    stand  255  30     -20    0     80   114  
19:15:05.450 D5F7.1    stand  255  140    -50    0     80   114  
19:15:05.450 D5F7.0    stand  255  30     -20    0     80   114  
19:15:06.444 D5F7.0    stand  255  30     -20    0     80   0    
19:15:06.444 D5F7.1    stand  255  140    -50    0     80   114  
19:15:07.456 D5F7.0    stand  255  30     -20    0     80   114  
19:15:07.456 D5F7.1    stand  255  140    -40    0     80   111  
19:15:08.468 D5F7.0    stand  255  30     -20    0     80   111  
19:15:08.468 D5F7.1    stand  255  140    -40    0     80   111  
19:15:09.482 D5F7.1    stand  255  140    -40    0     80   0    
19:15:09.482 D5F7.0    stand  255  30     -20    94    80   111  
19:15:10.505 D5F7.1    stand  255  140    -40    0     80   111  
19:15:10.505 D5F7.0    stand  255  40     -20    0     80   101  
19:15:11.404 D5F7.1    stand  255  140    -40    0     80   101  
19:15:11.404 D5F7.0    stand  255  40     -20    0     80   101  
19:15:12.419 D5F7.1    stand  255  140    -40    0     80   101  
19:15:12.419 D5F7.0    stand  255  40     -20    0     80   101  
19:15:13.456 D5F7.0    stand  255  40     -20    0     80   0    
19:15:13.456 D5F7.1    stand  255  140    -40    0     80   101  
19:15:14.468 D5F7.1    stand  255  140    -40    0     80   0    
19:15:14.468 D5F7.0    stand  255  40     -20    100   80   101  
19:15:15.377 D5F7.1    stand  255  150    -40    0     80   111  
19:15:15.377 D5F7.0    stand  255  40     -30    98    80   110  
19:15:16.406 D5F7.1    stand  255  150    -50    0     80   111  
19:15:16.406 D5F7.0    stand  255  50     -20    0     80   104  
19:15:17.409 D5F7.0    stand  255  50     -20    0     80   0    
19:15:17.409 D5F7.1    stand  255  140    -40    0     80   92   
19:15:18.412 D5F7.1    stand  255  120    -60    0     80   28   
19:15:18.412 D5F7.0    stand  255  20     -30    0     80   104  
19:15:19.431 D5F7.1    stand  255  120    -60    0     80   104  
19:15:19.431 D5F7.0    stand  255  30     -30    0     80   94   
19:15:20.335 D5F7.1    stand  255  120    -60    0     80   94   
19:15:20.335 D5F7.0    stand  255  30     -20    0     80   98   
19:15:21.351 D5F7.1    stand  255  160    -50    0     80   133  
19:15:21.351 D5F7.0    stand  255  30     -20    76    80   133  
19:15:22.369 D5F7.1    stand  255  160    -30    0     80   130  
19:15:22.369 D5F7.0    stand  255  40     0      91    80   123  
19:15:23.387 D5F7.1    stand  255  160    -10    0     80   120  
19:15:23.387 D5F7.0    stand  255  40     0      82    80   120  
19:15:24.394 D5F7.1    stand  255  170    -30    0     80   133  
19:15:24.394 D5F7.0    stand  255  40     10     82    80   136  
19:15:25.295 D5F7.1    stand  255  170    -20    0     80   133  
19:15:25.295 D5F7.0    stand  255  40     10     50    80   133  
19:15:26.343 D5F7.0    stand  255  20     20     64    80   22   
19:15:26.343 D5F7.1    stand  255  160    -10    0     80   143  
19:15:27.329 D5F7.0    stand  255  10     30     36    80   155  
19:15:27.329 D5F7.1    stand  255  150    10     0     80   141  
19:15:28.332 D5F7.1    stand  255  140    20     0     80   14   
19:15:28.332 D5F7.0    stand  255  0      40     0     80   141  
19:15:29.348 D5F7.0    stand  255  0      40     0     80   0    
19:15:29.348 D5F7.1    stand  255  140    30     0     80   140  
19:15:30.248 D5F7.0    sit    255  0      40     0     80   140  
19:15:30.248 D5F7.1    stand  255  140    30     0     80   140  
19:15:31.270 D5F7.1    stand  255  140    30     0     80   0    
19:15:31.270 D5F7.0    sit    255  0      40     57    80   140  
19:15:32.288 D5F7.0    sit    255  0      40     24    80   0    
19:15:32.288 D5F7.1    stand  255  140    30     0     80   140  
19:15:33.288 D5F7.0    sit    255  0      40     0     80   140  
19:15:33.288 D5F7.1    stand  255  140    30     0     80   140  
19:15:34.314 D5F7.1    stand  255  140    30     0     80   0    
19:15:34.314 D5F7.0    sit    255  0      40     51    80   140  
19:15:35.231 D5F7.1    stand  255  150    0      0     80   155  
19:15:35.231 D5F7.0    sit    255  -10    50     50    80   167  
19:15:36.229 D5F7.0    sit    255  10     20     59    80   36   
19:15:36.229 D5F7.1    stand  255  160    10     0     80   150  
19:15:37.243 D5F7.0    sit    255  10     20     60    80   150  
19:15:37.243 D5F7.1    stand  255  150    10     0     80   140  
19:15:38.251 D5F7.1    stand  255  150    10     0     80   0    
19:15:38.251 D5F7.0    sit    255  10     10     73    80   140  
19:15:39.268 D5F7.1    stand  255  130    10     0     80   120  
19:15:39.268 D5F7.0    sit    255  10     10     0     80   120  
19:15:40.171 D5F7.0    sit    255  20     20     60    80   14   
19:15:40.171 D5F7.1    stand  255  170    10     0     80   150  
19:15:41.185 D5F7.1    stand  255  140    20     0     80   31   
19:15:41.185 D5F7.0    sit    255  20     30     47    80   120  
19:15:42.197 D5F7.1    stand  255  140    30     0     80   120  
19:15:42.197 D5F7.0    sit    255  10     10     0     80   131  
19:15:43.215 D5F7.1    stand  255  150    30     0     80   141  
19:15:43.215 D5F7.0    sit    255  20     20     105   80   130  
19:15:44.232 D5F7.1    stand  255  150    20     0     80   130  
19:15:44.232 D5F7.0    sit    255  20     20     78    80   130  
19:15:45.133 D5F7.0    sit    255  20     10     79    80   10   
19:15:45.133 D5F7.1    stand  255  150    10     0     80   130  
19:15:46.171 D5F7.1    stand  255  150    0      0     80   10   
19:15:46.171 D5F7.0    sit    255  20     10     0     80   130  
19:15:47.165 D5F7.0    sit    255  10     20     87    80   14   
19:15:47.165 D5F7.1    stand  255  160    10     0     80   150  
19:15:48.177 D5F7.0    stand  255  20     20     90    80   140  
19:15:48.177 D5F7.1    stand  255  140    30     0     80   120  
19:15:49.273 D5F7.1    stand  255  140    30     0     80   0    
19:15:49.273 D5F7.0    stand  255  10     20     68    80   130  
19:15:50.096 D5F7.1    stand  255  150    0      0     80   141  
19:15:50.096 D5F7.0    stand  255  20     10     104   80   130  
19:15:51.111 D5F7.1    stand  255  170    0      0     80   150  
19:15:51.111 D5F7.0    stand  255  30     10     65    80   140  
19:15:52.118 D5F7.1    stand  255  160    -10    0     80   131  
19:15:52.118 D5F7.0    stand  255  30     0      91    80   130  
19:15:53.140 D5F7.1    stand  255  160    -10    0     80   130  
19:15:53.140 D5F7.0    stand  255  30     -20    94    80   130  
19:15:54.068 D5F7.0    stand  255  20     -10    51    80   14   
19:15:54.068 D5F7.1    stand  255  160    0      0     80   140  
19:15:55.080 D5F7.0    walk   255  0      0      59    80   160  
19:15:55.080 D5F7.1    stand  255  160    10     0     80   160  
19:15:56.093 D5F7.1    stand  255  190    -40    0     80   58   
19:15:56.093 D5F7.0    walk   255  -20    20     63    80   218  
19:15:57.106 D5F7.1    stand  255  190    -40    0     80   218  
19:15:57.106 D5F7.0    walk   255  -90    50     0     80   294  
19:15:58.127 D5F7.0    walk   255  -100   60     0     80   14   
19:15:58.127 D5F7.1    stand  255  180    -40    0     80   297  
19:15:59.019 D5F7.0    stand  255  -110   60     0     80   306  
19:15:59.019 D5F7.1    stand  255  180    -40    0     80   306  
19:16:00.032 D5F7.0    stand  255  -110   70     0     80   310  
19:16:00.032 D5F7.1    stand  255  180    -40    0     80   310  
19:16:01.042 D5F7.0    stand  255  -110   70     0     80   310  
19:16:01.042 D5F7.1    stand  255  180    -40    0     80   310  
19:16:02.119 D5F7.0    stand  0    -110   70     0     80   310  
19:16:02.119 D5F7.1    stand  255  180    -40    0     80   310  
19:16:03.137 D5F7.1    stand  255  180    -40    0     80   0    
19:16:04.008 D5F7.1    stand  255  180    -40    0     80   0    
19:16:05.018 D5F7.1    stand  255  180    -40    0     80   0    
19:16:06.033 D5F7.1    stand  255  180    -40    0     80   0    
19:16:07.043 D5F7.1    stand  255  180    -40    0     80   0    
19:16:07.945 D5F7.1    stand  255  180    -40    0     80   0    
19:16:08.959 D5F7.1    stand  255  180    -40    0     80   0    
19:16:10.000 D5F7.1    stand  255  180    -40    0     80   0    
19:16:11.016 D5F7.1    stand  255  180    -40    0     80   0    
19:16:11.912 D5F7.1    stand  255  180    -40    0     80   0    
19:16:12.931 D5F7.1    stand  255  180    -40    0     80   0    
19:16:13.942 D5F7.1    stand  255  180    -40    0     80   0    
19:16:14.953 D5F7.1    stand  255  180    -40    0     80   0    
19:16:15.972 D5F7.1    stand  255  180    -40    0     80   0    
19:16:16.875 D5F7.1    stand  255  180    -40    0     80   0    
19:16:17.885 D5F7.1    stand  255  180    -40    0     80   0    
19:16:18.915 D5F7.1    stand  255  180    -40    0     80   0    
19:16:19.914 D5F7.1    stand  255  180    -40    0     80   0    
19:16:20.926 D5F7.1    stand  255  180    -40    0     80   0    
19:16:21.840 D5F7.1    stand  255  180    -40    0     80   0    
19:16:22.851 D5F7.1    stand  255  180    -40    0     80   0    
19:16:23.873 D5F7.1    stand  255  180    -40    0     80   0    
19:16:24.873 D5F7.1    stand  255  180    -40    0     80   0    
19:16:25.906 D5F7.1    stand  255  180    -40    0     80   0    
19:16:26.798 D5F7.1    stand  255  180    -40    0     80   0    
19:16:27.813 D5F7.1    stand  255  180    -40    0     80   0    
19:16:28.824 D5F7.1    stand  255  180    -40    0     80   0    
19:16:29.834 D5F7.1    stand  255  180    -40    0     80   0    
19:16:30.848 D5F7.1    stand  255  180    -40    0     80   0    
19:16:31.750 D5F7.1    stand  255  180    -40    0     80   0    
19:16:32.766 D5F7.1    stand  255  180    -40    0     80   0    
19:16:33.784 D5F7.1    stand  255  180    -40    0     80   0    
19:16:34.788 D5F7.1    stand  255  180    -40    0     80   0    
19:16:35.813 D5F7.1    stand  255  180    -40    0     80   0    
19:16:36.717 D5F7.1    stand  255  180    -40    0     80   0    
19:16:37.730 D5F7.1    stand  255  180    -40    0     80   0    
19:16:38.741 D5F7.1    stand  255  180    -40    0     80   0    
19:16:39.763 D5F7.1    stand  255  180    -40    0     80   0    
19:16:40.765 D5F7.1    stand  255  180    -40    0     80   0    
19:16:41.675 D5F7.1    stand  255  180    -40    0     80   0    
19:16:42.688 D5F7.1    stand  255  180    -40    0     80   0    
19:16:43.699 D5F7.1    stand  255  180    -40    0     80   0    
19:16:44.713 D5F7.1    stand  255  180    -40    0     80   0    
19:16:45.735 D5F7.1    stand  255  180    -40    0     80   0    
19:16:46.642 D5F7.1    stand  255  180    -40    0     80   0    
19:16:47.674 D5F7.1    stand  255  180    -40    0     80   0    
19:16:48.743 D5F7.1    stand  255  180    -40    0     80   0    
19:16:49.708 D5F7.1    stand  255  180    -40    0     80   0    
19:16:50.598 D5F7.1    stand  255  180    -40    0     80   0    
19:16:51.614 D5F7.1    stand  255  180    -40    0     80   0    
19:16:52.632 D5F7.1    stand  255  180    -40    0     80   0    
19:16:53.640 D5F7.1    stand  255  180    -40    0     80   0    
19:16:54.657 D5F7.1    stand  255  180    -40    0     80   0    
19:16:55.560 D5F7.1    stand  255  180    -40    0     80   0    
19:16:56.572 D5F7.1    stand  255  180    -40    0     80   0    
19:16:57.589 D5F7.1    stand  255  180    -40    0     80   0    
19:16:58.637 D5F7.1    stand  255  180    -40    0     80   0    
19:16:59.527 D5F7.1    stand  255  180    -40    0     80   0    
19:17:00.540 D5F7.1    stand  255  180    -40    0     80   0    
19:17:01.551 D5F7.1    stand  255  180    -40    0     80   0    
19:17:02.570 D5F7.1    stand  255  180    -40    0     80   0    
19:17:03.593 D5F7.1    stand  255  180    -40    0     80   0    
19:17:04.492 D5F7.1    stand  255  180    -40    0     80   0    
19:17:05.499 D5F7.1    stand  255  180    -40    0     80   0    
19:17:06.575 D5F7.88   88     -    -      -      -     -    -    
19:17:07.545 D5F7.88   88     -    -      -      -     -    -    
19:17:08.459 D5F7.88   88     -    -      -      -     -    -    
19:17:34.274 D5F7.88   88     -    -      -      -     -    -    
19:18:06.371 D5F7.88   88     -    -      -      -     -    -    
19:18:37.734 D5F7.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 995 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

## 完整原始记录（按时间排序，data_value 全文不删字段）
```
time     ms             device.tid   event          x      y      z     原始记录
19:02:13 1782867733702  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:02:13 1782867733702  D5F7.9       activity       -      -      -     {"track_id": 9, "heart_rate": -1, "event_since": 1782867733702, "track_count": 0, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "respiratory_rate": -1, "multi_person_duration": 0}
19:02:13 1782867733948  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:02:45 1782867765410  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:03:17 1782867797256  D5F7         deviceStatus   -      -      -     {"statuses": {"offline": 0}}
19:03:17 1782867797258  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782867797258, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
19:03:17 1782867797258  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:03:17 1782867797298  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:03:48 1782867828902  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:20 1782867860744  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782867860744, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 0}
19:04:20 1782867860744  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:04:20 1782867860781  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:35 1782867875965  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782867875965, "event_status": "start", "number_people": 1}
19:04:36 1782867876286  D5F7.0       EnterRoom      -      -      -     {"event": 1, "track_id": 0, "area_type": 4, "event_type": 1, "event_since": 1782867876286, "event_status": "start"}
19:04:36 1782867876288  D5F7.0       track          -60    60     83    {"pose": 4, "event": 1, "area_id": 0, "track_id": 0, "position_x": -60, "position_y": 60, "position_z": 83, "remaining_time": 0, "track_confidence": 80}
19:04:36 1782867876575  D5F7.0       track          -40    40     95    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -40, "position_y": 40, "position_z": 95, "remaining_time": 0, "track_confidence": 80}
19:04:37 1782867877585  D5F7.0       track          -40    40     78    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -40, "position_y": 40, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
19:04:38 1782867878604  D5F7.0       track          -30    30     86    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": -30, "position_y": 30, "position_z": 86, "remaining_time": 0, "track_confidence": 80}
19:04:39 1782867879505  D5F7.0       track          0      20     61    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 0, "position_y": 20, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
19:04:40 1782867880510  D5F7.0       track          30     30     89    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 30, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
19:04:41 1782867881521  D5F7.0       track          60     20     62    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": 20, "position_z": 62, "remaining_time": 0, "track_confidence": 80}
19:04:42 1782867882536  D5F7.0       track          80     0      68    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
19:04:43 1782867883548  D5F7.0       track          70     0      85    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": 0, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
19:04:44 1782867884470  D5F7.0       track          70     -10    58    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": -10, "position_z": 58, "remaining_time": 0, "track_confidence": 80}
19:04:45 1782867885471  D5F7.0       track          70     -10    63    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": -10, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
19:04:46 1782867886479  D5F7.0       track          90     -10    72    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": -10, "position_z": 72, "remaining_time": 0, "track_confidence": 80}
19:04:47 1782867887490  D5F7.0       track          110    -10    61    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 110, "position_y": -10, "position_z": 61, "remaining_time": 0, "track_confidence": 80}
19:04:48 1782867888510  D5F7.0       track          80     -10    63    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": -10, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
19:04:49 1782867889411  D5F7.0       track          90     -20    63    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": -20, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
19:04:50 1782867890432  D5F7.0       track          110    -30    73    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 110, "position_y": -30, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
19:04:51 1782867891440  D5F7.0       track          100    -10    82    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 100, "position_y": -10, "position_z": 82, "remaining_time": 0, "track_confidence": 80}
19:04:52 1782867892457  D5F7.0       track          90     0      47    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": 0, "position_z": 47, "remaining_time": 0, "track_confidence": 80}
19:04:53 1782867893471  D5F7.0       track          130    -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 130, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:54 1782867894383  D5F7.0       track          110    0      71    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 110, "position_y": 0, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
19:04:55 1782867895422  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782867895422, "event_status": "instant", "lie_duration": 0, "walk_distance": 2, "walk_duration": 11, "stand_duration": 4, "multi_person_duration": 0}
19:04:55 1782867895422  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:04:55 1782867895464  D5F7.0       track          30     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:56 1782867896417  D5F7.0       track          30     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:57 1782867897354  D5F7.0       track          30     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:58 1782867898363  D5F7.0       track          20     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:04:59 1782867899373  D5F7.0       track          30     -10    83    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 83, "remaining_time": 0, "track_confidence": 80}
19:05:00 1782867900388  D5F7.0       track          40     -20    57    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 57, "remaining_time": 0, "track_confidence": 80}
19:05:01 1782867901402  D5F7.0       track          40     -20    75    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 75, "remaining_time": 0, "track_confidence": 80}
19:05:02 1782867902306  D5F7.0       track          0      -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 0, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:03 1782867903325  D5F7.0       track          0      -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 0, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:04 1782867904334  D5F7.0       track          10     0      78    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 10, "position_y": 0, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
19:05:05 1782867905352  D5F7.0       track          20     -10    69    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 69, "remaining_time": 0, "track_confidence": 80}
19:05:06 1782867906363  D5F7.0       track          80     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:07 1782867907265  D5F7.0       track          30     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:08 1782867908282  D5F7.0       track          30     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:09 1782867909293  D5F7.0       track          40     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:10 1782867910314  D5F7.0       track          40     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:11 1782867911324  D5F7.0       track          40     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:12 1782867912225  D5F7.0       track          40     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:13 1782867913238  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:14 1782867914281  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:15 1782867915300  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:16 1782867916193  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:17 1782867917208  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:18 1782867918234  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:19 1782867919249  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:20 1782867920247  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:21 1782867921167  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:22 1782867922173  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:23 1782867923184  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:24 1782867924199  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:25 1782867925217  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:26 1782867926119  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:27 1782867927143  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:28 1782867928148  D5F7.0       track          40     -20    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:29 1782867929157  D5F7.0       track          30     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:30 1782867930174  D5F7.0       track          20     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:31 1782867931083  D5F7.0       track          20     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:32 1782867932091  D5F7.0       track          20     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:33 1782867933116  D5F7.0       track          20     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:34 1782867934128  D5F7.0       track          20     -10    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:35 1782867935127  D5F7.0       track          30     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:36 1782867936040  D5F7.0       track          30     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:37 1782867937074  D5F7.0       track          30     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:38 1782867938062  D5F7.0       track          30     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:39 1782867939077  D5F7.0       track          40     -20    60    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 60, "remaining_time": 0, "track_confidence": 80}
19:05:40 1782867940093  D5F7.0       track          40     -30    0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -30, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:40 1782867940994  D5F7.0       track          40     0      0     {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:42 1782867942022  D5F7.0       track          120    -10    73    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 120, "position_y": -10, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
19:05:43 1782867943026  D5F7.0       track          90     -10    69    {"pose": 3, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": -10, "position_z": 69, "remaining_time": 0, "track_confidence": 80}
19:05:44 1782867944040  D5F7.0       track          100    -10    89    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 100, "position_y": -10, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
19:05:45 1782867945046  D5F7.0       track          40     0      0     {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:45 1782867945960  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:46 1782867946976  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:47 1782867947980  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:48 1782867948997  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:50 1782867950015  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:50 1782867950920  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:51 1782867951939  D5F7.0       track          40     0      38    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 38, "remaining_time": 0, "track_confidence": 80}
19:05:52 1782867952945  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:53 1782867953994  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:05:53 1782867953994  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782867953994, "event_status": "instant", "lie_duration": 0, "walk_distance": 1, "walk_duration": 2, "stand_duration": 7, "multi_person_duration": 0}
19:05:54 1782867954044  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:54 1782867954889  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:55 1782867955897  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:56 1782867956908  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:57 1782867957922  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:58 1782867958934  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:05:59 1782867959841  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:00 1782867960862  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:01 1782867961915  D5F7.0       track          90     -10    90    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": -10, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
19:06:02 1782867962821  D5F7.0       track          160    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:03 1782867963837  D5F7.0       track          160    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:04 1782867964845  D5F7.0       track          120    -20    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 120, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:05 1782867965865  D5F7.0       track          70     -10    94    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": -10, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
19:06:06 1782867966873  D5F7.0       track          40     0      0     {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:07 1782867967785  D5F7.0       track          40     0      0     {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:08 1782867968794  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:09 1782867969806  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:10 1782867970818  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:11 1782867971830  D5F7.0       track          40     0      90    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 90, "remaining_time": 0, "track_confidence": 80}
19:06:12 1782867972746  D5F7.0       track          110    -10    91    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 110, "position_y": -10, "position_z": 91, "remaining_time": 0, "track_confidence": 80}
19:06:13 1782867973752  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:14 1782867974766  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:15 1782867975790  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:16 1782867976821  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:17 1782867977699  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:18 1782867978720  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:19 1782867979724  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:20 1782867980740  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:21 1782867981752  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:22 1782867982660  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:23 1782867983682  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:24 1782867984688  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:25 1782867985702  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:26 1782867986712  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:27 1782867987629  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:28 1782867988637  D5F7.0       track          50     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:29 1782867989647  D5F7.0       track          50     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:30 1782867990662  D5F7.0       track          50     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:31 1782867991677  D5F7.0       track          50     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:32 1782867992581  D5F7.0       track          50     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:33 1782867993600  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:34 1782867994609  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:35 1782867995619  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:36 1782867996635  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:37 1782867997540  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:38 1782867998558  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:39 1782867999570  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:40 1782868000582  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:41 1782868001595  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:42 1782868002504  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:43 1782868003515  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:44 1782868004562  D5F7.0       track          40     0      105   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 105, "remaining_time": 0, "track_confidence": 80}
19:06:45 1782868005549  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:46 1782868006551  D5F7.0       track          40     0      89    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
19:06:47 1782868007458  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:48 1782868008474  D5F7.0       track          30     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:49 1782868009444  D5F7.0       track          30     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:50 1782868010459  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:51 1782868011476  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:52 1782868012486  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:53 1782868013541  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868013541, "event_status": "instant", "lie_duration": 0, "walk_distance": 1, "walk_duration": 3, "stand_duration": 57, "multi_person_duration": 0}
19:06:53 1782868013541  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:06:53 1782868013700  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:54 1782868014408  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:55 1782868015417  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:56 1782868016445  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:57 1782868017445  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:58 1782868018456  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:06:59 1782868019366  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:00 1782868020376  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:01 1782868021390  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:02 1782868022406  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:03 1782868023416  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:04 1782868024328  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:05 1782868025344  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:06 1782868026374  D5F7.0       track          40     0      94    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 94, "remaining_time": 0, "track_confidence": 80}
19:07:07 1782868027401  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:08 1782868028293  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:09 1782868029309  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:10 1782868030330  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:11 1782868031340  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:12 1782868032344  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:13 1782868033252  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:14 1782868034268  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:15 1782868035311  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:16 1782868036290  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:17 1782868037302  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:18 1782868038222  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:19 1782868039226  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:20 1782868040238  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:21 1782868041249  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:22 1782868042272  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:23 1782868043170  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:24 1782868044191  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:25 1782868045198  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:26 1782868046214  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:27 1782868047226  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:28 1782868048132  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:29 1782868049146  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:30 1782868050158  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:31 1782868051172  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:31 1782868051801  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868051801, "event_status": "start", "number_people": 2}
19:07:31 1782868051838  D5F7.1       EnterRoom      -      -      -     {"event": 1, "track_id": 1, "area_type": 4, "event_type": 1, "event_since": 1782868051838, "event_status": "start"}
19:07:31 1782868051840  D5F7.1       track          -90    70     115   {"pose": 4, "event": 1, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 70, "position_z": 115, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:31 1782868051840  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:32 1782868052106  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:32 1782868052106  D5F7.1       track          -90    70     99    {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 70, "position_z": 99, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:33 1782868053113  D5F7.1       track          -90    70     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:33 1782868053113  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:34 1782868054141  D5F7.1       track          -90    80     111   {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 111, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:34 1782868054141  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:35 1782868055142  D5F7.0       track          40     0      104   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 104, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:35 1782868055142  D5F7.1       track          -90    80     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:36 1782868056158  D5F7.1       track          -100   80     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -100, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:36 1782868056158  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:37 1782868057063  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:37 1782868057063  D5F7.1       track          -90    80     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:38 1782868058075  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:38 1782868058075  D5F7.1       track          -90    80     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:39 1782868059092  D5F7.1       track          -90    80     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:39 1782868059092  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:40 1782868060104  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:40 1782868060104  D5F7.1       track          -90    80     0     {"pose": 4, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:41 1782868061145  D5F7.1       ExitRoom       -      -      -     {"event": 2, "track_id": 1, "area_type": 4, "event_type": 1, "event_since": 1782868061145, "event_status": "start"}
19:07:41 1782868061184  D5F7.1       track          -90    80     0     {"pose": 4, "event": 2, "area_id": 0, "track_id": 1, "position_x": -90, "position_y": 80, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:41 1782868061184  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:07:42 1782868062024  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868062024, "event_status": "start", "number_people": 1}
19:07:42 1782868062067  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:43 1782868063064  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:44 1782868064077  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:45 1782868065095  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:46 1782868066009  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:47 1782868067009  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:48 1782868068018  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:49 1782868069034  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:50 1782868070039  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:50 1782868070955  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:51 1782868071970  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:53 1782868073013  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868073013, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 51, "multi_person_duration": 9}
19:07:53 1782868073013  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:07:53 1782868073049  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:54 1782868074018  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:54 1782868074918  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:55 1782868075930  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:56 1782868076943  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:57 1782868077960  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:58 1782868078967  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:07:59 1782868079876  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:00 1782868080898  D5F7.0       track          50     0      71    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 71, "remaining_time": 0, "track_confidence": 80}
19:08:01 1782868081906  D5F7.0       track          10     -40    93    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 10, "position_y": -40, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
19:08:02 1782868082951  D5F7.0       track          20     -30    99    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 20, "position_y": -30, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
19:08:03 1782868083847  D5F7.0       track          40     -20    68    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 68, "remaining_time": 0, "track_confidence": 80}
19:08:04 1782868084869  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:05 1782868085879  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:06 1782868086888  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:07 1782868087902  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:08 1782868088802  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:09 1782868089828  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:10 1782868090831  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:11 1782868091846  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:12 1782868092865  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:13 1782868093769  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:14 1782868094779  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:15 1782868095794  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:16 1782868096805  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:17 1782868097818  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:18 1782868098725  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:19 1782868099739  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:20 1782868100761  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:21 1782868101765  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:22 1782868102779  D5F7.0       track          70     0      101   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": 0, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
19:08:23 1782868103689  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:24 1782868104707  D5F7.0       track          50     0      78    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 78, "remaining_time": 0, "track_confidence": 80}
19:08:25 1782868105711  D5F7.0       track          70     0      89    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": 0, "position_z": 89, "remaining_time": 0, "track_confidence": 80}
19:08:26 1782868106725  D5F7.0       track          90     0      77    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": 0, "position_z": 77, "remaining_time": 0, "track_confidence": 80}
19:08:27 1782868107742  D5F7.0       track          70     0      111   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": 0, "position_z": 111, "remaining_time": 0, "track_confidence": 80}
19:08:28 1782868108647  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:29 1782868109657  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:30 1782868110670  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:31 1782868111685  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:32 1782868112709  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:33 1782868113609  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:34 1782868114618  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:35 1782868115631  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:36 1782868116655  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:37 1782868117661  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:38 1782868118570  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:39 1782868119579  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:40 1782868120597  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:41 1782868121629  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:42 1782868122541  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:43 1782868123547  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:44 1782868124559  D5F7.0       track          40     0      98    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
19:08:45 1782868125576  D5F7.0       track          50     -10    99    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 99, "remaining_time": 0, "track_confidence": 80}
19:08:46 1782868126585  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:47 1782868127494  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:48 1782868128517  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:49 1782868129522  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:50 1782868130544  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:51 1782868131551  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:52 1782868132458  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:53 1782868133547  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:08:53 1782868133547  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868133547, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "multi_person_duration": 0}
19:08:53 1782868133713  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:54 1782868134439  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:55 1782868135453  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:56 1782868136466  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:57 1782868137490  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:58 1782868138493  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:08:59 1782868139398  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:00 1782868140412  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:01 1782868141435  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:02 1782868142442  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:03 1782868143451  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:04 1782868144362  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:05 1782868145378  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:06 1782868146385  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:07 1782868147396  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:08 1782868148409  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:09 1782868149316  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:10 1782868150331  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:11 1782868151347  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:12 1782868152356  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:13 1782868153374  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:14 1782868154276  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:15 1782868155294  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:16 1782868156327  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:17 1782868157324  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:18 1782868158334  D5F7.0       track          50     -10    101   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
19:09:19 1782868159241  D5F7.0       track          100    0      115   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 100, "position_y": 0, "position_z": 115, "remaining_time": 0, "track_confidence": 80}
19:09:20 1782868160253  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:21 1782868161266  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:22 1782868162282  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:23 1782868163305  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:24 1782868164204  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:25 1782868165211  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:26 1782868166231  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:27 1782868167237  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:28 1782868168249  D5F7.0       track          40     0      98    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
19:09:29 1782868169161  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:30 1782868170149  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:31 1782868171168  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:32 1782868172184  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:33 1782868173188  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:34 1782868174201  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:35 1782868175111  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:36 1782868176127  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:37 1782868177142  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:38 1782868178152  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:39 1782868179166  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:40 1782868180069  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:41 1782868181093  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:42 1782868182095  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:43 1782868183129  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:44 1782868184130  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:45 1782868185037  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:46 1782868186091  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:47 1782868187096  D5F7.0       track          50     0      101   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
19:09:48 1782868188105  D5F7.0       track          140    0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 140, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:49 1782868189007  D5F7.0       track          130    0      109   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 130, "position_y": 0, "position_z": 109, "remaining_time": 0, "track_confidence": 80}
19:09:50 1782868190025  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:51 1782868191051  D5F7.0       track          40     0      101   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 101, "remaining_time": 0, "track_confidence": 80}
19:09:52 1782868192075  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:09:52 1782868192075  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868192075, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "multi_person_duration": 0}
19:09:52 1782868192120  D5F7.0       track          40     0      102   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 102, "remaining_time": 0, "track_confidence": 80}
19:09:52 1782868192971  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:53 1782868193980  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:54 1782868194993  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:56 1782868196006  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:57 1782868197026  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:57 1782868197928  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:58 1782868198941  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:09:59 1782868199955  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:00 1782868200982  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:01 1782868201979  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:02 1782868202885  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:03 1782868203904  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:04 1782868204911  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:05 1782868205927  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:06 1782868206941  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:07 1782868207849  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:08 1782868208863  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:09 1782868209881  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:10 1782868210884  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:11 1782868211910  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:12 1782868212810  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:13 1782868213819  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:14 1782868214834  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:15 1782868215847  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:16 1782868216859  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:17 1782868217768  D5F7.0       track          40     0      98    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
19:10:18 1782868218784  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:19 1782868219797  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:20 1782868220813  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:21 1782868221830  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:22 1782868222726  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:23 1782868223737  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:24 1782868224758  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:25 1782868225773  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:26 1782868226789  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:27 1782868227685  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:28 1782868228705  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:29 1782868229718  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:30 1782868230725  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:31 1782868231738  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:32 1782868232649  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:33 1782868233710  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:34 1782868234723  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:35 1782868235620  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:36 1782868236639  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:37 1782868237650  D5F7.0       track          40     0      98    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 98, "remaining_time": 0, "track_confidence": 80}
19:10:38 1782868238667  D5F7.0       track          80     -20    81    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": -20, "position_z": 81, "remaining_time": 0, "track_confidence": 80}
19:10:39 1782868239678  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:40 1782868240588  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:41 1782868241597  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:42 1782868242607  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:43 1782868243630  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:44 1782868244633  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:45 1782868245543  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:46 1782868246557  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:47 1782868247579  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:48 1782868248598  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:49 1782868249607  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:50 1782868250500  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:51 1782868251518  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:52 1782868252564  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:10:52 1782868252564  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868252564, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 60, "multi_person_duration": 0}
19:10:52 1782868252610  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:53 1782868253563  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:54 1782868254482  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:55 1782868255492  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:56 1782868256506  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:57 1782868257510  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:58 1782868258528  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:10:59 1782868259436  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:00 1782868260449  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:01 1782868261461  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:02 1782868262471  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:03 1782868263482  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:04 1782868264393  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:05 1782868265403  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:06 1782868266428  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:07 1782868267452  D5F7.0       track          60     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:08 1782868268441  D5F7.0       track          70     -10    65    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 70, "position_y": -10, "position_z": 65, "remaining_time": 0, "track_confidence": 80}
19:11:09 1782868269352  D5F7.0       track          90     -20    63    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": -20, "position_z": 63, "remaining_time": 0, "track_confidence": 80}
19:11:10 1782868270370  D5F7.0       track          40     -30    73    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -30, "position_z": 73, "remaining_time": 0, "track_confidence": 80}
19:11:11 1782868271382  D5F7.0       track          30     -50    88    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 30, "position_y": -50, "position_z": 88, "remaining_time": 0, "track_confidence": 80}
19:11:12 1782868272400  D5F7.0       track          90     -10    86    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 90, "position_y": -10, "position_z": 86, "remaining_time": 0, "track_confidence": 80}
19:11:13 1782868273409  D5F7.0       track          110    -20    93    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 110, "position_y": -20, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
19:11:14 1782868274317  D5F7.0       track          110    0      66    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 110, "position_y": 0, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
19:11:15 1782868275325  D5F7.0       track          80     0      93    {"pose": 1, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
19:11:16 1782868276337  D5F7.0       track          60     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:17 1782868277350  D5F7.0       track          60     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:18 1782868278377  D5F7.0       track          100    0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 100, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:19 1782868279270  D5F7.0       track          100    0      112   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 100, "position_y": 0, "position_z": 112, "remaining_time": 0, "track_confidence": 80}
19:11:20 1782868280286  D5F7.0       track          100    0      66    {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 100, "position_y": 0, "position_z": 66, "remaining_time": 0, "track_confidence": 80}
19:11:21 1782868281301  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:22 1782868282248  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:23 1782868283260  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:24 1782868284271  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:25 1782868285285  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:26 1782868286298  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:27 1782868287205  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:28 1782868288225  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:29 1782868289238  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:30 1782868290245  D5F7.0       track          50     -10    104   {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 104, "remaining_time": 0, "track_confidence": 80}
19:11:31 1782868291266  D5F7.0       track          60     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 60, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:32 1782868292168  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:33 1782868293181  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:34 1782868294195  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:35 1782868295204  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:36 1782868296217  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:37 1782868297128  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:38 1782868298161  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:39 1782868299172  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:40 1782868300192  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:41 1782868301101  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:42 1782868302111  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:43 1782868303130  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:44 1782868304137  D5F7.0       track          50     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:45 1782868305146  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:46 1782868306073  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:47 1782868307072  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:48 1782868308099  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:49 1782868309116  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:50 1782868310121  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:51 1782868311022  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:52 1782868312068  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868312068, "event_status": "instant", "lie_duration": 0, "walk_distance": 1, "walk_duration": 5, "stand_duration": 55, "multi_person_duration": 0}
19:11:52 1782868312068  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:11:52 1782868312117  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:53 1782868313066  D5F7.0       track          80     0      0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 80, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:54 1782868314001  D5F7.0       track          150    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 150, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:55 1782868315002  D5F7.0       track          150    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 150, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:56 1782868316019  D5F7.0       track          150    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 150, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:57 1782868317031  D5F7.0       track          150    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 150, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:11:58 1782868318061  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868318061, "event_status": "start", "number_people": 2}
19:11:58 1782868318099  D5F7.1       track          40     0      78    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 78, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:11:58 1782868318099  D5F7.0       track          140    -20    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:11:58 1782868318959  D5F7.1       track          60     -10    77    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 60, "position_y": -10, "position_z": 77, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:11:58 1782868318959  D5F7.0       track          160    -10    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:11:59 1782868319978  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:11:59 1782868319978  D5F7.1       track          40     0      105   {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 105, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:00 1782868320991  D5F7.1       track          50     -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:00 1782868320991  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:01 1782868321997  D5F7.1       track          50     -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:01 1782868321997  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:03 1782868323012  D5F7.1       track          50     -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:03 1782868323012  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:03 1782868323918  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:03 1782868323918  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:04 1782868324923  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:04 1782868324923  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:05 1782868325940  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:05 1782868325940  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:06 1782868326950  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:06 1782868326950  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:07 1782868327971  D5F7.0       track          160    -40    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:07 1782868327971  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:08 1782868328877  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:08 1782868328877  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:09 1782868329874  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:09 1782868329874  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:10 1782868330874  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:10 1782868330874  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:11 1782868331895  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:11 1782868331895  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:12 1782868332907  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:12 1782868332907  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:13 1782868333919  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:13 1782868333919  D5F7.1       track          40     -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:14 1782868334827  D5F7.1       track          40     -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:14 1782868334827  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:15 1782868335847  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:15 1782868335847  D5F7.1       track          40     -10    97    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": -10, "position_z": 97, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:16 1782868336860  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:16 1782868336860  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:17 1782868337866  D5F7.1       track          40     0      72    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 72, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:17 1782868337866  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:18 1782868338878  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:18 1782868338878  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:19 1782868339786  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:19 1782868339786  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:20 1782868340797  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:20 1782868340797  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:21 1782868341823  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:21 1782868341823  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:22 1782868342828  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:22 1782868342828  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:23 1782868343842  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:23 1782868343842  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:24 1782868344765  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:24 1782868344765  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:25 1782868345791  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:25 1782868345791  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:26 1782868346805  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:26 1782868346805  D5F7.1       track          50     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:27 1782868347825  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:27 1782868347825  D5F7.1       track          50     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:28 1782868348718  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:28 1782868348718  D5F7.1       track          40     0      104   {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 104, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:29 1782868349728  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:29 1782868349728  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:30 1782868350740  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:30 1782868350740  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:31 1782868351757  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:31 1782868351757  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:32 1782868352772  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:32 1782868352772  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:33 1782868353675  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:33 1782868353675  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:34 1782868354689  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:34 1782868354689  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:35 1782868355697  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:35 1782868355697  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:36 1782868356719  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:36 1782868356719  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:37 1782868357732  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:37 1782868357732  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:38 1782868358640  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:38 1782868358640  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:39 1782868359655  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:39 1782868359655  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:40 1782868360679  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:40 1782868360679  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:41 1782868361676  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:41 1782868361676  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:42 1782868362690  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:42 1782868362690  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:43 1782868363604  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:43 1782868363604  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:44 1782868364605  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:44 1782868364605  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:45 1782868365617  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:45 1782868365617  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:46 1782868366634  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:46 1782868366634  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:47 1782868367645  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:47 1782868367645  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:48 1782868368608  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:48 1782868368608  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:49 1782868369577  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:49 1782868369577  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:50 1782868370619  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868370619, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 8, "multi_person_duration": 52}
19:12:50 1782868370619  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:12:50 1782868370665  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:50 1782868370665  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:51 1782868371625  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:51 1782868371625  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:52 1782868372528  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:52 1782868372528  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:53 1782868373536  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:53 1782868373536  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:54 1782868374548  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:54 1782868374548  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:55 1782868375574  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:55 1782868375574  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:56 1782868376587  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:56 1782868376587  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:57 1782868377492  D5F7.0       track          160    -30    0     {"pose": 4, "area_id": 0, "track_id": 0, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:57 1782868377492  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:12:58 1782868378508  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868378508, "event_status": "start", "number_people": 1}
19:12:58 1782868378548  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:12:59 1782868379541  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:00 1782868380544  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:01 1782868381453  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:02 1782868382460  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:03 1782868383471  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:04 1782868384495  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:05 1782868385509  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:06 1782868386420  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:07 1782868387425  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:08 1782868388439  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:09 1782868389532  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:10 1782868390376  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:11 1782868391400  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:12 1782868392403  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:13 1782868393415  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:14 1782868394429  D5F7.1       track          50     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:15 1782868395342  D5F7.1       track          50     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 50, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:16 1782868396362  D5F7.1       track          80     -10    93    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 80, "position_y": -10, "position_z": 93, "remaining_time": 0, "track_confidence": 80}
19:13:17 1782868397371  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:18 1782868398388  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:19 1782868399396  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:20 1782868400303  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:21 1782868401310  D5F7.1       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 40, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:22 1782868402330  D5F7.1       track          100    -10    118   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 100, "position_y": -10, "position_z": 118, "remaining_time": 0, "track_confidence": 80}
19:13:23 1782868403336  D5F7.1       track          170    -10    103   {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 170, "position_y": -10, "position_z": 103, "remaining_time": 0, "track_confidence": 80}
19:13:24 1782868404363  D5F7.1       track          150    0      85    {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 85, "remaining_time": 0, "track_confidence": 80}
19:13:25 1782868405258  D5F7.1       track          110    0      0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 110, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:26 1782868406274  D5F7.1       track          170    0      0     {"pose": 1, "area_id": 255, "track_id": 1, "position_x": 170, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:27 1782868407291  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:28 1782868408314  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:29 1782868409316  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:30 1782868410220  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:31 1782868411239  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:32 1782868412248  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:33 1782868413259  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:34 1782868414278  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:35 1782868415187  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:36 1782868416194  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:37 1782868417207  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:38 1782868418212  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:39 1782868419239  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:40 1782868420134  D5F7.1       track          160    0      74    {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 74, "remaining_time": 0, "track_confidence": 80}
19:13:41 1782868421159  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:42 1782868422168  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:43 1782868423184  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:44 1782868424199  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:45 1782868425101  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:46 1782868426118  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:13:47 1782868427139  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868427139, "event_status": "start", "number_people": 2}
19:13:47 1782868427187  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:47 1782868427187  D5F7.0       track          40     0      71    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 71, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:48 1782868428158  D5F7.0       track          40     0      65    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 65, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:48 1782868428158  D5F7.1       track          150    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:49 1782868429068  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:49 1782868429068  D5F7.0       track          30     0      76    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 76, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:50 1782868430085  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:50 1782868430085  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:51 1782868431143  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:13:51 1782868431143  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868431143, "event_status": "instant", "lie_duration": 0, "walk_distance": 1, "walk_duration": 5, "stand_duration": 24, "multi_person_duration": 11}
19:13:51 1782868431623  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:51 1782868431623  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:52 1782868432131  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:52 1782868432131  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:53 1782868433036  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:53 1782868433036  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:54 1782868434046  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:54 1782868434046  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:55 1782868435064  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:55 1782868435064  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:56 1782868436077  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:56 1782868436077  D5F7.1       track          160    0      0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:57 1782868437096  D5F7.1       track          160    -20    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:57 1782868437096  D5F7.0       track          40     0      80    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 80, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:58 1782868438017  D5F7.0       track          40     -10    63    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 63, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:58 1782868438017  D5F7.1       track          160    -20    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:59 1782868439020  D5F7.1       track          150    -20    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:13:59 1782868439020  D5F7.0       track          30     -10    50    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 50, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:00 1782868440031  D5F7.0       track          40     -10    47    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 47, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:00 1782868440031  D5F7.1       track          160    -20    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:01 1782868441049  D5F7.1       track          160    -10    0     {"pose": 3, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:01 1782868441049  D5F7.0       track          40     -10    55    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 55, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:02 1782868442059  D5F7.1       track          160    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:02 1782868442059  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:02 1782868442956  D5F7.1       track          160    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:02 1782868442956  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:03 1782868443979  D5F7.1       track          160    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:03 1782868443979  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:04 1782868444980  D5F7.0       track          50     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:04 1782868444980  D5F7.1       track          140    -30    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:05 1782868445992  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:05 1782868445992  D5F7.0       track          60     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 60, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:07 1782868447010  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:07 1782868447010  D5F7.0       track          120    -50    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 120, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:07 1782868447919  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:07 1782868447919  D5F7.0       track          160    -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:08 1782868448932  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:08 1782868448932  D5F7.0       track          160    -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:09 1782868449946  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:09 1782868449946  D5F7.0       track          160    -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:10 1782868450961  D5F7.0       track          150    -40    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 150, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:10 1782868450961  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:11 1782868451965  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:11 1782868451965  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:12 1782868452878  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:12 1782868452878  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:13 1782868453887  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:13 1782868453887  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:14 1782868454913  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:14 1782868454913  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:15 1782868455916  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:15 1782868455916  D5F7.0       track          120    -60    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 120, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:16 1782868456930  D5F7.0       track          120    -60    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 120, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:16 1782868456930  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:17 1782868457836  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:17 1782868457836  D5F7.0       track          130    -60    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:18 1782868458849  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:18 1782868458849  D5F7.0       track          130    -60    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:19 1782868459864  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:19 1782868459864  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:20 1782868460872  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:20 1782868460872  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:21 1782868461897  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:21 1782868461897  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:22 1782868462804  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:22 1782868462804  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:23 1782868463808  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:23 1782868463808  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:24 1782868464821  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:24 1782868464821  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:25 1782868465838  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:25 1782868465838  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:26 1782868466846  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:26 1782868466846  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:27 1782868467753  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:27 1782868467753  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:28 1782868468766  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:28 1782868468766  D5F7.0       track          130    -70    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 130, "position_y": -70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:29 1782868469782  D5F7.0       track          100    -40    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 100, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:29 1782868469782  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:30 1782868470795  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:30 1782868470795  D5F7.0       track          160    -50    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 160, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:31 1782868471817  D5F7.0       track          150    -50    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 150, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:31 1782868471817  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:32 1782868472719  D5F7.0       track          150    -50    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 150, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:32 1782868472719  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:33 1782868473736  D5F7.0       track          150    -50    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 150, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:33 1782868473736  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:34 1782868474749  D5F7.0       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:34 1782868474749  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:35 1782868475765  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:35 1782868475765  D5F7.0       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:36 1782868476772  D5F7.0       track          100    -40    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 100, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:36 1782868476772  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:37 1782868477675  D5F7.0       track          100    -30    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 100, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:37 1782868477675  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:38 1782868478692  D5F7.1       track          140    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:38 1782868478692  D5F7.0       track          100    -30    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 100, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:39 1782868479704  D5F7.1       track          160    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:39 1782868479704  D5F7.0       track          70     -20    92    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 70, "position_y": -20, "position_z": 92, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:40 1782868480729  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:40 1782868480729  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:41 1782868481751  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:41 1782868481751  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:42 1782868482636  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:42 1782868482636  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:43 1782868483646  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:43 1782868483646  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:44 1782868484663  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:44 1782868484663  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:45 1782868485672  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:45 1782868485672  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:46 1782868486690  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:46 1782868486690  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:47 1782868487605  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:47 1782868487605  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:48 1782868488619  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:48 1782868488619  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:49 1782868489669  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868489669, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 60}
19:14:49 1782868489669  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:14:49 1782868489706  D5F7.1       track          150    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:49 1782868489706  D5F7.0       track          40     0      83    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 83, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:50 1782868490574  D5F7.1       track          150    -10    82    {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -10, "position_z": 82, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:50 1782868490574  D5F7.0       track          40     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:51 1782868491586  D5F7.1       track          130    -30    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 130, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:51 1782868491586  D5F7.0       track          30     -20    64    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 64, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:52 1782868492603  D5F7.1       track          130    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 130, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:52 1782868492603  D5F7.0       track          10     -30    76    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": -30, "position_z": 76, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:53 1782868493610  D5F7.1       track          130    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 130, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:53 1782868493610  D5F7.0       track          0      -40    71    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": -40, "position_z": 71, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:54 1782868494631  D5F7.1       track          150    -30    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:54 1782868494631  D5F7.0       track          20     -10    56    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 56, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:55 1782868495538  D5F7.1       track          160    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:55 1782868495538  D5F7.0       track          20     -30    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:56 1782868496545  D5F7.1       track          150    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:56 1782868496545  D5F7.0       track          20     -30    63    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": -30, "position_z": 63, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:57 1782868497560  D5F7.0       track          30     -10    75    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 75, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:57 1782868497560  D5F7.1       track          150    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:58 1782868498582  D5F7.0       track          40     -10    80    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 80, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:58 1782868498582  D5F7.1       track          160    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:59 1782868499599  D5F7.1       track          150    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:14:59 1782868499599  D5F7.0       track          30     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:00 1782868500511  D5F7.0       track          30     -10    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:00 1782868500511  D5F7.1       track          150    -20    114   {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -20, "position_z": 114, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:01 1782868501508  D5F7.0       track          40     -10    45    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -10, "position_z": 45, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:01 1782868501508  D5F7.1       track          150    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:02 1782868502526  D5F7.1       track          140    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:02 1782868502526  D5F7.0       track          40     0      0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:03 1782868503530  D5F7.0       track          50     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:03 1782868503530  D5F7.1       track          140    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:04 1782868504553  D5F7.1       track          140    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:04 1782868504553  D5F7.0       track          30     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:05 1782868505450  D5F7.1       track          140    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:05 1782868505450  D5F7.0       track          30     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:06 1782868506444  D5F7.0       track          30     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:06 1782868506444  D5F7.1       track          140    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:07 1782868507456  D5F7.0       track          30     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:07 1782868507456  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:08 1782868508468  D5F7.0       track          30     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:08 1782868508468  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:09 1782868509482  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:09 1782868509482  D5F7.0       track          30     -20    94    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 94, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:10 1782868510505  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:10 1782868510505  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:11 1782868511404  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:11 1782868511404  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:12 1782868512419  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:12 1782868512419  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:13 1782868513456  D5F7.0       track          40     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:13 1782868513456  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:14 1782868514468  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:14 1782868514468  D5F7.0       track          40     -20    100   {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -20, "position_z": 100, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:15 1782868515377  D5F7.1       track          150    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:15 1782868515377  D5F7.0       track          40     -30    98    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": -30, "position_z": 98, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:16 1782868516406  D5F7.1       track          150    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:16 1782868516406  D5F7.0       track          50     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:17 1782868517409  D5F7.0       track          50     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 50, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:17 1782868517409  D5F7.1       track          140    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:18 1782868518412  D5F7.1       track          120    -60    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 120, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:18 1782868518412  D5F7.0       track          20     -30    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:19 1782868519431  D5F7.1       track          120    -60    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 120, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:19 1782868519431  D5F7.0       track          30     -30    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:20 1782868520335  D5F7.1       track          120    -60    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 120, "position_y": -60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:20 1782868520335  D5F7.0       track          30     -20    0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:21 1782868521351  D5F7.1       track          160    -50    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:21 1782868521351  D5F7.0       track          30     -20    76    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 76, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:22 1782868522369  D5F7.1       track          160    -30    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:22 1782868522369  D5F7.0       track          40     0      91    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 91, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:23 1782868523387  D5F7.1       track          160    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:23 1782868523387  D5F7.0       track          40     0      82    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 0, "position_z": 82, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:24 1782868524394  D5F7.1       track          170    -30    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 170, "position_y": -30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:24 1782868524394  D5F7.0       track          40     10     82    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 10, "position_z": 82, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:25 1782868525295  D5F7.1       track          170    -20    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 170, "position_y": -20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:25 1782868525295  D5F7.0       track          40     10     50    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 40, "position_y": 10, "position_z": 50, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:26 1782868526343  D5F7.0       track          20     20     64    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 20, "position_z": 64, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:26 1782868526343  D5F7.1       track          160    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:27 1782868527329  D5F7.0       track          10     30     36    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 30, "position_z": 36, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:27 1782868527329  D5F7.1       track          150    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:28 1782868528332  D5F7.1       track          140    20     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:28 1782868528332  D5F7.0       track          0      40     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:29 1782868529348  D5F7.0       track          0      40     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:29 1782868529348  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:30 1782868530248  D5F7.0       track          0      40     0     {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:30 1782868530248  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:31 1782868531270  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:31 1782868531270  D5F7.0       track          0      40     57    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 57, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:32 1782868532288  D5F7.0       track          0      40     24    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 24, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:32 1782868532288  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:33 1782868533288  D5F7.0       track          0      40     0     {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:33 1782868533288  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:34 1782868534314  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:34 1782868534314  D5F7.0       track          0      40     51    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 40, "position_z": 51, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:35 1782868535231  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:35 1782868535231  D5F7.0       track          -10    50     50    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": -10, "position_y": 50, "position_z": 50, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:36 1782868536229  D5F7.0       track          10     20     59    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 20, "position_z": 59, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:36 1782868536229  D5F7.1       track          160    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:37 1782868537243  D5F7.0       track          10     20     60    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 20, "position_z": 60, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:37 1782868537243  D5F7.1       track          150    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:38 1782868538251  D5F7.1       track          150    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:38 1782868538251  D5F7.0       track          10     10     73    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 10, "position_z": 73, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:39 1782868539268  D5F7.1       track          130    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 130, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:39 1782868539268  D5F7.0       track          10     10     0     {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:40 1782868540171  D5F7.0       track          20     20     60    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 20, "position_z": 60, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:40 1782868540171  D5F7.1       track          170    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 170, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:41 1782868541185  D5F7.1       track          140    20     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:41 1782868541185  D5F7.0       track          20     30     47    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 30, "position_z": 47, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:42 1782868542197  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:42 1782868542197  D5F7.0       track          10     10     0     {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:43 1782868543215  D5F7.1       track          150    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:43 1782868543215  D5F7.0       track          20     20     105   {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 20, "position_z": 105, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:44 1782868544232  D5F7.1       track          150    20     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 20, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:44 1782868544232  D5F7.0       track          20     20     78    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 20, "position_z": 78, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:45 1782868545133  D5F7.0       track          20     10     79    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 10, "position_z": 79, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:45 1782868545133  D5F7.1       track          150    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:46 1782868546171  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:46 1782868546171  D5F7.0       track          20     10     0     {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:47 1782868547165  D5F7.0       track          10     20     87    {"pose": 3, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 20, "position_z": 87, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:47 1782868547165  D5F7.1       track          160    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:48 1782868548177  D5F7.0       track          20     20     90    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 20, "position_z": 90, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:48 1782868548177  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:49 1782868549224  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:15:49 1782868549224  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868549224, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 0, "multi_person_duration": 60}
19:15:49 1782868549273  D5F7.1       track          140    30     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 140, "position_y": 30, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:49 1782868549273  D5F7.0       track          10     20     68    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 10, "position_y": 20, "position_z": 68, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:50 1782868550096  D5F7.1       track          150    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 150, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:50 1782868550096  D5F7.0       track          20     10     104   {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": 10, "position_z": 104, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:51 1782868551111  D5F7.1       track          170    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 170, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:51 1782868551111  D5F7.0       track          30     10     65    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": 10, "position_z": 65, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:52 1782868552118  D5F7.1       track          160    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:52 1782868552118  D5F7.0       track          30     0      91    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": 0, "position_z": 91, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:53 1782868553140  D5F7.1       track          160    -10    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": -10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:53 1782868553140  D5F7.0       track          30     -20    94    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 30, "position_y": -20, "position_z": 94, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:54 1782868554068  D5F7.0       track          20     -10    51    {"pose": 4, "area_id": 255, "track_id": 0, "position_x": 20, "position_y": -10, "position_z": 51, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:54 1782868554068  D5F7.1       track          160    0      0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 0, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:55 1782868555080  D5F7.0       track          0      0      59    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": 0, "position_y": 0, "position_z": 59, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:55 1782868555080  D5F7.1       track          160    10     0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 160, "position_y": 10, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:56 1782868556093  D5F7.1       track          190    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 190, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:56 1782868556093  D5F7.0       track          -20    20     63    {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -20, "position_y": 20, "position_z": 63, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:57 1782868557106  D5F7.1       track          190    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 190, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:57 1782868557106  D5F7.0       track          -90    50     0     {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -90, "position_y": 50, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:58 1782868558127  D5F7.0       track          -100   60     0     {"pose": 1, "area_id": 255, "track_id": 0, "position_x": -100, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:58 1782868558127  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:59 1782868559019  D5F7.0       track          -110   60     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 60, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:15:59 1782868559019  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:00 1782868560032  D5F7.0       track          -110   70     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:00 1782868560032  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:01 1782868561042  D5F7.0       track          -110   70     0     {"pose": 4, "area_id": 255, "track_id": 0, "position_x": -110, "position_y": 70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:01 1782868561042  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:02 1782868562076  D5F7.0       ExitRoom       -      -      -     {"event": 2, "track_id": 0, "area_type": 4, "event_type": 1, "event_since": 1782868562076, "event_status": "start"}
19:16:02 1782868562119  D5F7.0       track          -110   70     0     {"pose": 4, "event": 2, "area_id": 0, "track_id": 0, "position_x": -110, "position_y": 70, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:02 1782868562119  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "track_count": 2, "remaining_time": 0, "track_confidence": 80}
19:16:03 1782868563090  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868563090, "event_status": "start", "number_people": 1}
19:16:03 1782868563137  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:04 1782868564008  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:05 1782868565018  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:06 1782868566033  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:07 1782868567043  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:07 1782868567945  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:08 1782868568959  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:10 1782868570000  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:11 1782868571016  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:11 1782868571912  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:12 1782868572931  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:13 1782868573942  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:14 1782868574953  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:15 1782868575972  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:16 1782868576875  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:17 1782868577885  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:18 1782868578915  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:19 1782868579914  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:20 1782868580926  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:21 1782868581840  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:22 1782868582851  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:23 1782868583873  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:24 1782868584873  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:25 1782868585906  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:26 1782868586798  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:27 1782868587813  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:28 1782868588824  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:29 1782868589834  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:30 1782868590848  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:31 1782868591750  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:32 1782868592766  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:33 1782868593784  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:34 1782868594788  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:35 1782868595813  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:36 1782868596717  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:37 1782868597730  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:38 1782868598741  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:39 1782868599763  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:40 1782868600765  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:41 1782868601675  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:42 1782868602688  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:43 1782868603699  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:44 1782868604713  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:45 1782868605735  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:46 1782868606642  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:47 1782868607674  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:48 1782868608703  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868608703, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 46, "multi_person_duration": 14}
19:16:48 1782868608703  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:16:48 1782868608743  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:49 1782868609708  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:50 1782868610598  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:51 1782868611614  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:52 1782868612632  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:53 1782868613640  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:54 1782868614657  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:55 1782868615560  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:56 1782868616572  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:57 1782868617589  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:58 1782868618637  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:16:59 1782868619527  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:00 1782868620540  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:01 1782868621551  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:02 1782868622570  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:03 1782868623593  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:04 1782868624492  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:05 1782868625499  D5F7.1       track          180    -40    0     {"pose": 4, "area_id": 255, "track_id": 1, "position_x": 180, "position_y": -40, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:06 1782868626536  D5F7         number_people  -      -      -     {"event_type": 3, "event_since": 1782868626536, "event_status": "start", "number_people": 0}
19:17:06 1782868626575  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:07 1782868627545  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:08 1782868628459  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:17:34 1782868654274  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:18:06 1782868686026  D5F7.9       activity       -      -      -     {"track_id": 9, "event_type": 9, "event_since": 1782868686026, "event_status": "instant", "lie_duration": 0, "walk_distance": 0, "walk_duration": 0, "stand_duration": 19, "multi_person_duration": 0}
19:18:06 1782868686026  D5F7.11      heart          -      -      -     {"pose": 0, "track_id": 11, "track_confidence": 80}
19:18:06 1782868686371  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
19:18:37 1782868717734  D5F7.88      track          0      0      0     {"pose": 0, "area_id": 0, "track_id": 88, "position_x": 0, "position_y": 0, "position_z": 0, "remaining_time": 0, "track_confidence": 80}
```

# case-d5f7-0627-17431752 — 每 tick belief 时间线 (room fd00:0:3:111:3:300, TZ America/Denver)

dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。
**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。
于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。

```
time     dev.tid  lid           pose    z    bed      event              src  pR   top        nr  still SFall SBed  SOpen SBliR SEmpt SLeft
17:43:00 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:00 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:01 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:02 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:03 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:04 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:05 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:06 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:07 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:08 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:09 D523.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:09 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:10 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:10 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:11 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:12 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:13 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:14 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:15 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:16 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:17 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:18 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:19 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:20 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:21 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:22 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:23 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:24 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:25 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:26 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:27 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:28 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:29 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:30 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:31 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:32 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:32 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:33 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:34 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:35 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:36 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:37 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:38 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:39 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:40 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:41 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:42 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:42 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:43 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:44 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:45 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:46 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:47 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:48 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:49 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:50 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:51 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:52 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:53 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:54 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:55 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:56 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:57 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:58 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:43:59 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:00 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:01 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:02 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:03 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:03 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:04 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:05 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:06 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:07 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:08 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:09 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:10 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:11 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:12 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:13 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:14 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:14 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:15 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:16 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:17 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:18 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:19 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:20 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:21 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:22 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:23 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:24 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:25 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:25 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:26 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:27 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:28 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:29 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:30 -.-      -             -       -    NoReport (no frame, held)   room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:31 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:32 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:33 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:34 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:35 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:35 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:35 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:36 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:37 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:38 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:39 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:40 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:41 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:42 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:43 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:44 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:45 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:45 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:46 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:47 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:48 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:49 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:50 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:51 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:52 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:53 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:54 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:55 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:56 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:57 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:58 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:44:59 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:00 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:01 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:02 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:03 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:04 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:05 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:06 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:07 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:07 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:08 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:09 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:10 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:11 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:12 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:13 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:14 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:15 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:16 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:17 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:17 D523.0   -             stand   88   NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:18 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:19 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:20 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:21 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:22 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:22 D5F7.E   -             -       0    NoReport np=1               room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:23 D5F7.E   -             -       0    NoReport EnterRoom(rdr)     room -    Empty      0   0     0.00  0.00  0.00  0.00  0.00  0.00
17:45:23 D5F7.0   D5F704523166  stand   83   NoReport stand              trk  1.00 Empty      1   0     0.00  0.02  0.26  0.00  0.69  0.03
17:45:23 D5F7.0   D5F704523166  stand   72   NoReport stand              trk  1.00 OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
17:45:23 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.52  0.00  0.40  0.01
17:45:24 D5F7.0   D5F704523166  walk    90   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.70  0.00  0.18  0.02
17:45:24 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.70  0.00  0.18  0.02
17:45:25 D5F7.0   D5F704523166  walk    80   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.80  0.00  0.07  0.02
17:45:25 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.80  0.00  0.07  0.02
17:45:26 D5F7.0   D5F704523166  walk    70   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.83  0.00  0.03  0.02
17:45:26 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.83  0.00  0.03  0.02
17:45:27 D5F7.0   D5F704523166  walk    57   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.84  0.00  0.02  0.02
17:45:27 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.84  0.00  0.02  0.02
17:45:28 D5F7.0   D5F704523166  walk    68   NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
17:45:28 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
17:45:29 D5F7.0   D5F704523166  walk    0    NoReport walk               trk  1.00 OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
17:45:29 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.02  0.85  0.00  0.01  0.02
17:45:30 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.00  0.04  0.70  0.00  0.02  0.03
17:45:30 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.00  0.04  0.70  0.00  0.02  0.03
17:45:31 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.04  0.62  0.01  0.03  0.03
17:45:31 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.62  0.01  0.03  0.03
17:45:32 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.04  0.56  0.01  0.03  0.03
17:45:32 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.56  0.01  0.03  0.03
17:45:33 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.04  0.51  0.01  0.04  0.03
17:45:33 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.51  0.01  0.04  0.03
17:45:34 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.04  0.48  0.02  0.04  0.03
17:45:34 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.48  0.02  0.04  0.03
17:45:35 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.04  0.45  0.02  0.03  0.03
17:45:35 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.45  0.02  0.03  0.03
17:45:36 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.04  0.43  0.02  0.03  0.03
17:45:36 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.04  0.43  0.02  0.03  0.03
17:45:37 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.03  0.41  0.02  0.03  0.03
17:45:37 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.03  0.41  0.02  0.03  0.03
17:45:38 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  1   0     0.01  0.03  0.40  0.02  0.03  0.02
17:45:38 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   0     0.01  0.03  0.40  0.02  0.03  0.02
17:45:39 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.39  0.02  0.03  0.02
17:45:39 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.39  0.02  0.03  0.02
17:45:40 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.38  0.02  0.03  0.02
17:45:40 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.38  0.02  0.03  0.02
17:45:41 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.37  0.02  0.03  0.02
17:45:41 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.37  0.02  0.03  0.02
17:45:42 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.36  0.02  0.03  0.02
17:45:42 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.36  0.02  0.03  0.02
17:45:43 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.36  0.02  0.03  0.02
17:45:43 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.36  0.02  0.03  0.02
17:45:44 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.36  0.02  0.03  0.02
17:45:44 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.36  0.02  0.03  0.02
17:45:45 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:45 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:46 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:46 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:47 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:47 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:48 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:48 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:49 09E7.88  -             88      -    NoReport no-target(88)      room -    Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:49 D5F7.0   D5F704523166  sit     70   NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:49 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.35  0.02  0.03  0.02
17:45:50 D5F7.0   D5F704523166  sit     75   NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:45:50 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:45:51 D5F7.0   D5F704523166  sit     60   NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:45:51 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:45:52 D5F7.0   D5F704523166  sit     79   NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.21  0.03  0.03  0.03
17:45:52 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.21  0.03  0.03  0.03
17:45:53 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.25  0.03  0.03  0.02
17:45:53 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.25  0.03  0.03  0.02
17:45:54 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.28  0.03  0.03  0.02
17:45:54 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.28  0.03  0.03  0.02
17:45:55 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.02  0.30  0.02  0.02  0.02
17:45:55 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.02  0.30  0.02  0.02  0.02
17:45:56 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.31  0.02  0.02  0.02
17:45:56 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.31  0.02  0.02  0.02
17:45:57 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.32  0.02  0.02  0.02
17:45:57 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.32  0.02  0.02  0.02
17:45:58 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.32  0.02  0.02  0.02
17:45:58 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.32  0.02  0.02  0.02
17:45:59 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:45:59 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:00 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:00 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:01 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:01 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:02 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:02 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.33  0.02  0.02  0.02
17:46:03 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:03 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:03 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:04 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:04 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:05 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:05 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:06 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:07 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:07 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:08 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:09 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:09 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:10 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:10 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:11 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:11 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:12 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.02  0.02
17:46:12 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:13 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:13 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:14 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:14 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:15 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:15 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:16 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:16 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:17 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:17 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:18 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:18 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:19 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:19 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:20 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:20 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:20 09E7.88  -             88      -    NoReport no-target(88)      room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:21 D523.0   -             stand   0    NoReport stand              room -    Sit        1   0     0.01  0.03  0.34  0.02  0.03  0.02
17:46:21 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   28    0.01  0.03  0.34  0.02  0.03  0.02
17:46:22 D523.0   -             stand   0    NoReport stand              room -    Sit        1   28    0.01  0.03  0.34  0.02  0.03  0.02
17:46:22 D5F7.0   D5F704523166  sit     55   NoReport sit                trk  1.00 Sit        1   29    0.01  0.03  0.34  0.02  0.03  0.02
17:46:23 D523.0   -             stand   0    NoReport stand              room -    Sit        1   29    0.01  0.03  0.34  0.02  0.03  0.02
17:46:23 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   30    0.01  0.03  0.34  0.02  0.03  0.02
17:46:24 D523.0   -             stand   0    NoReport stand              room -    Sit        1   30    0.01  0.03  0.34  0.02  0.03  0.02
17:46:24 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   31    0.01  0.03  0.34  0.02  0.03  0.02
17:46:25 D523.0   -             stand   0    NoReport stand              room -    Sit        1   31    0.01  0.03  0.34  0.02  0.03  0.02
17:46:25 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   32    0.01  0.03  0.34  0.02  0.03  0.02
17:46:26 D523.0   -             stand   0    NoReport stand              room -    Sit        1   32    0.01  0.03  0.34  0.02  0.03  0.02
17:46:26 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        1   33    0.01  0.03  0.34  0.02  0.03  0.02
17:46:27 D523.0   -             stand   0    NoReport stand              room -    Sit        1   33    0.01  0.03  0.34  0.02  0.03  0.02
17:46:27 D5F7.E   -             -       0    NoReport np=2               room -    Sit        1   33    0.01  0.03  0.34  0.02  0.03  0.02
17:46:27 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   34    0.00  0.02  0.38  0.01  0.01  0.01
17:46:27 D5F7.1   D5F714627904  stand   103  NoReport stand              trk  0.50 Sit        2   34    0.00  0.02  0.26  0.00  0.69  0.03
17:46:28 D523.0   -             stand   0    NoReport stand              room -    Sit        2   34    0.00  0.02  0.38  0.01  0.01  0.01
17:46:28 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   35    0.00  0.01  0.38  0.01  0.01  0.01
17:46:28 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.51 Sit        2   35    0.00  0.01  0.81  0.00  0.16  0.01
17:46:29 D523.0   -             stand   0    NoReport stand              room -    Sit        2   35    0.00  0.01  0.38  0.01  0.01  0.01
17:46:29 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   36    0.00  0.01  0.38  0.01  0.01  0.01
17:46:29 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.52 Sit        2   36    0.00  0.01  0.95  0.00  0.02  0.01
17:46:29 D523.0   -             stand   0    NoReport stand              room -    Sit        2   36    0.00  0.01  0.38  0.01  0.01  0.01
17:46:30 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.53 Sit        2   37    0.00  0.01  0.97  0.00  0.00  0.01
17:46:30 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   37    0.00  0.01  0.38  0.01  0.00  0.01
17:46:30 D523.0   -             stand   0    NoReport stand              room -    Sit        2   37    0.00  0.01  0.38  0.01  0.00  0.01
17:46:31 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   38    0.00  0.01  0.37  0.01  0.00  0.01
17:46:31 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.54 Sit        2   38    0.00  0.01  0.97  0.00  0.00  0.01
17:46:31 D523.0   -             stand   0    NoReport stand              room -    Sit        2   38    0.00  0.01  0.37  0.01  0.00  0.01
17:46:32 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   39    0.00  0.01  0.97  0.00  0.00  0.01
17:46:32 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   39    0.00  0.01  0.37  0.01  0.00  0.01
17:46:32 D523.0   -             stand   0    NoReport stand              room -    Sit        2   39    0.00  0.01  0.37  0.01  0.00  0.01
17:46:33 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   40    0.00  0.01  0.37  0.01  0.00  0.01
17:46:33 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   40    0.00  0.01  0.97  0.00  0.00  0.01
17:46:33 D523.0   -             stand   0    NoReport stand              room -    Sit        2   40    0.00  0.01  0.37  0.01  0.00  0.01
17:46:34 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   41    0.00  0.01  0.97  0.00  0.00  0.01
17:46:34 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   41    0.00  0.01  0.37  0.01  0.00  0.01
17:46:34 D523.0   -             stand   90   NoReport stand              room -    Sit        2   41    0.00  0.01  0.37  0.01  0.00  0.01
17:46:35 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   42    0.00  0.01  0.37  0.01  0.00  0.01
17:46:35 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   42    0.00  0.01  0.97  0.00  0.00  0.01
17:46:36 D523.0   -             stand   0    NoReport stand              room -    Sit        2   42    0.00  0.01  0.37  0.01  0.00  0.01
17:46:36 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   43    0.00  0.01  0.37  0.01  0.00  0.01
17:46:36 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   43    0.00  0.01  0.97  0.00  0.00  0.01
17:46:36 D523.0   -             stand   0    NoReport stand              room -    Sit        2   43    0.00  0.01  0.37  0.01  0.00  0.01
17:46:37 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   44    0.00  0.01  0.37  0.01  0.00  0.01
17:46:37 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   44    0.00  0.01  0.97  0.00  0.00  0.01
17:46:38 D523.0   -             stand   0    NoReport stand              room -    Sit        2   44    0.00  0.01  0.37  0.01  0.00  0.01
17:46:38 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   45    0.00  0.01  0.97  0.00  0.00  0.01
17:46:38 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   45    0.00  0.01  0.37  0.01  0.00  0.01
17:46:39 D523.0   -             stand   0    NoReport stand              room -    Sit        2   45    0.00  0.01  0.37  0.01  0.00  0.01
17:46:39 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   46    0.00  0.01  0.97  0.00  0.00  0.01
17:46:39 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   46    0.00  0.01  0.37  0.01  0.00  0.01
17:46:39 D523.0   -             stand   0    NoReport stand              room -    Sit        2   46    0.00  0.01  0.37  0.01  0.00  0.01
17:46:40 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   47    0.00  0.01  0.97  0.00  0.00  0.01
17:46:40 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   47    0.00  0.01  0.37  0.01  0.00  0.01
17:46:41 D523.0   -             stand   0    NoReport stand              room -    Sit        2   47    0.00  0.01  0.37  0.01  0.00  0.01
17:46:41 D5F7.0   D5F704523166  sit     65   NoReport sit                trk  1.00 Sit        2   48    0.00  0.01  0.37  0.01  0.00  0.01
17:46:41 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   48    0.00  0.01  0.97  0.00  0.00  0.01
17:46:41 D523.0   -             stand   0    NoReport stand              room -    Sit        2   48    0.00  0.01  0.37  0.01  0.00  0.01
17:46:42 D5F7.0   D5F704523166  sit     77   NoReport sit                trk  1.00 Sit        2   49    0.00  0.01  0.37  0.01  0.00  0.01
17:46:42 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   49    0.00  0.01  0.97  0.00  0.00  0.01
17:46:42 D523.0   -             stand   0    NoReport stand              room -    Sit        2   49    0.00  0.01  0.37  0.01  0.00  0.01
17:46:43 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   50    0.00  0.01  0.37  0.01  0.00  0.01
17:46:43 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   50    0.00  0.01  0.97  0.00  0.00  0.01
17:46:43 D523.0   -             stand   0    NoReport stand              room -    Sit        2   50    0.00  0.01  0.37  0.01  0.00  0.01
17:46:44 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   51    0.00  0.01  0.37  0.01  0.00  0.01
17:46:44 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   51    0.00  0.01  0.97  0.00  0.00  0.01
17:46:44 D523.0   -             stand   0    NoReport stand              room -    Sit        2   51    0.00  0.01  0.37  0.01  0.00  0.01
17:46:45 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   52    0.00  0.01  0.37  0.01  0.00  0.01
17:46:45 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   52    0.00  0.01  0.97  0.00  0.00  0.01
17:46:45 D523.0   -             stand   0    NoReport stand              room -    Sit        2   52    0.00  0.01  0.37  0.01  0.00  0.01
17:46:46 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   53    0.00  0.01  0.37  0.01  0.00  0.01
17:46:46 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   53    0.00  0.01  0.97  0.00  0.00  0.01
17:46:46 D523.0   -             stand   0    NoReport stand              room -    Sit        2   53    0.00  0.01  0.37  0.01  0.00  0.01
17:46:47 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   54    0.00  0.01  0.37  0.01  0.00  0.01
17:46:47 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   54    0.00  0.01  0.97  0.00  0.00  0.01
17:46:47 D523.0   -             stand   0    NoReport stand              room -    Sit        2   54    0.00  0.01  0.37  0.01  0.00  0.01
17:46:48 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   55    0.00  0.01  0.37  0.01  0.00  0.01
17:46:48 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   55    0.00  0.01  0.97  0.00  0.00  0.01
17:46:48 D523.0   -             stand   0    NoReport stand              room -    Sit        2   55    0.00  0.01  0.37  0.01  0.00  0.01
17:46:49 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   56    0.00  0.01  0.97  0.00  0.00  0.01
17:46:49 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   56    0.00  0.01  0.37  0.01  0.00  0.01
17:46:49 D523.0   -             stand   0    NoReport stand              room -    Sit        2   56    0.00  0.01  0.37  0.01  0.00  0.01
17:46:50 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   57    0.00  0.01  0.37  0.01  0.00  0.01
17:46:50 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   57    0.00  0.01  0.97  0.00  0.00  0.01
17:46:50 D523.0   -             stand   0    NoReport stand              room -    Sit        2   57    0.00  0.01  0.37  0.01  0.00  0.01
17:46:51 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   58    0.00  0.01  0.37  0.01  0.00  0.01
17:46:51 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   58    0.00  0.01  0.97  0.00  0.00  0.01
17:46:51 D523.0   -             stand   0    NoReport stand              room -    Sit        2   58    0.00  0.01  0.37  0.01  0.00  0.01
17:46:52 09E7.88  -             88      -    NoReport no-target(88)      room -    Sit        2   58    0.00  0.01  0.37  0.01  0.00  0.01
17:46:52 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   59    0.00  0.01  0.97  0.00  0.00  0.01
17:46:52 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   59    0.00  0.01  0.37  0.01  0.00  0.01
17:46:52 D523.0   -             stand   0    NoReport stand              room -    Sit        2   59    0.00  0.01  0.37  0.01  0.00  0.01
17:46:53 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   60    0.00  0.01  0.97  0.00  0.00  0.01
17:46:53 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   60    0.00  0.01  0.37  0.01  0.00  0.01
17:46:53 D523.0   -             stand   0    NoReport stand              room -    Sit        2   60    0.00  0.01  0.37  0.01  0.00  0.01
17:46:54 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   61    0.00  0.01  0.37  0.01  0.00  0.01
17:46:54 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   61    0.00  0.01  0.97  0.00  0.00  0.01
17:46:54 D523.0   -             stand   0    NoReport stand              room -    Sit        2   61    0.00  0.01  0.37  0.01  0.00  0.01
17:46:55 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   62    0.00  0.01  0.37  0.01  0.00  0.01
17:46:55 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   62    0.00  0.01  0.97  0.00  0.00  0.01
17:46:55 D523.0   -             stand   0    NoReport stand              room -    Sit        2   62    0.00  0.01  0.37  0.01  0.00  0.01
17:46:56 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   63    0.00  0.01  0.37  0.01  0.00  0.01
17:46:56 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   63    0.00  0.01  0.97  0.00  0.00  0.01
17:46:56 D523.0   -             stand   0    NoReport stand              room -    Sit        2   63    0.00  0.01  0.37  0.01  0.00  0.01
17:46:57 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   64    0.00  0.01  0.37  0.01  0.00  0.01
17:46:57 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   64    0.00  0.01  0.97  0.00  0.00  0.01
17:46:57 D523.0   -             stand   0    NoReport stand              room -    Sit        2   64    0.00  0.01  0.37  0.01  0.00  0.01
17:46:58 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   65    0.00  0.01  0.37  0.01  0.00  0.01
17:46:58 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   65    0.00  0.01  0.97  0.00  0.00  0.01
17:46:58 D523.0   -             stand   0    NoReport stand              room -    Sit        2   65    0.00  0.01  0.37  0.01  0.00  0.01
17:46:59 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   66    0.00  0.01  0.37  0.01  0.00  0.01
17:46:59 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   66    0.00  0.01  0.97  0.00  0.00  0.01
17:46:59 D523.0   -             stand   0    NoReport stand              room -    Sit        2   66    0.00  0.01  0.37  0.01  0.00  0.01
17:47:00 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   67    0.00  0.01  0.37  0.01  0.00  0.01
17:47:00 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   67    0.00  0.01  0.97  0.00  0.00  0.01
17:47:00 D523.0   -             stand   0    NoReport stand              room -    Sit        2   67    0.00  0.01  0.37  0.01  0.00  0.01
17:47:01 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   68    0.00  0.01  0.37  0.01  0.00  0.01
17:47:01 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   68    0.00  0.01  0.97  0.00  0.00  0.01
17:47:01 D523.0   -             stand   0    NoReport stand              room -    Sit        2   68    0.00  0.01  0.37  0.01  0.00  0.01
17:47:02 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   69    0.00  0.01  0.37  0.01  0.00  0.01
17:47:02 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   69    0.00  0.01  0.97  0.00  0.00  0.01
17:47:02 D523.0   -             stand   0    NoReport stand              room -    Sit        2   69    0.00  0.01  0.37  0.01  0.00  0.01
17:47:03 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   70    0.00  0.01  0.37  0.01  0.00  0.01
17:47:03 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   70    0.00  0.01  0.97  0.00  0.00  0.01
17:47:03 D523.0   -             stand   0    NoReport stand              room -    Sit        2   70    0.00  0.01  0.37  0.01  0.00  0.01
17:47:04 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   71    0.00  0.01  0.97  0.00  0.00  0.01
17:47:04 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   71    0.00  0.01  0.37  0.01  0.00  0.01
17:47:04 D523.0   -             stand   0    NoReport stand              room -    Sit        2   71    0.00  0.01  0.37  0.01  0.00  0.01
17:47:05 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   72    0.00  0.01  0.37  0.01  0.00  0.01
17:47:05 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   72    0.00  0.01  0.97  0.00  0.00  0.01
17:47:05 D523.0   -             stand   0    NoReport stand              room -    Sit        2   72    0.00  0.01  0.37  0.01  0.00  0.01
17:47:06 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   73    0.00  0.01  0.37  0.01  0.00  0.01
17:47:06 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   73    0.00  0.01  0.97  0.00  0.00  0.01
17:47:06 D523.0   -             stand   0    NoReport stand              room -    Sit        2   73    0.00  0.01  0.37  0.01  0.00  0.01
17:47:07 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   74    0.00  0.01  0.37  0.01  0.00  0.01
17:47:07 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   74    0.00  0.01  0.97  0.00  0.00  0.01
17:47:07 D523.0   -             stand   0    NoReport stand              room -    Sit        2   74    0.00  0.01  0.37  0.01  0.00  0.01
17:47:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   75    0.00  0.01  0.37  0.01  0.00  0.01
17:47:08 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   75    0.00  0.01  0.97  0.00  0.00  0.01
17:47:08 D523.0   -             stand   0    NoReport stand              room -    Sit        2   75    0.00  0.01  0.37  0.01  0.00  0.01
17:47:09 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   76    0.00  0.01  0.37  0.01  0.00  0.01
17:47:09 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   76    0.00  0.01  0.97  0.00  0.00  0.01
17:47:09 D523.0   -             stand   0    NoReport stand              room -    Sit        2   76    0.00  0.01  0.37  0.01  0.00  0.01
17:47:10 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   77    0.00  0.01  0.37  0.01  0.00  0.01
17:47:10 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   77    0.00  0.01  0.97  0.00  0.00  0.01
17:47:10 D523.0   -             stand   0    NoReport stand              room -    Sit        2   77    0.00  0.01  0.37  0.01  0.00  0.01
17:47:11 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   78    0.00  0.01  0.97  0.00  0.00  0.01
17:47:11 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   78    0.00  0.01  0.37  0.01  0.00  0.01
17:47:11 D523.0   -             stand   0    NoReport stand              room -    Sit        2   78    0.00  0.01  0.37  0.01  0.00  0.01
17:47:12 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   79    0.00  0.01  0.97  0.00  0.00  0.01
17:47:12 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   79    0.00  0.01  0.37  0.01  0.00  0.01
17:47:12 D523.0   -             stand   0    NoReport stand              room -    Sit        2   79    0.00  0.01  0.37  0.01  0.00  0.01
17:47:13 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   80    0.00  0.01  0.37  0.01  0.00  0.01
17:47:13 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   80    0.00  0.01  0.97  0.00  0.00  0.01
17:47:13 D523.0   -             stand   0    NoReport stand              room -    Sit        2   80    0.00  0.01  0.37  0.01  0.00  0.01
17:47:14 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   81    0.00  0.01  0.37  0.01  0.00  0.01
17:47:14 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   81    0.00  0.01  0.97  0.00  0.00  0.01
17:47:14 D523.0   -             stand   0    NoReport stand              room -    Sit        2   81    0.00  0.01  0.37  0.01  0.00  0.01
17:47:15 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   82    0.00  0.01  0.37  0.01  0.00  0.01
17:47:15 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   82    0.00  0.01  0.97  0.00  0.00  0.01
17:47:15 D523.0   -             stand   0    NoReport stand              room -    Sit        2   82    0.00  0.01  0.37  0.01  0.00  0.01
17:47:16 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   83    0.00  0.01  0.37  0.01  0.00  0.01
17:47:16 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   83    0.00  0.01  0.97  0.00  0.00  0.01
17:47:16 D523.0   -             stand   0    NoReport stand              room -    Sit        2   83    0.00  0.01  0.37  0.01  0.00  0.01
17:47:17 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   84    0.00  0.01  0.37  0.01  0.00  0.01
17:47:17 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   84    0.00  0.01  0.97  0.00  0.00  0.01
17:47:17 D523.0   -             stand   0    NoReport stand              room -    Sit        2   84    0.00  0.01  0.37  0.01  0.00  0.01
17:47:18 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   85    0.00  0.01  0.37  0.01  0.00  0.01
17:47:18 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   85    0.00  0.01  0.97  0.00  0.00  0.01
17:47:18 D523.0   -             stand   0    NoReport stand              room -    Sit        2   85    0.00  0.01  0.37  0.01  0.00  0.01
17:47:19 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   86    0.00  0.01  0.37  0.01  0.00  0.01
17:47:19 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   86    0.00  0.01  0.97  0.00  0.00  0.01
17:47:19 D523.0   -             stand   0    NoReport stand              room -    Sit        2   86    0.00  0.01  0.37  0.01  0.00  0.01
17:47:20 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   87    0.00  0.01  0.36  0.01  0.01  0.01
17:47:20 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   87    0.00  0.01  0.96  0.00  0.00  0.01
17:47:20 D523.0   -             stand   0    NoReport stand              room -    Sit        2   87    0.00  0.01  0.36  0.01  0.01  0.01
17:47:21 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   88    0.00  0.01  0.37  0.01  0.01  0.01
17:47:21 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   88    0.00  0.01  0.97  0.00  0.00  0.01
17:47:21 D523.0   -             stand   0    NoReport stand              room -    Sit        2   88    0.00  0.01  0.37  0.01  0.01  0.01
17:47:22 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   89    0.00  0.01  0.37  0.01  0.00  0.01
17:47:22 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   89    0.00  0.01  0.97  0.00  0.00  0.01
17:47:22 D523.0   -             stand   0    NoReport stand              room -    Sit        2   89    0.00  0.01  0.37  0.01  0.00  0.01
17:47:23 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   90    0.00  0.01  0.37  0.01  0.00  0.01
17:47:23 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   90    0.00  0.01  0.97  0.00  0.00  0.01
17:47:23 D523.0   -             stand   0    NoReport stand              room -    Sit        2   90    0.00  0.01  0.37  0.01  0.00  0.01
17:47:24 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   91    0.00  0.01  0.97  0.00  0.00  0.01
17:47:24 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   91    0.00  0.01  0.37  0.01  0.00  0.01
17:47:24 09E7.88  -             88      -    NoReport no-target(88)      room -    Sit        2   91    0.00  0.01  0.37  0.01  0.00  0.01
17:47:24 D523.0   -             stand   0    NoReport stand              room -    Sit        2   91    0.00  0.01  0.37  0.01  0.00  0.01
17:47:25 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   92    0.00  0.01  0.37  0.01  0.00  0.01
17:47:25 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   92    0.00  0.01  0.97  0.00  0.00  0.01
17:47:25 D523.0   -             stand   0    NoReport stand              room -    Sit        2   92    0.00  0.01  0.37  0.01  0.00  0.01
17:47:26 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   93    0.00  0.01  0.37  0.01  0.00  0.01
17:47:26 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   93    0.00  0.01  0.97  0.00  0.00  0.01
17:47:26 D523.0   -             stand   0    NoReport stand              room -    Sit        2   93    0.00  0.01  0.37  0.01  0.00  0.01
17:47:27 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   94    0.00  0.01  0.37  0.01  0.00  0.01
17:47:27 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   94    0.00  0.01  0.97  0.00  0.00  0.01
17:47:27 D523.0   -             stand   0    NoReport stand              room -    Sit        2   94    0.00  0.01  0.37  0.01  0.00  0.01
17:47:28 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   95    0.00  0.01  0.37  0.01  0.00  0.01
17:47:28 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   95    0.00  0.01  0.97  0.00  0.00  0.01
17:47:28 D523.0   -             stand   0    NoReport stand              room -    Sit        2   95    0.00  0.01  0.37  0.01  0.00  0.01
17:47:29 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   96    0.00  0.01  0.37  0.01  0.00  0.01
17:47:29 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   96    0.00  0.01  0.97  0.00  0.00  0.01
17:47:29 D523.0   -             stand   0    NoReport stand              room -    Sit        2   96    0.00  0.01  0.37  0.01  0.00  0.01
17:47:30 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   97    0.00  0.01  0.37  0.01  0.00  0.01
17:47:30 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   97    0.00  0.01  0.97  0.00  0.00  0.01
17:47:30 D523.0   -             stand   0    NoReport stand              room -    Sit        2   97    0.00  0.01  0.37  0.01  0.00  0.01
17:47:31 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   98    0.00  0.01  0.97  0.00  0.00  0.01
17:47:31 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   98    0.00  0.01  0.37  0.01  0.00  0.01
17:47:31 D523.0   -             stand   0    NoReport stand              room -    Sit        2   98    0.00  0.01  0.37  0.01  0.00  0.01
17:47:32 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   99    0.00  0.01  0.97  0.00  0.00  0.01
17:47:32 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   99    0.00  0.01  0.37  0.01  0.00  0.01
17:47:32 D523.0   -             stand   0    NoReport stand              room -    Sit        2   99    0.00  0.01  0.37  0.01  0.00  0.01
17:47:33 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   100   0.00  0.01  0.37  0.01  0.00  0.01
17:47:33 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   100   0.00  0.01  0.97  0.00  0.00  0.01
17:47:33 D523.0   -             stand   0    NoReport stand              room -    Sit        2   100   0.00  0.01  0.37  0.01  0.00  0.01
17:47:34 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   101   0.00  0.01  0.37  0.01  0.00  0.01
17:47:34 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   101   0.00  0.01  0.97  0.00  0.00  0.01
17:47:34 D523.0   -             stand   0    NoReport stand              room -    Sit        2   101   0.00  0.01  0.37  0.01  0.00  0.01
17:47:35 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   102   0.00  0.01  0.37  0.01  0.00  0.01
17:47:35 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   102   0.00  0.01  0.97  0.00  0.00  0.01
17:47:35 D523.0   -             stand   0    NoReport stand              room -    Sit        2   102   0.00  0.01  0.37  0.01  0.00  0.01
17:47:36 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   103   0.00  0.01  0.97  0.00  0.00  0.01
17:47:36 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   103   0.00  0.01  0.37  0.01  0.00  0.01
17:47:36 D523.0   -             stand   0    NoReport stand              room -    Sit        2   103   0.00  0.01  0.37  0.01  0.00  0.01
17:47:37 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   104   0.00  0.01  0.97  0.00  0.00  0.01
17:47:37 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   104   0.00  0.01  0.37  0.01  0.00  0.01
17:47:37 D523.0   -             stand   0    NoReport stand              room -    Sit        2   104   0.00  0.01  0.37  0.01  0.00  0.01
17:47:38 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   105   0.00  0.01  0.97  0.00  0.00  0.01
17:47:38 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   105   0.00  0.01  0.37  0.01  0.00  0.01
17:47:38 D523.0   -             stand   0    NoReport stand              room -    Sit        2   105   0.00  0.01  0.37  0.01  0.00  0.01
17:47:39 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   106   0.00  0.01  0.37  0.01  0.00  0.01
17:47:39 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   106   0.00  0.01  0.97  0.00  0.00  0.01
17:47:39 D523.0   -             stand   0    NoReport stand              room -    Sit        2   106   0.00  0.01  0.37  0.01  0.00  0.01
17:47:40 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   107   0.00  0.01  0.37  0.01  0.00  0.01
17:47:40 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   107   0.00  0.01  0.97  0.00  0.00  0.01
17:47:40 D523.0   -             stand   0    NoReport stand              room -    Sit        2   107   0.00  0.01  0.37  0.01  0.00  0.01
17:47:41 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   108   0.00  0.01  0.37  0.01  0.00  0.01
17:47:41 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   108   0.00  0.01  0.97  0.00  0.00  0.01
17:47:41 D523.0   -             stand   0    NoReport stand              room -    Sit        2   108   0.00  0.01  0.37  0.01  0.00  0.01
17:47:42 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   109   0.00  0.01  0.37  0.01  0.00  0.01
17:47:42 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   109   0.00  0.01  0.97  0.00  0.00  0.01
17:47:42 D523.0   -             stand   0    NoReport stand              room -    Sit        2   109   0.00  0.01  0.37  0.01  0.00  0.01
17:47:43 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   110   0.00  0.01  0.37  0.01  0.00  0.01
17:47:43 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   110   0.00  0.01  0.97  0.00  0.00  0.01
17:47:43 D523.0   -             stand   0    NoReport stand              room -    Sit        2   110   0.00  0.01  0.37  0.01  0.00  0.01
17:47:44 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   111   0.00  0.01  0.37  0.01  0.00  0.01
17:47:44 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   111   0.00  0.01  0.97  0.00  0.00  0.01
17:47:44 D523.0   -             stand   0    NoReport stand              room -    Sit        2   111   0.00  0.01  0.37  0.01  0.00  0.01
17:47:45 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   112   0.00  0.01  0.97  0.00  0.00  0.01
17:47:45 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   112   0.00  0.01  0.37  0.01  0.00  0.01
17:47:45 D523.0   -             stand   0    NoReport stand              room -    Sit        2   112   0.00  0.01  0.37  0.01  0.00  0.01
17:47:46 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   113   0.00  0.01  0.97  0.00  0.00  0.01
17:47:46 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   113   0.00  0.01  0.37  0.01  0.00  0.01
17:47:46 D523.0   -             stand   0    NoReport stand              room -    Sit        2   113   0.00  0.01  0.37  0.01  0.00  0.01
17:47:47 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   114   0.00  0.01  0.37  0.01  0.00  0.01
17:47:47 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   114   0.00  0.01  0.97  0.00  0.00  0.01
17:47:47 D523.0   -             stand   0    NoReport stand              room -    Sit        2   114   0.00  0.01  0.37  0.01  0.00  0.01
17:47:48 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   115   0.00  0.01  0.37  0.01  0.00  0.01
17:47:48 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   115   0.00  0.01  0.97  0.00  0.00  0.01
17:47:48 D523.0   -             stand   0    NoReport stand              room -    Sit        2   115   0.00  0.01  0.37  0.01  0.00  0.01
17:47:49 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   116   0.00  0.01  0.37  0.01  0.00  0.01
17:47:49 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   116   0.00  0.01  0.97  0.00  0.00  0.01
17:47:49 D523.0   -             stand   0    NoReport stand              room -    Sit        2   116   0.00  0.01  0.37  0.01  0.00  0.01
17:47:50 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   117   0.00  0.01  0.37  0.01  0.00  0.01
17:47:50 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   117   0.00  0.01  0.97  0.00  0.00  0.01
17:47:50 D523.0   -             stand   0    NoReport stand              room -    Sit        2   117   0.00  0.01  0.37  0.01  0.00  0.01
17:47:51 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   117   0.00  0.01  0.37  0.01  0.00  0.01
17:47:51 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   117   0.00  0.01  0.97  0.00  0.00  0.01
17:47:51 D523.0   -             stand   0    NoReport stand              room -    Sit        2   117   0.00  0.01  0.37  0.01  0.00  0.01
17:47:52 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   118   0.00  0.01  0.37  0.01  0.00  0.01
17:47:52 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   118   0.00  0.01  0.97  0.00  0.00  0.01
17:47:52 D523.0   -             stand   0    NoReport stand              room -    Sit        2   118   0.00  0.01  0.37  0.01  0.00  0.01
17:47:53 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   120   0.00  0.01  0.37  0.01  0.00  0.01
17:47:53 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   120   0.00  0.01  0.97  0.00  0.00  0.01
17:47:53 D523.0   -             stand   0    NoReport stand              room -    Sit        2   120   0.00  0.01  0.37  0.01  0.00  0.01
17:47:54 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   121   0.00  0.01  0.37  0.01  0.00  0.01
17:47:54 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   121   0.00  0.01  0.97  0.00  0.00  0.01
17:47:54 D523.0   -             stand   0    NoReport stand              room -    Sit        2   121   0.00  0.01  0.37  0.01  0.00  0.01
17:47:55 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   121   0.00  0.01  0.97  0.00  0.00  0.01
17:47:55 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   121   0.00  0.01  0.37  0.01  0.00  0.01
17:47:55 D523.0   -             stand   0    NoReport stand              room -    Sit        2   121   0.00  0.01  0.37  0.01  0.00  0.01
17:47:56 09E7.88  -             88      -    NoReport no-target(88)      room -    Sit        2   121   0.00  0.01  0.37  0.01  0.00  0.01
17:47:56 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   122   0.00  0.01  0.37  0.01  0.00  0.01
17:47:56 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   122   0.00  0.01  0.97  0.00  0.00  0.01
17:47:56 D523.0   -             stand   0    NoReport stand              room -    Sit        2   122   0.00  0.01  0.37  0.01  0.00  0.01
17:47:57 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   123   0.00  0.01  0.97  0.00  0.00  0.01
17:47:57 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   123   0.00  0.01  0.37  0.01  0.00  0.01
17:47:57 D523.0   -             stand   0    NoReport stand              room -    Sit        2   123   0.00  0.01  0.37  0.01  0.00  0.01
17:47:58 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   124   0.00  0.01  0.97  0.00  0.00  0.01
17:47:58 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   124   0.00  0.01  0.37  0.01  0.00  0.01
17:47:58 D523.0   -             stand   0    NoReport stand              room -    Sit        2   124   0.00  0.01  0.37  0.01  0.00  0.01
17:47:59 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   125   0.00  0.01  0.97  0.00  0.00  0.01
17:47:59 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   125   0.00  0.01  0.37  0.01  0.00  0.01
17:47:59 D523.0   -             stand   0    NoReport stand              room -    Sit        2   125   0.00  0.01  0.37  0.01  0.00  0.01
17:48:00 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   126   0.00  0.01  0.37  0.01  0.00  0.01
17:48:00 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   126   0.00  0.01  0.97  0.00  0.00  0.01
17:48:00 D523.0   -             stand   0    NoReport stand              room -    Sit        2   126   0.00  0.01  0.37  0.01  0.00  0.01
17:48:01 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   127   0.00  0.01  0.37  0.01  0.00  0.01
17:48:01 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   127   0.00  0.01  0.97  0.00  0.00  0.01
17:48:01 D523.0   -             stand   85   NoReport stand              room -    Sit        2   127   0.00  0.01  0.37  0.01  0.00  0.01
17:48:02 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   128   0.00  0.01  0.37  0.01  0.00  0.01
17:48:02 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   128   0.00  0.01  0.97  0.00  0.00  0.01
17:48:02 D523.0   -             stand   105  NoReport stand              room -    Sit        2   128   0.00  0.01  0.37  0.01  0.00  0.01
17:48:03 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   129   0.00  0.01  0.37  0.01  0.00  0.01
17:48:03 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   129   0.00  0.01  0.97  0.00  0.00  0.01
17:48:03 D523.0   -             walk    106  NoReport walk               room -    Sit        2   129   0.00  0.01  0.37  0.01  0.00  0.01
17:48:04 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   130   0.00  0.01  0.97  0.00  0.00  0.01
17:48:04 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   130   0.00  0.01  0.37  0.01  0.00  0.01
17:48:04 D523.0   -             walk    114  NoReport walk               room -    Sit        2   130   0.00  0.01  0.37  0.01  0.00  0.01
17:48:05 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   131   0.00  0.01  0.97  0.00  0.00  0.01
17:48:05 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   131   0.00  0.01  0.37  0.01  0.00  0.01
17:48:05 D523.0   -             walk    106  NoReport walk               room -    Sit        2   131   0.00  0.01  0.37  0.01  0.00  0.01
17:48:06 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   132   0.00  0.01  0.97  0.00  0.00  0.01
17:48:06 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   132   0.00  0.01  0.37  0.01  0.00  0.01
17:48:06 D523.0   -             walk    109  NoReport walk               room -    Sit        2   132   0.00  0.01  0.37  0.01  0.00  0.01
17:48:07 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   133   0.00  0.01  0.37  0.01  0.00  0.01
17:48:07 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   133   0.00  0.01  0.97  0.00  0.00  0.01
17:48:07 D523.0   -             walk    98   NoReport walk               room -    Sit        2   133   0.00  0.01  0.37  0.01  0.00  0.01
17:48:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   134   0.00  0.01  0.37  0.01  0.00  0.01
17:48:08 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   134   0.00  0.01  0.97  0.00  0.00  0.01
17:48:08 D523.0   -             walk    107  NoReport walk               room -    Sit        2   134   0.00  0.01  0.37  0.01  0.00  0.01
17:48:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   135   0.00  0.01  0.37  0.01  0.00  0.01
17:48:08 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   135   0.00  0.01  0.97  0.00  0.00  0.01
17:48:09 D523.0   -             walk    96   NoReport walk               room -    Sit        2   135   0.00  0.01  0.37  0.01  0.00  0.01
17:48:09 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   136   0.00  0.01  0.37  0.01  0.00  0.01
17:48:09 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   136   0.00  0.01  0.97  0.00  0.00  0.01
17:48:10 D523.0   -             walk    106  NoReport walk               room -    Sit        2   136   0.00  0.01  0.37  0.01  0.00  0.01
17:48:10 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   137   0.00  0.01  0.37  0.01  0.00  0.01
17:48:10 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   137   0.00  0.01  0.97  0.00  0.00  0.01
17:48:11 D523.0   -             walk    102  NoReport walk               room -    Sit        2   137   0.00  0.01  0.37  0.01  0.00  0.01
17:48:12 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   138   0.00  0.01  0.97  0.00  0.00  0.01
17:48:12 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   138   0.00  0.01  0.37  0.01  0.00  0.01
17:48:12 D523.0   -             walk    100  NoReport walk               room -    Sit        2   138   0.00  0.01  0.37  0.01  0.00  0.01
17:48:13 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   139   0.00  0.01  0.97  0.00  0.00  0.01
17:48:13 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   139   0.00  0.01  0.37  0.01  0.00  0.01
17:48:13 D523.0   -             walk    95   NoReport walk               room -    Sit        2   139   0.00  0.01  0.37  0.01  0.00  0.01
17:48:13 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   140   0.00  0.01  0.97  0.00  0.00  0.01
17:48:13 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   140   0.00  0.01  0.37  0.01  0.00  0.01
17:48:14 D523.0   -             walk    0    NoReport walk               room -    Sit        2   140   0.00  0.01  0.37  0.01  0.00  0.01
17:48:14 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   141   0.00  0.01  0.37  0.01  0.00  0.01
17:48:14 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   141   0.00  0.01  0.97  0.00  0.00  0.01
17:48:15 D523.0   -             walk    0    NoReport walk               room -    Sit        2   141   0.00  0.01  0.37  0.01  0.00  0.01
17:48:15 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   142   0.00  0.01  0.97  0.00  0.00  0.01
17:48:15 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   142   0.00  0.01  0.37  0.01  0.00  0.01
17:48:16 D523.0   -             stand   0    NoReport stand              room -    Sit        2   142   0.00  0.01  0.37  0.01  0.00  0.01
17:48:16 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   143   0.00  0.01  0.37  0.01  0.00  0.01
17:48:16 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   143   0.00  0.01  0.97  0.00  0.00  0.01
17:48:17 D523.0   -             stand   0    NoReport stand              room -    Sit        2   143   0.00  0.01  0.37  0.01  0.00  0.01
17:48:17 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   144   0.00  0.01  0.97  0.00  0.00  0.01
17:48:17 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   144   0.00  0.01  0.37  0.01  0.00  0.01
17:48:18 D523.0   -             stand   0    NoReport stand              room -    Sit        2   144   0.00  0.01  0.37  0.01  0.00  0.01
17:48:18 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   145   0.00  0.01  0.97  0.00  0.00  0.01
17:48:18 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   145   0.00  0.01  0.37  0.01  0.00  0.01
17:48:19 D523.0   -             stand   0    NoReport stand              room -    Sit        2   145   0.00  0.01  0.37  0.01  0.00  0.01
17:48:19 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   146   0.00  0.01  0.97  0.00  0.00  0.01
17:48:19 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   146   0.00  0.01  0.37  0.01  0.00  0.01
17:48:20 D523.0   -             stand   0    NoReport stand              room -    Sit        2   146   0.00  0.01  0.37  0.01  0.00  0.01
17:48:20 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   147   0.00  0.01  0.96  0.00  0.00  0.01
17:48:20 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   147   0.00  0.01  0.36  0.01  0.01  0.01
17:48:21 D523.0   -             stand   0    NoReport stand              room -    Sit        2   147   0.00  0.01  0.36  0.01  0.01  0.01
17:48:21 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   148   0.00  0.01  0.97  0.00  0.00  0.01
17:48:21 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   148   0.00  0.01  0.37  0.01  0.01  0.01
17:48:22 D523.0   -             stand   0    NoReport stand              room -    Sit        2   148   0.00  0.01  0.37  0.01  0.01  0.01
17:48:22 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   149   0.00  0.01  0.97  0.00  0.00  0.01
17:48:22 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   149   0.00  0.01  0.37  0.01  0.00  0.01
17:48:23 D523.0   -             stand   0    NoReport stand              room -    Sit        2   149   0.00  0.01  0.37  0.01  0.00  0.01
17:48:23 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   150   0.00  0.01  0.97  0.00  0.00  0.01
17:48:23 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   150   0.00  0.01  0.37  0.01  0.00  0.01
17:48:24 D523.0   -             stand   96   NoReport stand              room -    Sit        2   150   0.00  0.01  0.37  0.01  0.00  0.01
17:48:24 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   151   0.00  0.01  0.97  0.00  0.00  0.01
17:48:24 D5F7.0   D5F704523166  sit     59   NoReport sit                trk  1.00 Sit        2   151   0.00  0.01  0.37  0.01  0.00  0.01
17:48:25 D523.0   -             stand   0    NoReport stand              room -    Sit        2   151   0.00  0.01  0.37  0.01  0.00  0.01
17:48:25 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   152   0.00  0.01  0.97  0.00  0.00  0.01
17:48:25 D5F7.0   D5F704523166  sit     77   NoReport sit                trk  1.00 Sit        2   152   0.00  0.01  0.37  0.01  0.00  0.01
17:48:26 D523.0   -             stand   0    NoReport stand              room -    Sit        2   152   0.00  0.01  0.37  0.01  0.00  0.01
17:48:26 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   153   0.00  0.01  0.94  0.00  0.00  0.01
17:48:26 D5F7.0   D5F704523166  sit     79   NoReport sit                trk  1.00 Sit        2   153   0.00  0.01  0.37  0.01  0.00  0.01
17:48:27 D523.0   -             stand   0    NoReport stand              room -    Sit        2   153   0.00  0.01  0.37  0.01  0.00  0.01
17:48:27 D5F7.0   D5F704523166  sit     64   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.37  0.01  0.00  0.01
17:48:27 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.87  0.00  0.00  0.02
17:48:28 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.87  0.00  0.00  0.02
17:48:28 09E7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   0     0.00  0.02  0.87  0.00  0.00  0.02
17:48:28 D5F7.0   D5F704523166  sit     83   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.00  0.70  0.00  0.00  0.01
17:48:28 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.92  0.00  0.00  0.01
17:48:29 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.92  0.00  0.00  0.01
17:48:29 D5F7.0   D5F704523166  sit     71   NoReport sit                trk  1.00 Sit        2   0     0.00  0.02  0.44  0.01  0.00  0.02
17:48:29 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.93  0.00  0.00  0.01
17:48:30 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.02  0.44  0.01  0.00  0.02
17:48:30 D5F7.0   D5F704523166  sit     67   NoReport sit                trk  1.00 Sit        2   0     0.00  0.02  0.27  0.01  0.01  0.02
17:48:30 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.93  0.00  0.00  0.01
17:48:31 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.02  0.27  0.01  0.01  0.02
17:48:31 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.93  0.00  0.00  0.01
17:48:31 D5F7.0   D5F704523166  sit     57   NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.31  0.01  0.01  0.01
17:48:32 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.31  0.01  0.01  0.01
17:48:32 D5F7.0   D5F704523166  sit     52   NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.32  0.01  0.00  0.01
17:48:32 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.93  0.00  0.00  0.01
17:48:33 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.32  0.01  0.00  0.01
17:48:33 D5F7.0   D5F704523166  sit     68   NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.20  0.01  0.00  0.01
17:48:33 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:48:34 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.20  0.01  0.00  0.01
17:48:34 D5F7.0   D5F704523166  sit     67   NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.25  0.01  0.00  0.01
17:48:34 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:48:35 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.25  0.01  0.00  0.01
17:48:35 D5F7.0   D5F704523166  sit     79   NoReport sit                trk  1.00 OpenFloor  2   0     0.00  0.01  0.28  0.01  0.00  0.01
17:48:35 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.88  0.00  0.00  0.02
17:48:36 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.88  0.00  0.00  0.02
17:48:36 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.30  0.01  0.00  0.01
17:48:36 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.92  0.00  0.00  0.01
17:48:37 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.30  0.01  0.00  0.01
17:48:37 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.32  0.01  0.00  0.01
17:48:37 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.96  0.00  0.00  0.01
17:48:38 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.32  0.01  0.00  0.01
17:48:38 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:48:38 D5F7.0   D5F704523166  sit     69   NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.33  0.01  0.00  0.01
17:48:39 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.33  0.01  0.00  0.01
17:48:39 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:48:39 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   0     0.00  0.01  0.34  0.01  0.00  0.01
17:48:40 D523.0   -             stand   0    NoReport stand              room -    Sit        2   0     0.00  0.01  0.34  0.01  0.00  0.01
17:48:40 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   14    0.00  0.01  0.97  0.00  0.00  0.01
17:48:40 D5F7.0   D5F704523166  sit     82   NoReport sit                trk  1.00 OpenFloor  2   14    0.00  0.01  0.51  0.01  0.00  0.01
17:48:41 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   14    0.00  0.01  0.51  0.01  0.00  0.01
17:48:41 D5F7.0   D5F704523166  sit     63   NoReport sit                trk  1.00 OpenFloor  2   15    0.00  0.01  0.47  0.01  0.00  0.01
17:48:41 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   15    0.00  0.01  0.97  0.00  0.00  0.01
17:48:42 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   15    0.00  0.01  0.47  0.01  0.00  0.01
17:48:42 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   16    0.00  0.01  0.44  0.01  0.00  0.01
17:48:42 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   16    0.00  0.01  0.97  0.00  0.00  0.01
17:48:42 D523.0   -             stand   0    NoReport stand              room -    Sit        2   16    0.00  0.01  0.44  0.01  0.00  0.01
17:48:43 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   17    0.00  0.01  0.43  0.01  0.00  0.01
17:48:43 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   17    0.00  0.01  0.94  0.00  0.00  0.01
17:48:43 D523.0   -             stand   0    NoReport stand              room -    Sit        2   17    0.00  0.01  0.43  0.01  0.00  0.01
17:48:44 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   18    0.00  0.01  0.41  0.01  0.00  0.01
17:48:44 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   18    0.00  0.01  0.97  0.00  0.00  0.01
17:48:44 D523.0   -             stand   0    NoReport stand              room -    Sit        2   18    0.00  0.01  0.41  0.01  0.00  0.01
17:48:45 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   19    0.00  0.01  0.97  0.00  0.00  0.01
17:48:45 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   19    0.00  0.01  0.40  0.01  0.00  0.01
17:48:45 D523.0   -             stand   0    NoReport stand              room -    Sit        2   19    0.00  0.01  0.40  0.01  0.00  0.01
17:48:46 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   20    0.00  0.01  0.39  0.01  0.00  0.01
17:48:46 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   20    0.00  0.01  0.97  0.00  0.00  0.01
17:48:46 D523.0   -             stand   0    NoReport stand              room -    Sit        2   20    0.00  0.01  0.39  0.01  0.00  0.01
17:48:47 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   21    0.00  0.01  0.97  0.00  0.00  0.01
17:48:47 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   21    0.00  0.01  0.39  0.01  0.00  0.01
17:48:47 D523.0   -             stand   0    NoReport stand              room -    Sit        2   21    0.00  0.01  0.39  0.01  0.00  0.01
17:48:48 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   22    0.00  0.01  0.38  0.01  0.00  0.01
17:48:48 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   22    0.00  0.01  0.97  0.00  0.00  0.01
17:48:48 D523.0   -             stand   0    NoReport stand              room -    Sit        2   22    0.00  0.01  0.38  0.01  0.00  0.01
17:48:49 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   23    0.00  0.01  0.38  0.01  0.00  0.01
17:48:49 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   23    0.00  0.01  0.97  0.00  0.00  0.01
17:48:49 D523.0   -             stand   0    NoReport stand              room -    Sit        2   23    0.00  0.01  0.38  0.01  0.00  0.01
17:48:50 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   24    0.00  0.01  0.38  0.01  0.00  0.01
17:48:50 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   24    0.00  0.01  0.97  0.00  0.00  0.01
17:48:50 D523.0   -             stand   0    NoReport stand              room -    Sit        2   24    0.00  0.01  0.38  0.01  0.00  0.01
17:48:51 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   25    0.00  0.01  0.37  0.01  0.00  0.01
17:48:51 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   25    0.00  0.01  0.97  0.00  0.00  0.01
17:48:51 D523.0   -             stand   0    NoReport stand              room -    Sit        2   25    0.00  0.01  0.37  0.01  0.00  0.01
17:48:52 D5F7.0   D5F704523166  sit     67   NoReport sit                trk  1.00 Sit        2   26    0.00  0.01  0.37  0.01  0.00  0.01
17:48:52 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   26    0.00  0.01  0.97  0.00  0.00  0.01
17:48:52 D523.0   -             stand   0    NoReport stand              room -    Sit        2   26    0.00  0.01  0.37  0.01  0.00  0.01
17:48:53 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   27    0.00  0.01  0.97  0.00  0.00  0.01
17:48:53 D5F7.0   D5F704523166  sit     79   NoReport sit                trk  1.00 Sit        2   27    0.00  0.01  0.37  0.01  0.00  0.01
17:48:53 D523.0   -             stand   0    NoReport stand              room -    Sit        2   27    0.00  0.01  0.37  0.01  0.00  0.01
17:48:54 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   28    0.00  0.01  0.37  0.01  0.00  0.01
17:48:54 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   28    0.00  0.01  0.97  0.00  0.00  0.01
17:48:54 D523.0   -             stand   0    NoReport stand              room -    Sit        2   28    0.00  0.01  0.37  0.01  0.00  0.01
17:48:55 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   29    0.00  0.01  0.37  0.01  0.00  0.01
17:48:55 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   29    0.00  0.01  0.97  0.00  0.00  0.01
17:48:55 D523.0   -             stand   0    NoReport stand              room -    Sit        2   29    0.00  0.01  0.37  0.01  0.00  0.01
17:48:56 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   30    0.00  0.01  0.37  0.01  0.00  0.01
17:48:56 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   30    0.00  0.01  0.97  0.00  0.00  0.01
17:48:56 D523.0   -             stand   0    NoReport stand              room -    Sit        2   30    0.00  0.01  0.37  0.01  0.00  0.01
17:48:57 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   31    0.00  0.01  0.37  0.01  0.00  0.01
17:48:57 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   31    0.00  0.01  0.97  0.00  0.00  0.01
17:48:57 D523.0   -             stand   0    NoReport stand              room -    Sit        2   31    0.00  0.01  0.37  0.01  0.00  0.01
17:48:58 D5F7.0   D5F704523166  sit     68   NoReport sit                trk  1.00 Sit        2   32    0.00  0.01  0.37  0.01  0.00  0.01
17:48:58 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   32    0.00  0.01  0.97  0.00  0.00  0.01
17:48:58 D523.0   -             stand   0    NoReport stand              room -    Sit        2   32    0.00  0.01  0.37  0.01  0.00  0.01
17:48:59 09E7.88  -             88      -    NoReport no-target(88)      room -    Sit        2   32    0.00  0.01  0.37  0.01  0.00  0.01
17:48:59 D5F7.0   D5F704523166  sit     84   NoReport sit                trk  1.00 OpenFloor  2   33    0.00  0.00  0.70  0.00  0.00  0.01
17:48:59 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   33    0.00  0.01  0.97  0.00  0.00  0.01
17:48:59 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   33    0.00  0.01  0.97  0.00  0.00  0.01
17:49:00 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   34    0.00  0.01  0.61  0.00  0.00  0.02
17:49:00 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   34    0.00  0.01  0.97  0.00  0.00  0.01
17:49:00 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   34    0.00  0.01  0.61  0.00  0.00  0.02
17:49:01 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   35    0.00  0.01  0.56  0.00  0.00  0.02
17:49:01 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   35    0.00  0.01  0.97  0.00  0.00  0.01
17:49:01 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   35    0.00  0.01  0.56  0.00  0.00  0.02
17:49:02 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   36    0.00  0.01  0.52  0.01  0.01  0.01
17:49:02 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   36    0.00  0.01  0.97  0.00  0.00  0.01
17:49:02 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   36    0.00  0.01  0.52  0.01  0.01  0.01
17:49:03 D5F7.0   D5F704523166  sit     87   NoReport sit                trk  1.00 OpenFloor  2   37    0.00  0.01  0.65  0.00  0.00  0.01
17:49:03 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   37    0.00  0.01  0.97  0.00  0.00  0.01
17:49:03 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   37    0.00  0.01  0.97  0.00  0.00  0.01
17:49:04 D5F7.0   D5F704523166  sit     81   NoReport sit                trk  1.00 OpenFloor  2   38    0.00  0.01  0.58  0.00  0.00  0.02
17:49:04 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   38    0.00  0.01  0.97  0.00  0.00  0.01
17:49:04 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   38    0.00  0.01  0.97  0.00  0.00  0.01
17:49:05 D5F7.0   D5F704523166  sit     103  NoReport sit                trk  1.00 OpenFloor  2   39    0.00  0.01  0.82  0.00  0.00  0.01
17:49:05 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   39    0.00  0.01  0.97  0.00  0.00  0.01
17:49:05 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   39    0.00  0.01  0.97  0.00  0.00  0.01
17:49:06 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   40    0.00  0.01  0.97  0.00  0.00  0.01
17:49:06 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   40    0.00  0.01  0.83  0.00  0.00  0.01
17:49:06 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   40    0.00  0.01  0.97  0.00  0.00  0.01
17:49:07 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   41    0.00  0.01  0.97  0.00  0.00  0.01
17:49:07 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   41    0.00  0.02  0.72  0.00  0.00  0.02
17:49:07 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   41    0.00  0.02  0.72  0.00  0.00  0.02
17:49:08 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   42    0.00  0.01  0.97  0.00  0.00  0.01
17:49:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   42    0.00  0.02  0.66  0.00  0.01  0.02
17:49:08 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   42    0.00  0.02  0.66  0.00  0.01  0.02
17:49:09 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   43    0.00  0.02  0.60  0.00  0.01  0.02
17:49:09 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   43    0.00  0.01  0.97  0.00  0.00  0.01
17:49:09 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   43    0.00  0.02  0.60  0.00  0.01  0.02
17:49:10 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   44    0.00  0.01  0.56  0.00  0.01  0.02
17:49:10 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   44    0.00  0.01  0.97  0.00  0.00  0.01
17:49:10 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   44    0.00  0.01  0.56  0.00  0.01  0.02
17:49:11 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   45    0.00  0.01  0.52  0.01  0.01  0.01
17:49:11 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   45    0.00  0.01  0.97  0.00  0.00  0.01
17:49:11 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   45    0.00  0.01  0.52  0.01  0.01  0.01
17:49:12 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   46    0.00  0.01  0.49  0.01  0.01  0.01
17:49:12 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   46    0.00  0.01  0.97  0.00  0.00  0.01
17:49:12 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   46    0.00  0.01  0.49  0.01  0.01  0.01
17:49:13 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   47    0.00  0.01  0.97  0.00  0.00  0.01
17:49:13 D5F7.0   D5F704523166  sit     77   NoReport sit                trk  1.00 Sit        2   47    0.00  0.01  0.46  0.01  0.01  0.01
17:49:13 D523.0   -             stand   0    NoReport stand              room -    Sit        2   47    0.00  0.01  0.46  0.01  0.01  0.01
17:49:14 D5F7.0   D5F704523166  sit     87   NoReport sit                trk  1.00 OpenFloor  2   48    0.00  0.01  0.76  0.00  0.00  0.01
17:49:14 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   48    0.00  0.02  0.88  0.00  0.00  0.02
17:49:14 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   48    0.00  0.02  0.88  0.00  0.00  0.02
17:49:15 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   49    0.00  0.02  0.49  0.00  0.00  0.02
17:49:15 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   49    0.00  0.01  0.92  0.00  0.00  0.01
17:49:15 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   49    0.00  0.02  0.49  0.00  0.00  0.02
17:49:16 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   50    0.00  0.02  0.87  0.00  0.00  0.02
17:49:16 D5F7.0   D5F704523166  sit     61   NoReport sit                trk  1.00 OpenFloor  2   50    0.00  0.01  0.48  0.01  0.01  0.01
17:49:16 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   50    0.00  0.02  0.87  0.00  0.00  0.02
17:49:17 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   51    0.00  0.01  0.96  0.00  0.00  0.01
17:49:17 D5F7.0   D5F704523166  sit     82   NoReport sit                trk  1.00 OpenFloor  2   51    0.00  0.01  0.63  0.00  0.00  0.01
17:49:17 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   51    0.00  0.01  0.96  0.00  0.00  0.01
17:49:18 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   52    0.00  0.01  0.56  0.00  0.00  0.02
17:49:18 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   52    0.00  0.01  0.97  0.00  0.00  0.01
17:49:18 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   52    0.00  0.01  0.56  0.00  0.00  0.02
17:49:19 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   53    0.00  0.02  0.48  0.01  0.01  0.01
17:49:19 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   53    0.00  0.01  0.96  0.00  0.00  0.01
17:49:19 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   53    0.00  0.02  0.48  0.01  0.01  0.01
17:49:20 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   54    0.00  0.01  0.97  0.00  0.00  0.01
17:49:20 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   54    0.00  0.01  0.46  0.01  0.01  0.01
17:49:20 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   54    0.00  0.01  0.46  0.01  0.01  0.01
17:49:21 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   55    0.00  0.01  0.44  0.01  0.01  0.01
17:49:21 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   55    0.00  0.01  0.97  0.00  0.00  0.01
17:49:21 D523.0   -             stand   0    NoReport stand              room -    Sit        2   55    0.00  0.01  0.44  0.01  0.01  0.01
17:49:22 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   56    0.00  0.02  0.88  0.00  0.00  0.02
17:49:22 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   56    0.00  0.01  0.42  0.01  0.00  0.01
17:49:22 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   56    0.00  0.02  0.88  0.00  0.00  0.02
17:49:23 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   57    0.00  0.01  0.41  0.01  0.00  0.01
17:49:23 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   57    0.00  0.01  0.92  0.00  0.00  0.01
17:49:23 D523.0   -             stand   0    NoReport stand              room -    Sit        2   57    0.00  0.01  0.41  0.01  0.00  0.01
17:49:24 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   58    0.00  0.01  0.40  0.01  0.00  0.01
17:49:24 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   58    0.00  0.01  0.96  0.00  0.00  0.01
17:49:24 D523.0   -             stand   0    NoReport stand              room -    Sit        2   58    0.00  0.01  0.40  0.01  0.00  0.01
17:49:25 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   59    0.00  0.01  0.97  0.00  0.00  0.01
17:49:25 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   59    0.00  0.01  0.39  0.01  0.00  0.01
17:49:25 D523.0   -             stand   0    NoReport stand              room -    Sit        2   59    0.00  0.01  0.39  0.01  0.00  0.01
17:49:26 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   60    0.00  0.01  0.39  0.01  0.00  0.01
17:49:26 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   60    0.00  0.01  0.97  0.00  0.00  0.01
17:49:26 D523.0   -             stand   0    NoReport stand              room -    Sit        2   60    0.00  0.01  0.39  0.01  0.00  0.01
17:49:27 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   61    0.00  0.01  0.38  0.01  0.00  0.01
17:49:27 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   61    0.00  0.01  0.97  0.00  0.00  0.01
17:49:27 D523.0   -             stand   0    NoReport stand              room -    Sit        2   61    0.00  0.01  0.38  0.01  0.00  0.01
17:49:28 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   62    0.00  0.01  0.97  0.00  0.00  0.01
17:49:28 D5F7.0   D5F704523166  sit     56   NoReport sit                trk  1.00 Sit        2   62    0.00  0.01  0.38  0.01  0.00  0.01
17:49:28 D523.0   -             stand   0    NoReport stand              room -    Sit        2   62    0.00  0.01  0.38  0.01  0.00  0.01
17:49:29 D5F7.0   D5F704523166  sit     62   NoReport sit                trk  1.00 Sit        2   63    0.00  0.01  0.38  0.01  0.00  0.01
17:49:29 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   63    0.00  0.01  0.97  0.00  0.00  0.01
17:49:29 D523.0   -             stand   0    NoReport stand              room -    Sit        2   63    0.00  0.01  0.38  0.01  0.00  0.01
17:49:30 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   64    0.00  0.01  0.94  0.00  0.00  0.01
17:49:30 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 Sit        2   64    0.00  0.01  0.37  0.01  0.00  0.01
17:49:30 D523.0   -             stand   0    NoReport stand              room -    Sit        2   64    0.00  0.01  0.37  0.01  0.00  0.01
17:49:31 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   65    0.00  0.01  0.37  0.01  0.00  0.01
17:49:31 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   65    0.00  0.02  0.87  0.00  0.00  0.02
17:49:31 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   65    0.00  0.02  0.87  0.00  0.00  0.02
17:49:31 09E7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   65    0.00  0.02  0.87  0.00  0.00  0.02
17:49:32 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   66    0.00  0.02  0.86  0.00  0.01  0.02
17:49:32 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   66    0.00  0.01  0.37  0.01  0.00  0.01
17:49:32 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   66    0.00  0.02  0.86  0.00  0.01  0.02
17:49:33 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   67    0.00  0.01  0.37  0.01  0.00  0.01
17:49:33 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   67    0.00  0.02  0.85  0.00  0.01  0.02
17:49:33 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   67    0.00  0.02  0.85  0.00  0.01  0.02
17:49:34 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   68    0.00  0.02  0.85  0.00  0.01  0.02
17:49:34 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   68    0.00  0.01  0.37  0.01  0.00  0.01
17:49:34 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   68    0.00  0.02  0.85  0.00  0.01  0.02
17:49:35 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   69    0.00  0.02  0.85  0.00  0.01  0.02
17:49:35 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   69    0.00  0.01  0.37  0.01  0.00  0.01
17:49:35 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   69    0.00  0.02  0.85  0.00  0.01  0.02
17:49:36 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   70    0.00  0.02  0.85  0.00  0.01  0.02
17:49:36 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   70    0.00  0.01  0.37  0.01  0.00  0.01
17:49:36 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   70    0.00  0.02  0.85  0.00  0.01  0.02
17:49:37 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   71    0.00  0.02  0.85  0.00  0.01  0.02
17:49:37 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   71    0.00  0.01  0.37  0.01  0.00  0.01
17:49:37 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   71    0.00  0.02  0.85  0.00  0.01  0.02
17:49:38 D5F7.0   D5F704523166  sit     73   NoReport sit                trk  1.00 OpenFloor  2   72    0.00  0.01  0.37  0.01  0.00  0.01
17:49:38 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   72    0.00  0.02  0.85  0.00  0.01  0.02
17:49:38 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   72    0.00  0.02  0.85  0.00  0.01  0.02
17:49:39 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   73    0.00  0.02  0.85  0.00  0.01  0.02
17:49:39 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   73    0.00  0.01  0.37  0.01  0.00  0.01
17:49:39 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   73    0.00  0.02  0.85  0.00  0.01  0.02
17:49:40 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   74    0.00  0.01  0.37  0.01  0.00  0.01
17:49:40 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   74    0.00  0.02  0.85  0.00  0.01  0.02
17:49:40 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   74    0.00  0.02  0.85  0.00  0.01  0.02
17:49:41 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   75    0.00  0.01  0.37  0.01  0.00  0.01
17:49:41 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   75    0.00  0.02  0.85  0.00  0.01  0.02
17:49:41 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   75    0.00  0.02  0.85  0.00  0.01  0.02
17:49:42 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   76    0.00  0.01  0.37  0.01  0.00  0.01
17:49:42 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   76    0.00  0.02  0.85  0.00  0.01  0.02
17:49:42 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   76    0.00  0.02  0.85  0.00  0.01  0.02
17:49:43 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   77    0.00  0.01  0.37  0.01  0.00  0.01
17:49:43 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   77    0.00  0.02  0.85  0.00  0.01  0.02
17:49:43 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   77    0.00  0.02  0.85  0.00  0.01  0.02
17:49:44 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   78    0.00  0.02  0.85  0.00  0.01  0.02
17:49:44 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   78    0.00  0.01  0.37  0.01  0.00  0.01
17:49:44 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   78    0.00  0.02  0.85  0.00  0.01  0.02
17:49:45 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   79    0.00  0.01  0.37  0.01  0.00  0.01
17:49:45 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   79    0.00  0.02  0.85  0.00  0.01  0.02
17:49:45 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   79    0.00  0.02  0.85  0.00  0.01  0.02
17:49:46 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   80    0.00  0.02  0.85  0.00  0.01  0.02
17:49:46 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   80    0.00  0.01  0.37  0.01  0.00  0.01
17:49:46 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   80    0.00  0.02  0.85  0.00  0.01  0.02
17:49:47 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   81    0.00  0.02  0.85  0.00  0.01  0.02
17:49:47 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   81    0.00  0.01  0.37  0.01  0.00  0.01
17:49:47 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   81    0.00  0.02  0.85  0.00  0.01  0.02
17:49:48 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   82    0.00  0.01  0.37  0.01  0.00  0.01
17:49:48 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   82    0.00  0.02  0.85  0.00  0.01  0.02
17:49:48 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   82    0.00  0.02  0.85  0.00  0.01  0.02
17:49:49 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   83    0.00  0.01  0.37  0.01  0.00  0.01
17:49:49 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   83    0.00  0.02  0.85  0.00  0.01  0.02
17:49:49 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   83    0.00  0.02  0.85  0.00  0.01  0.02
17:49:50 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   84    0.00  0.02  0.85  0.00  0.01  0.02
17:49:50 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   84    0.00  0.01  0.37  0.01  0.00  0.01
17:49:50 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   84    0.00  0.02  0.85  0.00  0.01  0.02
17:49:51 D5F7.0   D5F704523166  sit     82   NoReport sit                trk  1.00 OpenFloor  2   85    0.00  0.00  0.70  0.00  0.00  0.01
17:49:51 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   85    0.00  0.02  0.85  0.00  0.01  0.02
17:49:51 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   85    0.00  0.02  0.85  0.00  0.01  0.02
17:49:52 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   86    0.00  0.01  0.96  0.00  0.00  0.01
17:49:52 D5F7.0   D5F704523166  sit     74   NoReport sit                trk  1.00 OpenFloor  2   86    0.00  0.01  0.76  0.00  0.00  0.01
17:49:52 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   86    0.00  0.01  0.96  0.00  0.00  0.01
17:49:53 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   87    0.00  0.01  0.66  0.00  0.00  0.02
17:49:53 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   87    0.00  0.01  0.97  0.00  0.00  0.01
17:49:53 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   87    0.00  0.01  0.66  0.00  0.00  0.02
17:49:54 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   88    0.00  0.01  0.94  0.00  0.00  0.01
17:49:54 D5F7.0   D5F704523166  sit     70   NoReport sit                trk  1.00 OpenFloor  2   88    0.00  0.02  0.60  0.00  0.01  0.02
17:49:54 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   88    0.00  0.02  0.60  0.00  0.01  0.02
17:49:55 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   89    0.00  0.01  0.56  0.00  0.01  0.02
17:49:55 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   89    0.00  0.01  0.93  0.00  0.00  0.01
17:49:55 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   89    0.00  0.01  0.56  0.00  0.01  0.02
17:49:56 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   90    0.00  0.01  0.93  0.00  0.00  0.01
17:49:56 D5F7.0   D5F704523166  sit     61   NoReport sit                trk  1.00 OpenFloor  2   90    0.00  0.01  0.52  0.01  0.01  0.01
17:49:56 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   90    0.00  0.01  0.52  0.01  0.01  0.01
17:49:57 D5F7.0   D5F704523166  sit     53   NoReport sit                trk  1.00 Sit        2   91    0.00  0.02  0.32  0.01  0.01  0.02
17:49:57 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 Sit        2   91    0.00  0.01  0.97  0.00  0.00  0.01
17:49:57 D523.0   -             stand   0    NoReport stand              room -    Sit        2   91    0.00  0.02  0.32  0.01  0.01  0.02
17:49:58 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   92    0.00  0.02  0.88  0.00  0.00  0.02
17:49:58 D5F7.0   D5F704523166  sit     82   NoReport sit                trk  1.00 OpenFloor  2   92    0.00  0.01  0.51  0.01  0.01  0.01
17:49:58 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   92    0.00  0.02  0.88  0.00  0.00  0.02
17:49:59 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   93    0.00  0.01  0.92  0.00  0.00  0.01
17:49:59 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   93    0.00  0.01  0.64  0.00  0.00  0.01
17:49:59 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   93    0.00  0.01  0.92  0.00  0.00  0.01
17:50:00 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   94    0.00  0.01  0.57  0.00  0.00  0.02
17:50:00 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   94    0.00  0.01  0.93  0.00  0.00  0.01
17:50:00 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   94    0.00  0.01  0.57  0.00  0.00  0.02
17:50:01 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   95    0.00  0.02  0.87  0.00  0.00  0.02
17:50:01 D5F7.0   D5F704523166  sit     68   NoReport sit                trk  1.00 OpenFloor  2   95    0.00  0.01  0.52  0.01  0.00  0.01
17:50:01 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   95    0.00  0.02  0.87  0.00  0.00  0.02
17:50:02 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   96    0.00  0.02  0.86  0.00  0.01  0.02
17:50:02 D5F7.0   D5F704523166  sit     73   NoReport sit                trk  1.00 OpenFloor  2   96    0.00  0.01  0.49  0.01  0.01  0.01
17:50:02 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   96    0.00  0.02  0.86  0.00  0.01  0.02
17:50:02 09E7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   96    0.00  0.02  0.86  0.00  0.01  0.02
17:50:03 D5F7.0   D5F704523166  sit     49   NoReport sit                trk  1.00 OpenFloor  2   97    0.00  0.01  0.46  0.01  0.01  0.01
17:50:03 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   97    0.00  0.02  0.85  0.00  0.01  0.02
17:50:03 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   97    0.00  0.02  0.85  0.00  0.01  0.02
17:50:04 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   98    0.00  0.01  0.96  0.00  0.00  0.01
17:50:04 D5F7.0   D5F704523166  stand   59   NoReport stand              trk  1.00 OpenFloor  2   98    0.00  0.01  0.74  0.01  0.00  0.01
17:50:04 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   98    0.00  0.01  0.74  0.01  0.00  0.01
17:50:05 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   99    0.00  0.01  0.97  0.00  0.00  0.01
17:50:05 D5F7.0   D5F704523166  stand   67   NoReport stand              trk  1.00 OpenFloor  2   99    0.00  0.01  0.95  0.00  0.00  0.01
17:50:05 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   99    0.00  0.01  0.95  0.00  0.00  0.01
17:50:06 D5F7.0   D5F704523166  stand   80   NoReport stand              trk  1.00 OpenFloor  2   100   0.00  0.01  0.97  0.00  0.00  0.01
17:50:06 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   100   0.00  0.01  0.97  0.00  0.00  0.01
17:50:06 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   100   0.00  0.01  0.97  0.00  0.00  0.01
17:50:07 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   101   0.00  0.01  0.94  0.00  0.00  0.01
17:50:07 D5F7.0   D5F704523166  sit     66   NoReport sit                trk  1.00 OpenFloor  2   101   0.00  0.01  0.93  0.00  0.00  0.01
17:50:07 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   101   0.00  0.01  0.94  0.00  0.00  0.01
17:50:08 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   102   0.00  0.02  0.81  0.00  0.00  0.02
17:50:08 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   102   0.00  0.01  0.93  0.00  0.00  0.01
17:50:08 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   102   0.00  0.02  0.81  0.00  0.00  0.02
17:50:09 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   103   0.00  0.02  0.87  0.00  0.00  0.02
17:50:09 D5F7.0   D5F704523166  sit     0    NoReport sit                trk  1.00 OpenFloor  2   103   0.00  0.02  0.73  0.00  0.01  0.02
17:50:09 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   103   0.00  0.02  0.73  0.00  0.01  0.02
17:50:10 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   104   0.00  0.01  0.96  0.00  0.00  0.01
17:50:10 D5F7.0   D5F704523166  sit     94   NoReport sit                trk  1.00 OpenFloor  2   104   0.00  0.01  0.80  0.00  0.00  0.01
17:50:10 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   104   0.00  0.01  0.96  0.00  0.00  0.01
17:50:11 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   105   0.00  0.01  0.94  0.00  0.00  0.01
17:50:11 D5F7.0   D5F704523166  sit     56   NoReport sit                trk  1.00 OpenFloor  2   105   0.00  0.01  0.83  0.00  0.00  0.01
17:50:11 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   105   0.00  0.01  0.94  0.00  0.00  0.01
17:50:12 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   106   0.00  0.01  0.93  0.00  0.00  0.01
17:50:12 D5F7.0   D5F704523166  sit     53   NoReport sit                trk  1.00 OpenFloor  2   106   0.00  0.02  0.72  0.00  0.00  0.02
17:50:12 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   106   0.00  0.02  0.72  0.00  0.00  0.02
17:50:12 D5F7.0   D5F704523166  sit     68   NoReport sit                trk  1.00 OpenFloor  2   107   0.00  0.02  0.66  0.00  0.01  0.02
17:50:12 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   107   0.00  0.01  0.97  0.00  0.00  0.01
17:50:13 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   107   0.00  0.02  0.66  0.00  0.01  0.02
17:50:13 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   108   0.00  0.02  0.88  0.00  0.00  0.02
17:50:13 D5F7.0   D5F704523166  sit     81   NoReport sit                trk  1.00 OpenFloor  2   108   0.00  0.01  0.75  0.00  0.00  0.01
17:50:14 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   108   0.00  0.02  0.88  0.00  0.00  0.02
17:50:14 D5F7.0   D5F704523166  sit     103  NoReport sit                trk  1.00 OpenFloor  2   109   0.00  0.01  0.89  0.00  0.00  0.01
17:50:14 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   109   0.00  0.02  0.86  0.00  0.01  0.02
17:50:15 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   109   0.00  0.02  0.86  0.00  0.01  0.02
17:50:16 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   110   0.00  0.01  0.96  0.00  0.00  0.01
17:50:16 D5F7.0   D5F704523166  sit     61   NoReport sit                trk  1.00 OpenFloor  2   110   0.00  0.01  0.87  0.00  0.00  0.01
17:50:16 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   110   0.00  0.01  0.96  0.00  0.00  0.01
17:50:17 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   111   0.00  0.01  0.97  0.00  0.00  0.01
17:50:17 D5F7.0   D5F704523166  sit     81   NoReport sit                trk  1.00 OpenFloor  2   111   0.00  0.01  0.86  0.00  0.00  0.01
17:50:17 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   111   0.00  0.01  0.97  0.00  0.00  0.01
17:50:17 D5F7.0   D5F704523166  sit     86   NoReport sit                trk  1.00 OpenFloor  2   112   0.00  0.01  0.92  0.00  0.00  0.01
17:50:17 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   112   0.00  0.01  0.94  0.00  0.00  0.01
17:50:18 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   112   0.00  0.01  0.94  0.00  0.00  0.01
17:50:18 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   113   0.00  0.02  0.87  0.00  0.00  0.02
17:50:18 D5F7.0   D5F704523166  sit     94   NoReport sit                trk  1.00 OpenFloor  2   113   0.00  0.00  0.94  0.00  0.00  0.01
17:50:19 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   113   0.00  0.02  0.87  0.00  0.00  0.02
17:50:20 D5F7.0   D5F704523166  sit     105  NoReport sit                trk  1.00 OpenFloor  2   114   0.00  0.01  0.93  0.00  0.00  0.01
17:50:20 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   114   0.00  0.03  0.81  0.00  0.02  0.02
17:50:20 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   114   0.00  0.03  0.81  0.00  0.02  0.02
17:50:20 D5F7.0   D5F704523166  sit     56   NoReport sit                trk  1.00 OpenFloor  2   115   0.00  0.02  0.80  0.00  0.00  0.02
17:50:20 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   115   0.00  0.02  0.84  0.00  0.01  0.02
17:50:21 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   115   0.00  0.02  0.84  0.00  0.01  0.02
17:50:21 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   115   0.00  0.01  0.92  0.00  0.01  0.01
17:50:21 D5F7.0   D5F704523166  sit     49   NoReport sit                trk  1.00 OpenFloor  2   115   0.00  0.02  0.72  0.00  0.01  0.02
17:50:22 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   115   0.00  0.02  0.72  0.00  0.01  0.02
17:50:22 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   117   0.00  0.01  0.93  0.00  0.00  0.01
17:50:22 D5F7.0   D5F704523166  sit     70   NoReport sit                trk  1.00 OpenFloor  2   117   0.00  0.02  0.66  0.00  0.01  0.02
17:50:23 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   117   0.00  0.02  0.66  0.00  0.01  0.02
17:50:23 D5F7.0   D5F704523166  sit     65   NoReport sit                trk  1.00 OpenFloor  2   118   0.00  0.02  0.60  0.00  0.01  0.02
17:50:23 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   118   0.00  0.02  0.87  0.00  0.00  0.02
17:50:24 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   118   0.00  0.02  0.87  0.00  0.00  0.02
17:50:24 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   119   0.00  0.01  0.96  0.00  0.00  0.01
17:50:24 D5F7.0   D5F704523166  sit     22   NoReport sit                trk  1.00 OpenFloor  2   119   0.00  0.02  0.56  0.00  0.01  0.02
17:50:25 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   119   0.00  0.02  0.56  0.00  0.01  0.02
17:50:25 D5F7.0   D5F704523166  stand   51   NoReport stand              trk  1.00 OpenFloor  2   120   0.00  0.01  0.92  0.00  0.00  0.01
17:50:25 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   120   0.00  0.01  0.97  0.00  0.00  0.01
17:50:26 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   120   0.00  0.01  0.92  0.00  0.00  0.01
17:50:26 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   120   0.00  0.01  0.97  0.00  0.00  0.01
17:50:26 D5F7.0   D5F704523166  stand   63   NoReport stand              trk  1.00 OpenFloor  2   120   0.00  0.01  0.96  0.00  0.00  0.01
17:50:27 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   120   0.00  0.01  0.96  0.00  0.00  0.01
17:50:27 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   121   0.00  0.01  0.97  0.00  0.00  0.01
17:50:27 D5F7.0   D5F704523166  stand   70   NoReport stand              trk  1.00 OpenFloor  2   121   0.00  0.01  0.97  0.00  0.00  0.01
17:50:28 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   121   0.00  0.01  0.97  0.00  0.00  0.01
17:50:28 D5F7.0   D5F704523166  walk    69   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:28 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
17:50:29 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
17:50:29 D5F7.0   D5F704523166  walk    88   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:29 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.87  0.00  0.00  0.02
17:50:30 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.87  0.00  0.00  0.02
17:50:30 D5F7.0   D5F704523166  walk    93   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:30 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.02  0.86  0.00  0.01  0.02
17:50:31 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.02  0.86  0.00  0.01  0.02
17:50:31 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.96  0.00  0.00  0.01
17:50:31 D5F7.0   D5F704523166  walk    82   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:32 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.96  0.00  0.00  0.01
17:50:32 D5F7.0   D5F704523166  walk    98   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:32 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:33 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:33 D5F7.0   D5F704523166  walk    92   NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:33 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
17:50:34 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.94  0.00  0.00  0.01
17:50:34 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:34 D5F7.0   D5F704523166  walk    0    NoReport walk               trk  1.00 OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:34 09E7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:35 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   0     0.00  0.01  0.97  0.00  0.00  0.01
17:50:35 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   14    0.00  0.01  0.97  0.00  0.00  0.01
17:50:35 D5F7.0   D5F704523166  walk    83   NoReport walk               trk  1.00 OpenFloor  2   14    0.00  0.01  0.97  0.00  0.00  0.01
17:50:36 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   14    0.00  0.01  0.97  0.00  0.00  0.01
17:50:36 D5F7.0   D5F704523166  walk    72   NoReport walk               trk  1.00 OpenFloor  2   15    0.00  0.01  0.97  0.00  0.00  0.01
17:50:36 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   15    0.00  0.01  0.97  0.00  0.00  0.01
17:50:37 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   15    0.00  0.01  0.97  0.00  0.00  0.01
17:50:37 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   16    0.00  0.01  0.97  0.00  0.00  0.01
17:50:37 D5F7.0   D5F704523166  walk    69   NoReport walk               trk  1.00 OpenFloor  2   16    0.00  0.01  0.97  0.00  0.00  0.01
17:50:38 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   16    0.00  0.01  0.97  0.00  0.00  0.01
17:50:38 D5F7.0   D5F704523166  walk    107  NoReport walk               trk  1.00 OpenFloor  2   17    0.00  0.01  0.97  0.00  0.00  0.01
17:50:38 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   17    0.00  0.01  0.97  0.00  0.00  0.01
17:50:39 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   17    0.00  0.01  0.97  0.00  0.00  0.01
17:50:39 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   18    0.00  0.01  0.94  0.00  0.00  0.01
17:50:39 D5F7.0   D5F704523166  walk    77   NoReport walk               trk  1.00 OpenFloor  2   18    0.00  0.01  0.97  0.00  0.00  0.01
17:50:40 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   18    0.00  0.01  0.94  0.00  0.00  0.01
17:50:40 D5F7.0   D5F704523166  walk    76   NoReport walk               trk  1.00 OpenFloor  2   19    0.00  0.01  0.97  0.00  0.00  0.01
17:50:40 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   19    0.00  0.01  0.93  0.00  0.00  0.01
17:50:41 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   19    0.00  0.01  0.93  0.00  0.00  0.01
17:50:41 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
17:50:41 D5F7.0   D5F704523166  walk    0    NoReport walk               trk  1.00 OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
17:50:41 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   20    0.00  0.01  0.97  0.00  0.00  0.01
17:50:42 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
17:50:42 D5F7.0   D5F704523166  walk    0    NoReport walk               trk  1.00 OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
17:50:42 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   21    0.00  0.01  0.97  0.00  0.00  0.01
17:50:43 D5F7.0   D5F704523166  stand   0    NoReport stand              trk  1.00 OpenFloor  2   22    0.00  0.01  0.97  0.00  0.00  0.01
17:50:43 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   22    0.00  0.01  0.97  0.00  0.00  0.01
17:50:43 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   22    0.00  0.01  0.97  0.00  0.00  0.01
17:50:44 D5F7.0   D5F704523166  stand   0    NoReport stand              trk  1.00 OpenFloor  2   23    0.00  0.01  0.97  0.00  0.00  0.01
17:50:44 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   23    0.00  0.01  0.97  0.00  0.00  0.01
17:50:44 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   23    0.00  0.01  0.97  0.00  0.00  0.01
17:50:45 D5F7.0   D5F704523166  stand   0    NoReport stand              trk  1.00 OpenFloor  2   24    0.00  0.01  0.97  0.00  0.00  0.01
17:50:45 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   24    0.00  0.01  0.97  0.00  0.00  0.01
17:50:45 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   24    0.00  0.01  0.97  0.00  0.00  0.01
17:50:46 D5F7.E   -             -       0    NoReport ExitRoom(rdr)      room -    OpenFloor  2   24    0.00  0.01  0.97  0.00  0.00  0.01
17:50:46 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   25    0.00  0.01  0.97  0.00  0.00  0.01
17:50:46 D5F7.0   D5F704523166  stand   0    NoReport stand              trk  1.00 OpenFloor  2   25    0.00  0.01  0.97  0.00  0.00  0.01
17:50:46 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   25    0.00  0.01  0.97  0.00  0.00  0.01
17:50:47 D5F7.E   -             -       0    NoReport np=1               room -    OpenFloor  2   25    0.00  0.01  0.97  0.00  0.00  0.01
17:50:47 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  2   26    0.00  0.02  0.88  0.00  0.00  0.02
17:50:47 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  2   26    0.00  0.02  0.88  0.00  0.00  0.02
17:50:48 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  0.55 OpenFloor  1   27    0.00  0.02  0.86  0.00  0.01  0.02
17:50:48 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   27    0.00  0.02  0.86  0.00  0.01  0.02
17:50:49 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
17:50:49 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   28    0.00  0.02  0.85  0.00  0.01  0.02
17:50:50 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
17:50:50 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   29    0.00  0.02  0.85  0.00  0.01  0.02
17:50:51 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
17:50:51 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   30    0.00  0.02  0.85  0.00  0.01  0.02
17:50:52 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
17:50:52 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   31    0.00  0.02  0.85  0.00  0.01  0.02
17:50:53 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
17:50:53 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   32    0.00  0.02  0.85  0.00  0.01  0.02
17:50:54 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
17:50:54 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   33    0.00  0.02  0.85  0.00  0.01  0.02
17:50:55 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   34    0.00  0.02  0.85  0.00  0.01  0.02
17:50:55 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   34    0.00  0.02  0.85  0.00  0.01  0.02
17:50:56 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
17:50:56 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   35    0.00  0.02  0.85  0.00  0.01  0.02
17:50:57 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
17:50:57 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   36    0.00  0.02  0.85  0.00  0.01  0.02
17:50:58 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
17:50:58 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   37    0.00  0.02  0.85  0.00  0.01  0.02
17:50:59 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   38    0.00  0.02  0.85  0.00  0.01  0.02
17:50:59 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   38    0.00  0.02  0.85  0.00  0.01  0.02
17:51:00 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   39    0.00  0.02  0.85  0.00  0.01  0.02
17:51:00 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   39    0.00  0.02  0.85  0.00  0.01  0.02
17:51:01 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   40    0.00  0.02  0.85  0.00  0.01  0.02
17:51:01 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   40    0.00  0.02  0.85  0.00  0.01  0.02
17:51:02 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   41    0.00  0.02  0.84  0.00  0.01  0.03
17:51:02 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   41    0.00  0.02  0.84  0.00  0.01  0.03
17:51:03 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   42    0.00  0.02  0.84  0.00  0.01  0.03
17:51:03 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   42    0.00  0.02  0.84  0.00  0.01  0.03
17:51:04 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   43    0.00  0.02  0.83  0.00  0.01  0.04
17:51:04 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   43    0.00  0.02  0.83  0.00  0.01  0.04
17:51:05 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   44    0.00  0.02  0.82  0.00  0.02  0.05
17:51:05 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   44    0.00  0.02  0.82  0.00  0.02  0.05
17:51:06 09E7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   44    0.00  0.02  0.82  0.00  0.02  0.05
17:51:06 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   45    0.00  0.02  0.80  0.00  0.02  0.07
17:51:06 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   45    0.00  0.02  0.80  0.00  0.02  0.07
17:51:07 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   46    0.00  0.02  0.78  0.00  0.03  0.08
17:51:07 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   46    0.00  0.02  0.78  0.00  0.03  0.08
17:51:08 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   47    0.00  0.02  0.75  0.00  0.03  0.11
17:51:08 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   47    0.00  0.02  0.75  0.00  0.03  0.11
17:51:09 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   48    0.00  0.02  0.72  0.00  0.04  0.14
17:51:09 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   48    0.00  0.02  0.72  0.00  0.04  0.14
17:51:10 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   49    0.00  0.02  0.68  0.00  0.05  0.18
17:51:10 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   49    0.00  0.02  0.68  0.00  0.05  0.18
17:51:11 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   50    0.00  0.02  0.62  0.00  0.07  0.23
17:51:11 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   50    0.00  0.02  0.62  0.00  0.07  0.23
17:51:12 D5F7.1   D5F714627904  stand   0    NoReport stand              trk  1.00 OpenFloor  1   51    0.00  0.01  0.55  0.00  0.08  0.29
17:51:12 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   51    0.00  0.01  0.55  0.00  0.08  0.29
17:51:13 D5F7.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  1   51    0.00  0.01  0.55  0.00  0.08  0.29
17:51:13 D5F7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  1   52    0.00  0.02  0.71  0.00  0.15  0.04
17:51:13 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  1   52    0.00  0.02  0.71  0.00  0.15  0.04
17:51:14 D5F7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.05  0.49  0.01  0.17  0.05
17:51:14 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.05  0.49  0.01  0.17  0.05
17:51:15 D5F7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:15 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:16 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:17 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:18 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:19 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:20 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:21 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:22 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:23 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:24 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:25 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:26 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:27 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.07  0.37  0.02  0.20  0.04
17:51:28 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.01  0.08  0.29  0.03  0.22  0.03
17:51:28 D5F7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:29 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:30 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:31 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:32 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:33 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:34 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:35 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:36 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:37 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:38 09E7.E   -             -       0    NoReport np=0  ★0           room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:38 09E7.88  -             88      -    NoReport no-target(88)      room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:38 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:39 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:40 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:41 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:42 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:43 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:44 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:45 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:46 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:47 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:48 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:49 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:50 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:51 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:52 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:53 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:54 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:55 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:56 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:57 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:58 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:51:59 D523.0   -             stand   0    NoReport stand              room -    OpenFloor  0   0     0.02  0.09  0.24  0.04  0.23  0.03
17:52:00 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:00 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:01 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:02 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:03 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:04 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:05 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:06 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:07 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:08 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:09 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:10 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:10 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:11 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:12 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:13 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:14 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:15 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:16 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:17 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:18 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:19 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:20 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:21 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:22 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:23 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:24 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:25 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:26 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:27 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:28 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:29 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:30 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:31 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.02  0.09  0.22  0.05  0.24  0.03
17:52:32 D5F7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:32 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:33 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:34 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:35 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:36 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:37 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:38 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:39 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:40 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:41 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:41 09E7.88  -             88      -    NoReport no-target(88)      room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:42 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:43 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:44 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:45 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:46 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:46 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:47 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:48 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:49 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:50 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:51 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:52 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:53 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:54 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:55 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:56 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:57 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
17:52:58 D523.0   -             stand   0    NoReport stand              room -    Empty      0   0     0.03  0.09  0.19  0.07  0.24  0.02
```

## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）
```
time         dev.tid   pose   area x      y      z     conf Δcm  
17:43:10.294 09E7.88   88     -    -      -      -     -    -    
17:43:42.399 09E7.88   88     -    -      -      -     -    -    
17:44:14.039 09E7.88   88     -    -      -      -     -    -    
17:44:45.576 09E7.88   88     -    -      -      -     -    -    
17:45:17.358 09E7.88   88     -    -      -      -     -    -    
17:45:49.010 09E7.88   88     -    -      -      -     -    -    
17:46:20.962 09E7.88   88     -    -      -      -     -    -    
17:46:52.500 09E7.88   88     -    -      -      -     -    -    
17:47:24.521 09E7.88   88     -    -      -      -     -    -    
17:47:56.047 09E7.88   88     -    -      -      -     -    -    
17:48:28.096 09E7.88   88     -    -      -      -     -    -    
17:48:59.532 09E7.88   88     -    -      -      -     -    -    
17:49:31.689 09E7.88   88     -    -      -      -     -    -    
17:50:02.960 09E7.88   88     -    -      -      -     -    -    
17:50:34.969 09E7.88   88     -    -      -      -     -    -    
17:51:06.485 09E7.88   88     -    -      -      -     -    -    
17:51:38.306 09E7.88   88     -    -      -      -     -    -    
17:52:10.209 09E7.88   88     -    -      -      -     -    -    
17:52:41.682 09E7.88   88     -    -      -      -     -    -    

17:43:00.732 D523.0    stand  None -310   480    0     80        
17:43:01.733 D523.0    stand  None -310   480    0     80   0    
17:43:02.732 D523.0    stand  None -310   480    0     80   0    
17:43:03.732 D523.0    stand  None -310   480    0     80   0    
17:43:04.732 D523.0    stand  None -310   480    0     80   0    
17:43:05.628 D523.0    stand  None -310   480    0     80   0    
17:43:06.698 D523.0    stand  None -310   480    0     80   0    
17:43:07.634 D523.0    stand  None -310   480    0     80   0    
17:43:08.628 D523.0    stand  None -310   480    0     80   0    
17:43:09.704 D523.0    stand  None -310   480    0     80   0    
17:43:10.585 D523.0    stand  None -290   500    0     80   28   
17:43:11.608 D523.0    stand  None -290   500    0     80   0    
17:43:12.596 D523.0    stand  None -290   500    0     80   0    
17:43:13.602 D523.0    stand  None -290   500    0     80   0    
17:43:14.589 D523.0    stand  None -290   500    0     80   0    
17:43:15.590 D523.0    stand  None -290   500    0     80   0    
17:43:16.592 D523.0    stand  None -290   500    0     80   0    
17:43:17.600 D523.0    stand  None -290   500    0     80   0    
17:43:18.598 D523.0    stand  None -290   500    0     80   0    
17:43:19.594 D523.0    stand  None -290   500    0     80   0    
17:43:20.595 D523.0    stand  None -290   500    0     80   0    
17:43:21.599 D523.0    stand  None -290   500    0     80   0    
17:43:22.490 D523.0    stand  None -290   500    0     80   0    
17:43:23.490 D523.0    stand  None -290   500    0     80   0    
17:43:24.496 D523.0    stand  None -290   500    0     80   0    
17:43:25.493 D523.0    stand  None -290   500    0     80   0    
17:43:26.542 D523.0    stand  None -290   500    0     80   0    
17:43:27.546 D523.0    stand  None -290   500    0     80   0    
17:43:28.448 D523.0    stand  None -290   500    0     80   0    
17:43:29.461 D523.0    stand  None -290   500    0     80   0    
17:43:30.445 D523.0    stand  None -280   500    0     80   10   
17:43:31.445 D523.0    stand  None -280   500    0     80   0    
17:43:32.445 D523.0    stand  None -310   490    0     80   31   
17:43:33.451 D523.0    stand  None -310   490    0     80   0    
17:43:34.448 D523.0    stand  None -310   490    0     80   0    
17:43:35.458 D523.0    stand  None -310   490    0     80   0    
17:43:36.460 D523.0    stand  None -310   490    0     80   0    
17:43:37.500 D523.0    stand  None -310   490    0     80   0    
17:43:38.502 D523.0    stand  None -310   490    0     80   0    
17:43:39.452 D523.0    stand  None -310   490    0     80   0    
17:43:40.347 D523.0    stand  None -310   490    0     80   0    
17:43:41.346 D523.0    stand  None -310   490    0     80   0    
17:43:42.353 D523.0    stand  None -310   490    0     80   0    
17:43:43.356 D523.0    stand  None -310   490    0     80   0    
17:43:44.356 D523.0    stand  None -310   490    0     80   0    
17:43:45.365 D523.0    stand  None -310   490    0     80   0    
17:43:46.368 D523.0    stand  None -310   490    0     80   0    
17:43:47.360 D523.0    stand  None -310   490    0     80   0    
17:43:48.360 D523.0    stand  None -310   490    0     80   0    
17:43:49.361 D523.0    stand  None -310   490    0     80   0    
17:43:50.364 D523.0    stand  None -310   490    0     80   0    
17:43:51.259 D523.0    stand  None -310   490    0     80   0    
17:43:52.264 D523.0    stand  None -310   490    0     80   0    
17:43:53.261 D523.0    stand  None -290   500    0     80   22   
17:43:54.261 D523.0    stand  None -290   500    0     80   0    
17:43:55.262 D523.0    stand  None -290   500    0     80   0    
17:43:56.264 D523.0    stand  None -280   500    0     80   10   
17:43:57.268 D523.0    stand  None -280   500    0     80   0    
17:43:58.264 D523.0    stand  None -280   500    0     80   0    
17:43:59.266 D523.0    stand  None -280   500    0     80   0    
17:44:00.265 D523.0    stand  None -280   500    0     80   0    
17:44:01.274 D523.0    stand  None -280   500    0     80   0    
17:44:02.273 D523.0    stand  None -280   500    0     80   0    
17:44:03.164 D523.0    stand  None -280   500    0     80   0    
17:44:04.163 D523.0    stand  None -290   510    0     80   14   
17:44:05.164 D523.0    stand  None -290   510    0     80   0    
17:44:06.172 D523.0    stand  None -290   510    0     80   0    
17:44:07.168 D523.0    stand  None -290   510    0     80   0    
17:44:08.166 D523.0    stand  None -290   510    0     80   0    
17:44:09.168 D523.0    stand  None -290   510    0     80   0    
17:44:10.168 D523.0    stand  None -290   510    0     80   0    
17:44:11.170 D523.0    stand  None -290   510    0     80   0    
17:44:12.177 D523.0    stand  None -290   510    0     80   0    
17:44:13.176 D523.0    stand  None -290   510    0     80   0    
17:44:14.074 D523.0    stand  None -290   510    0     80   0    
17:44:15.077 D523.0    stand  None -290   510    0     80   0    
17:44:16.079 D523.0    stand  None -290   510    0     80   0    
17:44:17.077 D523.0    stand  None -290   510    0     80   0    
17:44:18.084 D523.0    stand  None -290   510    0     80   0    
17:44:19.089 D523.0    stand  None -290   510    0     80   0    
17:44:20.083 D523.0    stand  None -290   510    0     80   0    
17:44:21.081 D523.0    stand  None -290   510    0     80   0    
17:44:22.084 D523.0    stand  None -290   510    0     80   0    
17:44:23.083 D523.0    stand  None -290   510    0     80   0    
17:44:24.084 D523.0    stand  None -290   510    0     80   0    
17:44:25.093 D523.0    stand  None -290   510    0     80   0    
17:44:25.981 D523.0    stand  None -290   510    0     80   0    
17:44:26.982 D523.0    stand  None -290   510    0     80   0    
17:44:27.981 D523.0    stand  None -290   510    0     80   0    
17:44:28.982 D523.0    stand  None -290   510    0     80   0    
17:44:29.988 D523.0    stand  None -290   510    0     80   0    
17:44:31.000 D523.0    stand  None -290   510    0     80   0    
17:44:32.003 D523.0    stand  None -290   510    0     80   0    
17:44:33.000 D523.0    stand  None -290   510    0     80   0    
17:44:34.000 D523.0    stand  None -290   510    0     80   0    
17:44:35.011 D523.0    stand  None -290   510    0     80   0    
17:44:35.915 D523.0    stand  None -290   510    0     80   0    
17:44:36.908 D523.0    stand  None -290   510    0     80   0    
17:44:37.957 D523.0    stand  None -290   510    0     80   0    
17:44:38.904 D523.0    stand  None -290   510    0     80   0    
17:44:39.902 D523.0    stand  None -290   510    0     80   0    
17:44:40.905 D523.0    stand  None -290   510    0     80   0    
17:44:41.906 D523.0    stand  None -290   510    0     80   0    
17:44:42.911 D523.0    stand  None -290   510    0     80   0    
17:44:43.905 D523.0    stand  None -290   510    0     80   0    
17:44:44.914 D523.0    stand  None -290   510    0     80   0    
17:44:45.908 D523.0    stand  None -290   510    0     80   0    
17:44:46.813 D523.0    stand  None -290   510    0     80   0    
17:44:47.817 D523.0    stand  None -290   510    0     80   0    
17:44:48.815 D523.0    stand  None -290   510    0     80   0    
17:44:49.815 D523.0    stand  None -290   510    0     80   0    
17:44:50.814 D523.0    stand  None -290   510    0     80   0    
17:44:51.818 D523.0    stand  None -290   510    0     80   0    
17:44:52.819 D523.0    stand  None -290   510    0     80   0    
17:44:53.826 D523.0    stand  None -290   510    0     80   0    
17:44:54.817 D523.0    stand  None -290   510    0     80   0    
17:44:55.818 D523.0    stand  None -290   510    0     80   0    
17:44:56.824 D523.0    stand  None -290   510    0     80   0    
17:44:57.825 D523.0    stand  None -290   510    0     80   0    
17:44:58.717 D523.0    stand  None -290   510    0     80   0    
17:44:59.714 D523.0    stand  None -290   510    0     80   0    
17:45:00.718 D523.0    stand  None -290   510    0     80   0    
17:45:01.721 D523.0    stand  None -290   510    0     80   0    
17:45:02.682 D523.0    stand  None -290   510    0     80   0    
17:45:03.686 D523.0    stand  None -290   510    0     80   0    
17:45:04.688 D523.0    stand  None -290   510    0     80   0    
17:45:05.686 D523.0    stand  None -290   510    0     80   0    
17:45:06.686 D523.0    stand  None -290   500    0     80   10   
17:45:07.686 D523.0    stand  None -310   500    0     80   20   
17:45:08.688 D523.0    stand  None -310   500    0     80   0    
17:45:09.700 D523.0    stand  None -310   500    0     80   0    
17:45:10.690 D523.0    stand  None -310   500    0     80   0    
17:45:11.693 D523.0    stand  None -310   500    0     80   0    
17:45:12.693 D523.0    stand  None -310   500    0     80   0    
17:45:13.694 D523.0    stand  None -310   500    0     80   0    
17:45:14.593 D523.0    stand  None -310   500    0     80   0    
17:45:15.591 D523.0    stand  None -310   500    0     80   0    
17:45:16.587 D523.0    stand  None -310   500    0     80   0    
17:45:17.588 D523.0    stand  None -300   500    88    80   10   
17:45:18.639 D523.0    stand  None -300   500    0     80   0    
17:45:19.622 D523.0    stand  None -300   500    0     80   0    
17:45:20.631 D523.0    stand  None -300   500    0     80   0    
17:45:21.627 D523.0    stand  None -300   500    0     80   0    
17:45:22.527 D523.0    stand  None -300   500    0     80   0    
17:45:23.524 D523.0    stand  None -300   500    0     80   0    
17:45:24.529 D523.0    stand  None -300   500    0     80   0    
17:45:25.524 D523.0    stand  None -300   500    0     80   0    
17:45:26.524 D523.0    stand  None -300   500    0     80   0    
17:45:27.526 D523.0    stand  None -300   500    0     80   0    
17:45:28.530 D523.0    stand  None -300   500    0     80   0    
17:45:29.529 D523.0    stand  None -300   500    0     80   0    
17:45:30.538 D523.0    stand  None -300   500    0     80   0    
17:45:31.536 D523.0    stand  None -300   500    0     80   0    
17:45:32.538 D523.0    stand  None -300   500    0     80   0    
17:45:33.535 D523.0    stand  None -300   500    0     80   0    
17:45:34.431 D523.0    stand  None -300   500    0     80   0    
17:45:35.428 D523.0    stand  None -300   500    0     80   0    
17:45:36.484 D523.0    stand  None -300   500    0     80   0    
17:45:37.437 D523.0    stand  None -300   500    0     80   0    
17:45:38.437 D523.0    stand  None -300   500    0     80   0    
17:45:39.435 D523.0    stand  None -300   500    0     80   0    
17:45:40.435 D523.0    stand  None -300   500    0     80   0    
17:45:41.452 D523.0    stand  None -300   500    0     80   0    
17:45:42.436 D523.0    stand  None -300   500    0     80   0    
17:45:43.437 D523.0    stand  None -300   500    0     80   0    
17:45:44.436 D523.0    stand  None -300   500    0     80   0    
17:45:45.441 D523.0    stand  None -300   500    0     80   0    
17:45:46.335 D523.0    stand  None -300   500    0     80   0    
17:45:47.336 D523.0    stand  None -300   500    0     80   0    
17:45:48.337 D523.0    stand  None -300   500    0     80   0    
17:45:49.341 D523.0    stand  None -300   500    0     80   0    
17:45:50.341 D523.0    stand  None -300   500    0     80   0    
17:45:51.335 D523.0    stand  None -300   500    0     80   0    
17:45:52.338 D523.0    stand  None -300   500    0     80   0    
17:45:53.344 D523.0    stand  None -300   490    0     80   10   
17:45:54.344 D523.0    stand  None -300   500    0     80   10   
17:45:55.347 D523.0    stand  None -300   500    0     80   0    
17:45:56.343 D523.0    stand  None -300   500    0     80   0    
17:45:57.341 D523.0    stand  None -300   500    0     80   0    
17:45:58.246 D523.0    stand  None -300   500    0     80   0    
17:45:59.239 D523.0    stand  None -300   500    0     80   0    
17:46:00.235 D523.0    stand  None -300   500    0     80   0    
17:46:01.237 D523.0    stand  None -300   500    0     80   0    
17:46:02.237 D523.0    stand  None -300   500    0     80   0    
17:46:03.238 D523.0    stand  None -300   500    0     80   0    
17:46:04.242 D523.0    stand  None -300   500    0     80   0    
17:46:05.240 D523.0    stand  None -300   500    0     80   0    
17:46:06.256 D523.0    stand  None -300   500    0     80   0    
17:46:07.258 D523.0    stand  None -300   500    0     80   0    
17:46:08.153 D523.0    stand  None -300   500    0     80   0    
17:46:09.165 D523.0    stand  None -300   500    0     80   0    
17:46:10.157 D523.0    stand  None -300   500    0     80   0    
17:46:11.156 D523.0    stand  None -300   500    0     80   0    
17:46:12.162 D523.0    stand  None -300   500    0     80   0    
17:46:13.165 D523.0    stand  None -300   500    0     80   0    
17:46:14.160 D523.0    stand  None -300   500    0     80   0    
17:46:15.162 D523.0    stand  None -300   500    0     80   0    
17:46:16.162 D523.0    stand  None -300   500    0     80   0    
17:46:17.166 D523.0    stand  None -300   500    0     80   0    
17:46:18.165 D523.0    stand  None -300   500    0     80   0    
17:46:19.165 D523.0    stand  None -300   500    0     80   0    
17:46:20.067 D523.0    stand  None -300   500    0     80   0    
17:46:21.073 D523.0    stand  None -300   500    0     80   0    
17:46:22.069 D523.0    stand  None -300   500    0     80   0    
17:46:23.076 D523.0    stand  None -300   500    0     80   0    
17:46:24.075 D523.0    stand  None -300   500    0     80   0    
17:46:25.083 D523.0    stand  None -300   500    0     80   0    
17:46:26.086 D523.0    stand  None -300   500    0     80   0    
17:46:27.081 D523.0    stand  None -300   500    0     80   0    
17:46:28.080 D523.0    stand  None -300   500    0     80   0    
17:46:29.080 D523.0    stand  None -300   500    0     80   0    
17:46:29.984 D523.0    stand  None -300   500    0     80   0    
17:46:30.985 D523.0    stand  None -300   500    0     80   0    
17:46:31.980 D523.0    stand  None -300   500    0     80   0    
17:46:32.987 D523.0    stand  None -300   500    0     80   0    
17:46:33.992 D523.0    stand  None -300   500    0     80   0    
17:46:34.992 D523.0    stand  None -290   500    90    80   10   
17:46:36.054 D523.0    stand  None -310   500    0     80   20   
17:46:36.996 D523.0    stand  None -300   500    0     80   10   
17:46:38.013 D523.0    stand  None -300   500    0     80   0    
17:46:39.004 D523.0    stand  None -300   500    0     80   0    
17:46:39.999 D523.0    stand  None -300   500    0     80   0    
17:46:41.001 D523.0    stand  None -300   500    0     80   0    
17:46:41.902 D523.0    stand  None -300   500    0     80   0    
17:46:42.887 D523.0    stand  None -300   500    0     80   0    
17:46:43.886 D523.0    stand  None -300   500    0     80   0    
17:46:44.893 D523.0    stand  None -300   500    0     80   0    
17:46:45.891 D523.0    stand  None -300   500    0     80   0    
17:46:46.887 D523.0    stand  None -300   500    0     80   0    
17:46:47.888 D523.0    stand  None -300   500    0     80   0    
17:46:48.890 D523.0    stand  None -300   500    0     80   0    
17:46:49.890 D523.0    stand  None -300   500    0     80   0    
17:46:50.915 D523.0    stand  None -300   500    0     80   0    
17:46:51.901 D523.0    stand  None -300   500    0     80   0    
17:46:52.893 D523.0    stand  None -300   500    0     80   0    
17:46:53.786 D523.0    stand  None -300   500    0     80   0    
17:46:54.793 D523.0    stand  None -300   500    0     80   0    
17:46:55.788 D523.0    stand  None -300   500    0     80   0    
17:46:56.810 D523.0    stand  None -300   500    0     80   0    
17:46:57.792 D523.0    stand  None -300   500    0     80   0    
17:46:58.793 D523.0    stand  None -300   500    0     80   0    
17:46:59.799 D523.0    stand  None -300   500    0     80   0    
17:47:00.798 D523.0    stand  None -300   500    0     80   0    
17:47:01.796 D523.0    stand  None -300   500    0     80   0    
17:47:02.797 D523.0    stand  None -300   500    0     80   0    
17:47:03.798 D523.0    stand  None -300   500    0     80   0    
17:47:04.807 D523.0    stand  None -300   500    0     80   0    
17:47:05.700 D523.0    stand  None -300   500    0     80   0    
17:47:06.696 D523.0    stand  None -300   500    0     80   0    
17:47:07.693 D523.0    stand  None -300   500    0     80   0    
17:47:08.694 D523.0    stand  None -300   500    0     80   0    
17:47:09.696 D523.0    stand  None -300   500    0     80   0    
17:47:10.705 D523.0    stand  None -300   500    0     80   0    
17:47:11.702 D523.0    stand  None -300   500    0     80   0    
17:47:12.708 D523.0    stand  None -300   500    0     80   0    
17:47:13.708 D523.0    stand  None -300   500    0     80   0    
17:47:14.710 D523.0    stand  None -300   500    0     80   0    
17:47:15.706 D523.0    stand  None -300   500    0     80   0    
17:47:16.601 D523.0    stand  None -290   500    0     80   10   
17:47:17.604 D523.0    stand  None -290   500    0     80   0    
17:47:18.604 D523.0    stand  None -290   500    0     80   0    
17:47:19.604 D523.0    stand  None -290   500    0     80   0    
17:47:20.606 D523.0    stand  None -290   500    0     80   0    
17:47:21.610 D523.0    stand  None -290   500    0     80   0    
17:47:22.612 D523.0    stand  None -290   500    0     80   0    
17:47:23.610 D523.0    stand  None -290   500    0     80   0    
17:47:24.609 D523.0    stand  None -290   500    0     80   0    
17:47:25.614 D523.0    stand  None -290   500    0     80   0    
17:47:26.522 D523.0    stand  None -290   500    0     80   0    
17:47:27.525 D523.0    stand  None -290   500    0     80   0    
17:47:28.529 D523.0    stand  None -290   500    0     80   0    
17:47:29.524 D523.0    stand  None -290   500    0     80   0    
17:47:30.528 D523.0    stand  None -290   500    0     80   0    
17:47:31.538 D523.0    stand  None -290   500    0     80   0    
17:47:32.529 D523.0    stand  None -290   500    0     80   0    
17:47:33.537 D523.0    stand  None -290   500    0     80   0    
17:47:34.529 D523.0    stand  None -290   500    0     80   0    
17:47:35.588 D523.0    stand  None -290   500    0     80   0    
17:47:36.532 D523.0    stand  None -290   500    0     80   0    
17:47:37.566 D523.0    stand  None -290   500    0     80   0    
17:47:38.428 D523.0    stand  None -290   500    0     80   0    
17:47:39.434 D523.0    stand  None -290   500    0     80   0    
17:47:40.428 D523.0    stand  None -290   500    0     80   0    
17:47:41.428 D523.0    stand  None -290   500    0     80   0    
17:47:42.393 D523.0    stand  None -290   500    0     80   0    
17:47:43.399 D523.0    stand  None -290   500    0     80   0    
17:47:44.395 D523.0    stand  None -290   500    0     80   0    
17:47:45.401 D523.0    stand  None -290   500    0     80   0    
17:47:46.424 D523.0    stand  None -290   500    0     80   0    
17:47:47.432 D523.0    stand  None -290   500    0     80   0    
17:47:48.407 D523.0    stand  None -290   500    0     80   0    
17:47:49.407 D523.0    stand  None -290   500    0     80   0    
17:47:50.404 D523.0    stand  None -290   500    0     80   0    
17:47:51.412 D523.0    stand  None -290   500    0     80   0    
17:47:52.404 D523.0    stand  None -290   500    0     80   0    
17:47:53.406 D523.0    stand  None -290   500    0     80   0    
17:47:54.304 D523.0    stand  None -290   500    0     80   0    
17:47:55.303 D523.0    stand  None -290   500    0     80   0    
17:47:56.303 D523.0    stand  None -290   500    0     80   0    
17:47:57.303 D523.0    stand  None -290   500    0     80   0    
17:47:58.339 D523.0    stand  None -290   500    0     80   0    
17:47:59.340 D523.0    stand  None -290   500    0     80   0    
17:48:00.344 D523.0    stand  None -300   510    0     80   14   
17:48:01.339 D523.0    stand  None -290   500    85    80   14   
17:48:02.238 D523.0    stand  None -270   470    105   80   36   
17:48:03.239 D523.0    walk   None -240   430    106   80   50   
17:48:04.245 D523.0    walk   None -260   340    114   80   92   
17:48:05.241 D523.0    walk   None -290   250    106   80   94   
17:48:06.242 D523.0    walk   None -330   160    109   80   98   
17:48:07.242 D523.0    walk   None -350   120    98    80   44   
17:48:08.247 D523.0    walk   None -390   110    107   80   41   
17:48:09.248 D523.0    walk   None -360   120    96    80   31   
17:48:10.244 D523.0    walk   None -360   210    106   80   90   
17:48:11.248 D523.0    walk   None -300   310    102   80   116  
17:48:12.247 D523.0    walk   None -250   410    100   80   111  
17:48:13.247 D523.0    walk   None -280   490    95    80   85   
17:48:14.146 D523.0    walk   None -290   510    0     80   22   
17:48:15.152 D523.0    walk   None -290   520    0     80   10   
17:48:16.152 D523.0    stand  None -280   490    0     80   31   
17:48:17.155 D523.0    stand  None -280   490    0     80   0    
17:48:18.156 D523.0    stand  None -280   490    0     80   0    
17:48:19.156 D523.0    stand  None -280   490    0     80   0    
17:48:20.158 D523.0    stand  None -280   490    0     80   0    
17:48:21.160 D523.0    stand  None -280   490    0     80   0    
17:48:22.160 D523.0    stand  None -280   490    0     80   0    
17:48:23.166 D523.0    stand  None -280   490    0     80   0    
17:48:24.063 D523.0    stand  None -300   480    96    80   22   
17:48:25.060 D523.0    stand  None -310   480    0     80   10   
17:48:26.064 D523.0    stand  None -310   480    0     80   0    
17:48:27.065 D523.0    stand  None -310   480    0     80   0    
17:48:28.064 D523.0    stand  None -310   480    0     80   0    
17:48:29.089 D523.0    stand  None -310   480    0     80   0    
17:48:30.076 D523.0    stand  None -300   490    0     80   14   
17:48:31.002 D523.0    stand  None -290   510    0     80   22   
17:48:32.002 D523.0    stand  None -290   510    0     80   0    
17:48:33.008 D523.0    stand  None -290   510    0     80   0    
17:48:34.010 D523.0    stand  None -290   510    0     80   0    
17:48:35.060 D523.0    stand  None -290   510    0     80   0    
17:48:36.008 D523.0    stand  None -290   510    0     80   0    
17:48:37.009 D523.0    stand  None -290   510    0     80   0    
17:48:38.008 D523.0    stand  None -290   510    0     80   0    
17:48:39.012 D523.0    stand  None -290   510    0     80   0    
17:48:40.015 D523.0    stand  None -290   510    0     80   0    
17:48:41.012 D523.0    stand  None -290   510    0     80   0    
17:48:42.018 D523.0    stand  None -290   510    0     80   0    
17:48:42.906 D523.0    stand  None -290   510    0     80   0    
17:48:43.907 D523.0    stand  None -290   510    0     80   0    
17:48:44.908 D523.0    stand  None -290   510    0     80   0    
17:48:45.912 D523.0    stand  None -290   510    0     80   0    
17:48:46.970 D523.0    stand  None -290   510    0     80   0    
17:48:47.866 D523.0    stand  None -290   510    0     80   0    
17:48:48.865 D523.0    stand  None -290   510    0     80   0    
17:48:49.873 D523.0    stand  None -290   510    0     80   0    
17:48:50.874 D523.0    stand  None -290   510    0     80   0    
17:48:51.874 D523.0    stand  None -290   510    0     80   0    
17:48:52.874 D523.0    stand  None -290   510    0     80   0    
17:48:53.883 D523.0    stand  None -290   510    0     80   0    
17:48:54.881 D523.0    stand  None -290   510    0     80   0    
17:48:55.896 D523.0    stand  None -290   490    0     80   20   
17:48:56.875 D523.0    stand  None -280   490    0     80   10   
17:48:57.876 D523.0    stand  None -280   480    0     80   10   
17:48:58.884 D523.0    stand  None -280   480    0     80   0    
17:48:59.770 D523.0    stand  None -280   480    0     80   0    
17:49:00.771 D523.0    stand  None -280   480    0     80   0    
17:49:01.773 D523.0    stand  None -280   480    0     80   0    
17:49:02.784 D523.0    stand  None -280   480    0     80   0    
17:49:03.812 D523.0    stand  None -280   480    0     80   0    
17:49:04.783 D523.0    stand  None -280   480    0     80   0    
17:49:05.784 D523.0    stand  None -280   480    0     80   0    
17:49:06.785 D523.0    stand  None -280   480    0     80   0    
17:49:07.785 D523.0    stand  None -280   480    0     80   0    
17:49:08.788 D523.0    stand  None -280   480    0     80   0    
17:49:09.786 D523.0    stand  None -280   480    0     80   0    
17:49:10.686 D523.0    stand  None -280   480    0     80   0    
17:49:11.689 D523.0    stand  None -280   480    0     80   0    
17:49:12.685 D523.0    stand  None -280   480    0     80   0    
17:49:13.687 D523.0    stand  None -280   480    0     80   0    
17:49:14.688 D523.0    stand  None -280   480    0     80   0    
17:49:15.691 D523.0    stand  None -280   480    0     80   0    
17:49:16.692 D523.0    stand  None -280   480    0     80   0    
17:49:17.692 D523.0    stand  None -280   480    0     80   0    
17:49:18.692 D523.0    stand  None -280   480    0     80   0    
17:49:19.696 D523.0    stand  None -280   480    0     80   0    
17:49:20.696 D523.0    stand  None -280   480    0     80   0    
17:49:21.694 D523.0    stand  None -280   480    0     80   0    
17:49:22.592 D523.0    stand  None -280   480    0     80   0    
17:49:23.602 D523.0    stand  None -280   480    0     80   0    
17:49:24.594 D523.0    stand  None -280   480    0     80   0    
17:49:25.594 D523.0    stand  None -280   480    0     80   0    
17:49:26.590 D523.0    stand  None -280   480    0     80   0    
17:49:27.592 D523.0    stand  None -280   480    0     80   0    
17:49:28.600 D523.0    stand  None -280   480    0     80   0    
17:49:29.595 D523.0    stand  None -280   480    0     80   0    
17:49:30.595 D523.0    stand  None -290   500    0     80   22   
17:49:31.594 D523.0    stand  None -290   500    0     80   0    
17:49:32.601 D523.0    stand  None -290   500    0     80   0    
17:49:33.606 D523.0    stand  None -290   500    0     80   0    
17:49:34.493 D523.0    stand  None -290   500    0     80   0    
17:49:35.540 D523.0    stand  None -290   500    0     80   0    
17:49:36.499 D523.0    stand  None -290   490    0     80   10   
17:49:37.492 D523.0    stand  None -290   490    0     80   0    
17:49:38.493 D523.0    stand  None -290   490    0     80   0    
17:49:39.495 D523.0    stand  None -290   490    0     80   0    
17:49:40.497 D523.0    stand  None -290   490    0     80   0    
17:49:41.496 D523.0    stand  None -290   490    0     80   0    
17:49:42.500 D523.0    stand  None -290   490    0     80   0    
17:49:43.510 D523.0    stand  None -290   490    0     80   0    
17:49:44.501 D523.0    stand  None -290   490    0     80   0    
17:49:45.500 D523.0    stand  None -290   490    0     80   0    
17:49:46.396 D523.0    stand  None -290   490    0     80   0    
17:49:47.395 D523.0    stand  None -290   490    0     80   0    
17:49:48.396 D523.0    stand  None -290   490    0     80   0    
17:49:49.398 D523.0    stand  None -290   490    0     80   0    
17:49:50.417 D523.0    stand  None -290   490    0     80   0    
17:49:51.419 D523.0    stand  None -290   490    0     80   0    
17:49:52.416 D523.0    stand  None -290   490    0     80   0    
17:49:53.418 D523.0    stand  None -290   490    0     80   0    
17:49:54.417 D523.0    stand  None -290   490    0     80   0    
17:49:55.418 D523.0    stand  None -290   490    0     80   0    
17:49:56.320 D523.0    stand  None -290   490    0     80   0    
17:49:57.316 D523.0    stand  None -290   490    0     80   0    
17:49:58.317 D523.0    stand  None -290   490    0     80   0    
17:49:59.317 D523.0    stand  None -290   490    0     80   0    
17:50:00.324 D523.0    stand  None -290   490    0     80   0    
17:50:01.322 D523.0    stand  None -290   490    0     80   0    
17:50:02.324 D523.0    stand  None -290   490    0     80   0    
17:50:03.320 D523.0    stand  None -290   490    0     80   0    
17:50:04.332 D523.0    stand  None -290   490    0     80   0    
17:50:05.325 D523.0    stand  None -290   490    0     80   0    
17:50:06.330 D523.0    stand  None -290   490    0     80   0    
17:50:07.328 D523.0    stand  None -290   490    0     80   0    
17:50:08.233 D523.0    stand  None -290   490    0     80   0    
17:50:09.241 D523.0    stand  None -290   490    0     80   0    
17:50:10.236 D523.0    stand  None -290   490    0     80   0    
17:50:11.226 D523.0    stand  None -290   490    0     80   0    
17:50:12.229 D523.0    stand  None -290   490    0     80   0    
17:50:13.238 D523.0    stand  None -290   490    0     80   0    
17:50:14.227 D523.0    stand  None -290   490    0     80   0    
17:50:15.225 D523.0    stand  None -290   490    0     80   0    
17:50:16.229 D523.0    stand  None -290   490    0     80   0    
17:50:17.234 D523.0    stand  None -290   490    0     80   0    
17:50:18.228 D523.0    stand  None -290   490    0     80   0    
17:50:19.228 D523.0    stand  None -290   490    0     80   0    
17:50:20.156 D523.0    stand  None -290   490    0     80   0    
17:50:21.127 D523.0    stand  None -290   490    0     80   0    
17:50:22.125 D523.0    stand  None -290   490    0     80   0    
17:50:23.126 D523.0    stand  None -290   490    0     80   0    
17:50:24.126 D523.0    stand  None -290   490    0     80   0    
17:50:25.136 D523.0    stand  None -290   490    0     80   0    
17:50:26.131 D523.0    stand  None -290   490    0     80   0    
17:50:27.144 D523.0    stand  None -290   490    0     80   0    
17:50:28.130 D523.0    stand  None -290   490    0     80   0    
17:50:29.132 D523.0    stand  None -290   490    0     80   0    
17:50:30.132 D523.0    stand  None -290   490    0     80   0    
17:50:31.136 D523.0    stand  None -290   490    0     80   0    
17:50:32.033 D523.0    stand  None -290   490    0     80   0    
17:50:33.030 D523.0    stand  None -290   490    0     80   0    
17:50:34.080 D523.0    stand  None -290   490    0     80   0    
17:50:35.033 D523.0    stand  None -290   490    0     80   0    
17:50:36.030 D523.0    stand  None -290   490    0     80   0    
17:50:37.032 D523.0    stand  None -290   490    0     80   0    
17:50:38.032 D523.0    stand  None -290   490    0     80   0    
17:50:39.046 D523.0    stand  None -290   490    0     80   0    
17:50:40.048 D523.0    stand  None -290   490    0     80   0    
17:50:41.049 D523.0    stand  None -290   490    0     80   0    
17:50:41.956 D523.0    stand  None -290   490    0     80   0    
17:50:42.948 D523.0    stand  None -290   490    0     80   0    
17:50:43.958 D523.0    stand  None -290   490    0     80   0    
17:50:44.957 D523.0    stand  None -290   490    0     80   0    
17:50:45.951 D523.0    stand  None -290   490    0     80   0    
17:50:46.956 D523.0    stand  None -290   490    0     80   0    
17:50:47.958 D523.0    stand  None -290   490    0     80   0    
17:50:48.958 D523.0    stand  None -290   490    0     80   0    
17:50:49.965 D523.0    stand  None -290   490    0     80   0    
17:50:50.958 D523.0    stand  None -290   490    0     80   0    
17:50:51.957 D523.0    stand  None -290   490    0     80   0    
17:50:52.956 D523.0    stand  None -290   490    0     80   0    
17:50:53.853 D523.0    stand  None -290   490    0     80   0    
17:50:54.861 D523.0    stand  None -290   490    0     80   0    
17:50:55.864 D523.0    stand  None -290   490    0     80   0    
17:50:56.860 D523.0    stand  None -290   490    0     80   0    
17:50:57.865 D523.0    stand  None -290   490    0     80   0    
17:50:58.864 D523.0    stand  None -290   490    0     80   0    
17:50:59.867 D523.0    stand  None -290   490    0     80   0    
17:51:00.865 D523.0    stand  None -290   490    0     80   0    
17:51:01.873 D523.0    stand  None -290   490    0     80   0    
17:51:02.868 D523.0    stand  None -290   490    0     80   0    
17:51:03.868 D523.0    stand  None -290   490    0     80   0    
17:51:04.764 D523.0    stand  None -290   490    0     80   0    
17:51:05.764 D523.0    stand  None -290   490    0     80   0    
17:51:06.775 D523.0    stand  None -290   490    0     80   0    
17:51:07.768 D523.0    stand  None -290   490    0     80   0    
17:51:08.766 D523.0    stand  None -290   490    0     80   0    
17:51:09.776 D523.0    stand  None -290   490    0     80   0    
17:51:10.771 D523.0    stand  None -290   490    0     80   0    
17:51:11.784 D523.0    stand  None -290   490    0     80   0    
17:51:12.772 D523.0    stand  None -290   490    0     80   0    
17:51:13.772 D523.0    stand  None -290   490    0     80   0    
17:51:14.778 D523.0    stand  None -290   490    0     80   0    
17:51:15.776 D523.0    stand  None -290   490    0     80   0    
17:51:16.665 D523.0    stand  None -290   490    0     80   0    
17:51:17.668 D523.0    stand  None -290   490    0     80   0    
17:51:18.668 D523.0    stand  None -290   490    0     80   0    
17:51:19.678 D523.0    stand  None -290   490    0     80   0    
17:51:20.676 D523.0    stand  None -290   490    0     80   0    
17:51:21.676 D523.0    stand  None -290   490    0     80   0    
17:51:22.672 D523.0    stand  None -290   490    0     80   0    
17:51:23.672 D523.0    stand  None -290   490    0     80   0    
17:51:24.682 D523.0    stand  None -290   490    0     80   0    
17:51:25.677 D523.0    stand  None -290   490    0     80   0    
17:51:26.683 D523.0    stand  None -290   490    0     80   0    
17:51:27.581 D523.0    stand  None -290   490    0     80   0    
17:51:28.584 D523.0    stand  None -290   490    0     80   0    
17:51:29.583 D523.0    stand  None -290   490    0     80   0    
17:51:30.594 D523.0    stand  None -290   490    0     80   0    
17:51:31.593 D523.0    stand  None -290   490    0     80   0    
17:51:32.590 D523.0    stand  None -290   490    0     80   0    
17:51:33.642 D523.0    stand  None -290   490    0     80   0    
17:51:34.587 D523.0    stand  None -290   490    0     80   0    
17:51:35.585 D523.0    stand  None -290   490    0     80   0    
17:51:36.588 D523.0    stand  None -290   490    0     80   0    
17:51:37.595 D523.0    stand  None -290   490    0     80   0    
17:51:38.590 D523.0    stand  None -290   490    0     80   0    
17:51:39.485 D523.0    stand  None -290   490    0     80   0    
17:51:40.486 D523.0    stand  None -290   490    0     80   0    
17:51:41.484 D523.0    stand  None -290   490    0     80   0    
17:51:42.494 D523.0    stand  None -290   490    0     80   0    
17:51:43.494 D523.0    stand  None -290   490    0     80   0    
17:51:44.494 D523.0    stand  None -290   490    0     80   0    
17:51:45.495 D523.0    stand  None -290   490    0     80   0    
17:51:46.513 D523.0    stand  None -290   490    0     80   0    
17:51:47.496 D523.0    stand  None -290   490    0     80   0    
17:51:48.498 D523.0    stand  None -290   490    0     80   0    
17:51:49.498 D523.0    stand  None -290   490    0     80   0    
17:51:50.394 D523.0    stand  None -290   490    0     80   0    
17:51:51.398 D523.0    stand  None -290   490    0     80   0    
17:51:52.401 D523.0    stand  None -290   490    0     80   0    
17:51:53.402 D523.0    stand  None -290   490    0     80   0    
17:51:54.404 D523.0    stand  None -290   490    0     80   0    
17:51:55.400 D523.0    stand  None -290   490    0     80   0    
17:51:56.402 D523.0    stand  None -290   490    0     80   0    
17:51:57.408 D523.0    stand  None -290   490    0     80   0    
17:51:58.408 D523.0    stand  None -290   490    0     80   0    
17:51:59.411 D523.0    stand  None -290   490    0     80   0    
17:52:00.404 D523.0    stand  None -290   490    0     80   0    
17:52:01.410 D523.0    stand  None -290   490    0     80   0    
17:52:02.300 D523.0    stand  None -290   490    0     80   0    
17:52:03.301 D523.0    stand  None -290   490    0     80   0    
17:52:04.300 D523.0    stand  None -290   490    0     80   0    
17:52:05.303 D523.0    stand  None -300   510    0     80   22   
17:52:06.302 D523.0    stand  None -300   510    0     80   0    
17:52:07.308 D523.0    stand  None -300   510    0     80   0    
17:52:08.307 D523.0    stand  None -300   510    0     80   0    
17:52:09.312 D523.0    stand  None -300   510    0     80   0    
17:52:10.309 D523.0    stand  None -300   510    0     80   0    
17:52:11.309 D523.0    stand  None -300   510    0     80   0    
17:52:12.308 D523.0    stand  None -300   510    0     80   0    
17:52:13.309 D523.0    stand  None -300   510    0     80   0    
17:52:14.202 D523.0    stand  None -300   510    0     80   0    
17:52:15.205 D523.0    stand  None -300   510    0     80   0    
17:52:16.207 D523.0    stand  None -300   510    0     80   0    
17:52:17.208 D523.0    stand  None -300   510    0     80   0    
17:52:18.208 D523.0    stand  None -300   490    0     80   20   
17:52:19.212 D523.0    stand  None -300   490    0     80   0    
17:52:20.209 D523.0    stand  None -300   490    0     80   0    
17:52:21.215 D523.0    stand  None -300   490    0     80   0    
17:52:22.214 D523.0    stand  None -300   490    0     80   0    
17:52:23.214 D523.0    stand  None -300   490    0     80   0    
17:52:24.213 D523.0    stand  None -300   490    0     80   0    
17:52:25.218 D523.0    stand  None -300   490    0     80   0    
17:52:26.112 D523.0    stand  None -300   490    0     80   0    
17:52:27.110 D523.0    stand  None -300   490    0     80   0    
17:52:28.112 D523.0    stand  None -300   490    0     80   0    
17:52:29.110 D523.0    stand  None -300   490    0     80   0    
17:52:30.110 D523.0    stand  None -300   490    0     80   0    
17:52:31.129 D523.0    stand  None -300   490    0     80   0    
17:52:32.126 D523.0    stand  None -300   490    0     80   0    
17:52:33.177 D523.0    stand  None -300   490    0     80   0    
17:52:34.128 D523.0    stand  None -300   490    0     80   0    
17:52:35.134 D523.0    stand  None -300   490    0     80   0    
17:52:36.026 D523.0    stand  None -300   490    0     80   0    
17:52:37.029 D523.0    stand  None -300   490    0     80   0    
17:52:38.033 D523.0    stand  None -300   490    0     80   0    
17:52:39.032 D523.0    stand  None -300   490    0     80   0    
17:52:40.033 D523.0    stand  None -300   490    0     80   0    
17:52:41.035 D523.0    stand  None -300   490    0     80   0    
17:52:42.035 D523.0    stand  None -300   490    0     80   0    
17:52:43.035 D523.0    stand  None -300   490    0     80   0    
17:52:44.036 D523.0    stand  None -300   490    0     80   0    
17:52:45.036 D523.0    stand  None -300   490    0     80   0    
17:52:46.036 D523.0    stand  None -300   490    0     80   0    
17:52:46.938 D523.0    stand  None -300   490    0     80   0    
17:52:47.940 D523.0    stand  None -300   490    0     80   0    
17:52:48.941 D523.0    stand  None -300   490    0     80   0    
17:52:49.942 D523.0    stand  None -300   490    0     80   0    
17:52:50.944 D523.0    stand  None -300   490    0     80   0    
17:52:51.944 D523.0    stand  None -300   490    0     80   0    
17:52:52.946 D523.0    stand  None -300   490    0     80   0    
17:52:53.949 D523.0    stand  None -300   490    0     80   0    
17:52:54.951 D523.0    stand  None -300   490    0     80   0    
17:52:55.955 D523.0    stand  None -300   490    0     80   0    
17:52:56.953 D523.0    stand  None -300   490    0     80   0    
17:52:57.954 D523.0    stand  None -300   490    0     80   0    
17:52:58.849 D523.0    stand  None -300   490    0     80   0    

17:43:00.539 D5F7.88   88     -    -      -      -     -    -    
17:43:32.298 D5F7.88   88     -    -      -      -     -    -    
17:44:03.941 D5F7.88   88     -    -      -      -     -    -    
17:44:35.855 D5F7.88   88     -    -      -      -     -    -    
17:45:07.426 D5F7.88   88     -    -      -      -     -    -    
17:45:23.166 D5F7.0    stand  None -60    60     83    80        
17:45:23.303 D5F7.0    stand  None -40    50     72    80   22   
17:45:24.315 D5F7.0    walk   None 0      10     90    80   56   
17:45:25.326 D5F7.0    walk   None 10     10     80    80   10   
17:45:26.342 D5F7.0    walk   None 20     0      70    80   14   
17:45:27.352 D5F7.0    walk   None 30     0      57    80   10   
17:45:28.258 D5F7.0    walk   None 30     10     68    80   10   
17:45:29.277 D5F7.0    walk   None 30     30     0     80   20   
17:45:30.284 D5F7.0    sit    None 30     30     0     80   0    
17:45:31.300 D5F7.0    sit    None 30     30     0     80   0    
17:45:32.309 D5F7.0    sit    None 40     30     0     80   10   
17:45:33.221 D5F7.0    sit    None 40     30     0     80   0    
17:45:34.236 D5F7.0    sit    None 40     30     0     80   0    
17:45:35.248 D5F7.0    sit    None 40     30     0     80   0    
17:45:36.259 D5F7.0    sit    None 40     30     0     80   0    
17:45:37.271 D5F7.0    sit    None 40     30     0     80   0    
17:45:38.299 D5F7.0    sit    None 40     30     0     80   0    
17:45:39.169 D5F7.0    sit    None 40     30     0     80   0    
17:45:40.183 D5F7.0    sit    None 40     30     0     80   0    
17:45:41.196 D5F7.0    sit    None 40     30     0     80   0    
17:45:42.233 D5F7.0    sit    None 40     30     0     80   0    
17:45:43.222 D5F7.0    sit    None 40     30     0     80   0    
17:45:44.134 D5F7.0    sit    None 40     30     0     80   0    
17:45:45.145 D5F7.0    sit    None 40     30     0     80   0    
17:45:46.169 D5F7.0    sit    None 40     30     0     80   0    
17:45:47.171 D5F7.0    sit    None 40     30     0     80   0    
17:45:48.184 D5F7.0    sit    None 40     30     0     80   0    
17:45:49.095 D5F7.0    sit    None 30     30     70    80   10   
17:45:50.103 D5F7.0    sit    None 80     10     75    80   53   
17:45:51.123 D5F7.0    sit    None 90     10     60    80   10   
17:45:52.132 D5F7.0    sit    None 100    10     79    80   10   
17:45:53.143 D5F7.0    sit    None 40     30     0     80   63   
17:45:54.049 D5F7.0    sit    None 40     30     0     80   0    
17:45:55.067 D5F7.0    sit    None 40     30     0     80   0    
17:45:56.076 D5F7.0    sit    None 40     30     0     80   0    
17:45:57.090 D5F7.0    sit    None 40     30     0     80   0    
17:45:58.101 D5F7.0    sit    None 40     30     0     80   0    
17:45:59.013 D5F7.0    sit    None 40     30     0     80   0    
17:46:00.025 D5F7.0    sit    None 40     30     0     80   0    
17:46:01.040 D5F7.0    sit    None 40     30     0     80   0    
17:46:02.046 D5F7.0    sit    None 40     30     0     80   0    
17:46:03.062 D5F7.0    sit    None 40     30     0     80   0    
17:46:03.983 D5F7.0    sit    None 40     30     0     80   0    
17:46:04.981 D5F7.0    sit    None 40     30     0     80   0    
17:46:05.993 D5F7.0    sit    None 40     30     0     80   0    
17:46:07.007 D5F7.0    sit    None 40     30     0     80   0    
17:46:08.022 D5F7.0    sit    None 40     30     0     80   0    
17:46:08.929 D5F7.0    sit    None 40     30     0     80   0    
17:46:09.943 D5F7.0    sit    None 40     30     0     80   0    
17:46:10.913 D5F7.0    sit    None 40     30     0     80   0    
17:46:11.929 D5F7.0    sit    None 40     30     0     80   0    
17:46:12.943 D5F7.0    sit    None 40     30     0     80   0    
17:46:13.952 D5F7.0    sit    None 40     30     0     80   0    
17:46:14.967 D5F7.0    sit    None 40     30     0     80   0    
17:46:15.872 D5F7.0    sit    None 40     30     0     80   0    
17:46:16.887 D5F7.0    sit    None 40     30     0     80   0    
17:46:17.899 D5F7.0    sit    None 40     30     0     80   0    
17:46:18.916 D5F7.0    sit    None 40     30     0     80   0    
17:46:19.929 D5F7.0    sit    None 40     30     0     80   0    
17:46:20.833 D5F7.0    sit    None 40     30     0     80   0    
17:46:21.925 D5F7.0    sit    None 40     30     0     80   0    
17:46:22.882 D5F7.0    sit    None 40     30     55    80   0    
17:46:23.891 D5F7.0    sit    None 40     30     0     80   0    
17:46:24.800 D5F7.0    sit    None 40     30     0     80   0    
17:46:25.814 D5F7.0    sit    None 40     30     0     80   0    
17:46:26.828 D5F7.0    sit    None 40     30     0     80   0    
17:46:27.904 D5F7.0    sit    None 40     30     0     80   0    
17:46:27.904 D5F7.1    stand  255  160    0      103   80   123  
17:46:28.772 D5F7.0    sit    None 40     30     0     80   123  
17:46:28.772 D5F7.1    stand  255  160    0      0     80   123  
17:46:29.785 D5F7.0    sit    None 40     30     0     80   123  
17:46:29.785 D5F7.1    stand  255  160    0      0     80   123  
17:46:30.824 D5F7.1    stand  255  160    0      0     80   0    
17:46:30.824 D5F7.0    sit    None 40     30     0     80   123  
17:46:31.820 D5F7.0    sit    None 40     30     0     80   0    
17:46:31.820 D5F7.1    stand  255  160    0      0     80   123  
17:46:32.824 D5F7.1    stand  255  160    0      0     80   0    
17:46:32.824 D5F7.0    sit    None 40     30     0     80   123  
17:46:33.728 D5F7.0    sit    None 40     30     0     80   0    
17:46:33.728 D5F7.1    stand  255  160    0      0     80   123  
17:46:34.745 D5F7.1    stand  255  160    0      0     80   0    
17:46:34.745 D5F7.0    sit    None 40     30     0     80   123  
17:46:35.757 D5F7.0    sit    None 40     30     0     80   0    
17:46:35.757 D5F7.1    stand  255  160    0      0     80   123  
17:46:36.806 D5F7.0    sit    None 40     30     0     80   123  
17:46:36.806 D5F7.1    stand  255  160    0      0     80   123  
17:46:37.781 D5F7.0    sit    None 40     30     0     80   123  
17:46:37.781 D5F7.1    stand  255  160    0      0     80   123  
17:46:38.715 D5F7.1    stand  255  160    0      0     80   0    
17:46:38.715 D5F7.0    sit    None 40     30     0     80   123  
17:46:39.711 D5F7.1    stand  255  160    0      0     80   123  
17:46:39.711 D5F7.0    sit    None 40     30     0     80   123  
17:46:40.721 D5F7.1    stand  255  160    0      0     80   123  
17:46:40.721 D5F7.0    sit    None 40     30     0     80   123  
17:46:41.733 D5F7.0    sit    None 40     30     65    80   0    
17:46:41.733 D5F7.1    stand  255  150    10     0     80   111  
17:46:42.688 D5F7.0    sit    None 30     20     77    80   120  
17:46:42.688 D5F7.1    stand  255  150    0      0     80   121  
17:46:43.687 D5F7.0    sit    None 30     30     0     80   123  
17:46:43.687 D5F7.1    stand  255  150    20     0     80   120  
17:46:44.693 D5F7.0    sit    None 30     30     0     80   120  
17:46:44.693 D5F7.1    stand  255  150    20     0     80   120  
17:46:45.700 D5F7.0    sit    None 30     30     0     80   120  
17:46:45.700 D5F7.1    stand  255  150    20     0     80   120  
17:46:46.715 D5F7.0    sit    None 30     30     0     80   120  
17:46:46.715 D5F7.1    stand  255  160    0      0     80   133  
17:46:47.616 D5F7.0    sit    None 30     30     0     80   133  
17:46:47.616 D5F7.1    stand  255  160    0      0     80   133  
17:46:48.642 D5F7.0    sit    None 30     30     0     80   133  
17:46:48.642 D5F7.1    stand  255  150    0      0     80   123  
17:46:49.641 D5F7.1    stand  255  150    0      0     80   0    
17:46:49.641 D5F7.0    sit    None 30     30     0     80   123  
17:46:50.657 D5F7.0    sit    None 30     30     0     80   0    
17:46:50.657 D5F7.1    stand  255  150    0      0     80   123  
17:46:51.673 D5F7.0    sit    None 30     30     0     80   123  
17:46:51.673 D5F7.1    stand  255  150    0      0     80   123  
17:46:52.577 D5F7.1    stand  255  150    0      0     80   0    
17:46:52.577 D5F7.0    sit    None 30     30     0     80   123  
17:46:53.590 D5F7.1    stand  255  150    0      0     80   123  
17:46:53.590 D5F7.0    sit    None 30     30     0     80   123  
17:46:54.604 D5F7.0    sit    None 30     30     0     80   0    
17:46:54.604 D5F7.1    stand  255  150    0      0     80   123  
17:46:55.618 D5F7.0    sit    None 30     30     0     80   123  
17:46:55.618 D5F7.1    stand  255  150    0      0     80   123  
17:46:56.629 D5F7.0    sit    None 30     30     0     80   123  
17:46:56.629 D5F7.1    stand  255  150    0      0     80   123  
17:46:57.546 D5F7.0    sit    None 30     30     0     80   123  
17:46:57.546 D5F7.1    stand  255  150    0      0     80   123  
17:46:58.529 D5F7.0    sit    None 30     30     0     80   123  
17:46:58.529 D5F7.1    stand  255  150    0      0     80   123  
17:46:59.541 D5F7.0    sit    None 30     30     0     80   123  
17:46:59.541 D5F7.1    stand  255  150    0      0     80   123  
17:47:00.562 D5F7.0    sit    None 30     30     0     80   123  
17:47:00.562 D5F7.1    stand  255  150    0      0     80   123  
17:47:01.578 D5F7.0    sit    None 30     30     0     80   123  
17:47:01.578 D5F7.1    stand  255  150    0      0     80   123  
17:47:02.585 D5F7.0    sit    None 30     30     0     80   123  
17:47:02.585 D5F7.1    stand  255  150    0      0     80   123  
17:47:03.492 D5F7.0    sit    None 30     30     0     80   123  
17:47:03.492 D5F7.1    stand  255  150    0      0     80   123  
17:47:04.500 D5F7.1    stand  255  150    0      0     80   0    
17:47:04.500 D5F7.0    sit    None 30     30     0     80   123  
17:47:05.524 D5F7.0    sit    None 30     30     0     80   0    
17:47:05.524 D5F7.1    stand  255  150    0      0     80   123  
17:47:06.537 D5F7.0    sit    None 30     30     0     80   123  
17:47:06.537 D5F7.1    stand  255  150    0      0     80   123  
17:47:07.548 D5F7.0    sit    None 30     30     0     80   123  
17:47:07.548 D5F7.1    stand  255  150    0      0     80   123  
17:47:08.456 D5F7.0    sit    None 30     30     0     80   123  
17:47:08.456 D5F7.1    stand  255  150    0      0     80   123  
17:47:09.465 D5F7.0    sit    None 30     30     0     80   123  
17:47:09.465 D5F7.1    stand  255  150    0      0     80   123  
17:47:10.473 D5F7.0    sit    None 30     30     0     80   123  
17:47:10.473 D5F7.1    stand  255  150    0      0     80   123  
17:47:11.493 D5F7.1    stand  255  150    0      0     80   0    
17:47:11.493 D5F7.0    sit    None 30     30     0     80   123  
17:47:12.500 D5F7.1    stand  255  150    0      0     80   123  
17:47:12.500 D5F7.0    sit    None 30     30     0     80   123  
17:47:13.409 D5F7.0    sit    None 30     30     0     80   0    
17:47:13.409 D5F7.1    stand  255  150    0      0     80   123  
17:47:14.449 D5F7.0    sit    None 30     30     0     80   123  
17:47:14.449 D5F7.1    stand  255  150    0      0     80   123  
17:47:15.464 D5F7.0    sit    None 30     30     0     80   123  
17:47:15.464 D5F7.1    stand  255  150    0      0     80   123  
17:47:16.481 D5F7.0    sit    None 30     30     0     80   123  
17:47:16.481 D5F7.1    stand  255  150    0      0     80   123  
17:47:17.380 D5F7.0    sit    None 30     30     0     80   123  
17:47:17.380 D5F7.1    stand  255  150    0      0     80   123  
17:47:18.394 D5F7.0    sit    None 30     30     0     80   123  
17:47:18.394 D5F7.1    stand  255  150    0      0     80   123  
17:47:19.409 D5F7.0    sit    None 30     30     0     80   123  
17:47:19.409 D5F7.1    stand  255  150    0      0     80   123  
17:47:20.487 D5F7.0    sit    None 30     30     0     80   123  
17:47:20.487 D5F7.1    stand  255  150    0      0     80   123  
17:47:21.346 D5F7.0    sit    None 30     30     0     80   123  
17:47:21.346 D5F7.1    stand  255  150    0      0     80   123  
17:47:22.357 D5F7.0    sit    None 30     30     0     80   123  
17:47:22.357 D5F7.1    stand  255  150    0      0     80   123  
17:47:23.371 D5F7.0    sit    None 30     30     0     80   123  
17:47:23.371 D5F7.1    stand  255  150    0      0     80   123  
17:47:24.385 D5F7.1    stand  255  150    0      0     80   0    
17:47:24.385 D5F7.0    sit    None 30     30     0     80   123  
17:47:25.398 D5F7.0    sit    None 30     30     0     80   0    
17:47:25.398 D5F7.1    stand  255  150    0      0     80   123  
17:47:26.313 D5F7.0    sit    None 30     30     0     80   123  
17:47:26.313 D5F7.1    stand  255  150    0      0     80   123  
17:47:27.319 D5F7.0    sit    None 30     30     0     80   123  
17:47:27.319 D5F7.1    stand  255  150    0      0     80   123  
17:47:28.363 D5F7.0    sit    None 30     30     0     80   123  
17:47:28.363 D5F7.1    stand  255  150    0      0     80   123  
17:47:29.348 D5F7.0    sit    None 30     30     0     80   123  
17:47:29.348 D5F7.1    stand  255  150    0      0     80   123  
17:47:30.364 D5F7.0    sit    None 30     30     0     80   123  
17:47:30.364 D5F7.1    stand  255  150    0      0     80   123  
17:47:31.265 D5F7.1    stand  255  150    0      0     80   0    
17:47:31.265 D5F7.0    sit    None 30     30     0     80   123  
17:47:32.279 D5F7.1    stand  255  150    0      0     80   123  
17:47:32.279 D5F7.0    sit    None 30     30     0     80   123  
17:47:33.290 D5F7.0    sit    None 30     30     0     80   0    
17:47:33.290 D5F7.1    stand  255  150    0      0     80   123  
17:47:34.304 D5F7.0    sit    None 30     30     0     80   123  
17:47:34.304 D5F7.1    stand  255  150    0      0     80   123  
17:47:35.317 D5F7.0    sit    None 30     30     0     80   123  
17:47:35.317 D5F7.1    stand  255  150    0      0     80   123  
17:47:36.229 D5F7.1    stand  255  150    0      0     80   0    
17:47:36.229 D5F7.0    sit    None 30     30     0     80   123  
17:47:37.241 D5F7.1    stand  255  150    0      0     80   123  
17:47:37.241 D5F7.0    sit    None 30     30     0     80   123  
17:47:38.268 D5F7.1    stand  255  150    0      0     80   123  
17:47:38.268 D5F7.0    sit    None 30     30     0     80   123  
17:47:39.264 D5F7.0    sit    None 30     30     0     80   0    
17:47:39.264 D5F7.1    stand  255  150    0      0     80   123  
17:47:40.280 D5F7.0    sit    None 30     30     0     80   123  
17:47:40.280 D5F7.1    stand  255  150    0      0     80   123  
17:47:41.188 D5F7.0    sit    None 30     30     0     80   123  
17:47:41.188 D5F7.1    stand  255  150    0      0     80   123  
17:47:42.199 D5F7.0    sit    None 30     30     0     80   123  
17:47:42.199 D5F7.1    stand  255  150    0      0     80   123  
17:47:43.213 D5F7.0    sit    None 30     30     0     80   123  
17:47:43.213 D5F7.1    stand  255  150    0      0     80   123  
17:47:44.224 D5F7.0    sit    None 30     30     0     80   123  
17:47:44.224 D5F7.1    stand  255  150    0      0     80   123  
17:47:45.240 D5F7.1    stand  255  150    0      0     80   0    
17:47:45.240 D5F7.0    sit    None 30     30     0     80   123  
17:47:46.161 D5F7.1    stand  255  150    0      0     80   123  
17:47:46.161 D5F7.0    sit    None 30     30     0     80   123  
17:47:47.161 D5F7.0    sit    None 30     30     0     80   0    
17:47:47.161 D5F7.1    stand  255  150    0      0     80   123  
17:47:48.171 D5F7.0    sit    None 30     30     0     80   123  
17:47:48.171 D5F7.1    stand  255  150    0      0     80   123  
17:47:49.188 D5F7.0    sit    None 30     30     0     80   123  
17:47:49.188 D5F7.1    stand  255  150    0      0     80   123  
17:47:50.201 D5F7.0    sit    None 30     30     0     80   123  
17:47:50.201 D5F7.1    stand  255  150    0      0     80   123  
17:47:51.105 D5F7.0    sit    None 30     30     0     80   123  
17:47:51.105 D5F7.1    stand  255  150    0      0     80   123  
17:47:52.127 D5F7.0    sit    None 30     30     0     80   123  
17:47:52.127 D5F7.1    stand  255  150    0      0     80   123  
17:47:53.194 D5F7.0    sit    None 30     30     0     80   123  
17:47:53.194 D5F7.1    stand  255  150    0      0     80   123  
17:47:54.179 D5F7.0    sit    None 30     30     0     80   123  
17:47:54.179 D5F7.1    stand  255  150    0      0     80   123  
17:47:55.080 D5F7.1    stand  255  150    0      0     80   0    
17:47:55.080 D5F7.0    sit    None 30     30     0     80   123  
17:47:56.098 D5F7.0    sit    None 30     30     0     80   0    
17:47:56.098 D5F7.1    stand  255  150    0      0     80   123  
17:47:57.100 D5F7.1    stand  255  150    0      0     80   0    
17:47:57.100 D5F7.0    sit    None 30     30     0     80   123  
17:47:58.117 D5F7.1    stand  255  150    0      0     80   123  
17:47:58.117 D5F7.0    sit    None 30     30     0     80   123  
17:47:59.131 D5F7.1    stand  255  150    0      0     80   123  
17:47:59.131 D5F7.0    sit    None 30     30     0     80   123  
17:48:00.033 D5F7.0    sit    None 30     30     0     80   0    
17:48:00.033 D5F7.1    stand  255  150    0      0     80   123  
17:48:01.057 D5F7.0    sit    None 30     30     0     80   123  
17:48:01.057 D5F7.1    stand  255  150    0      0     80   123  
17:48:02.060 D5F7.0    sit    None 30     30     0     80   123  
17:48:02.060 D5F7.1    stand  255  150    0      0     80   123  
17:48:03.091 D5F7.0    sit    None 30     30     0     80   123  
17:48:03.091 D5F7.1    stand  255  150    0      0     80   123  
17:48:04.034 D5F7.1    stand  255  150    0      0     80   0    
17:48:04.034 D5F7.0    sit    None 30     30     0     80   123  
17:48:05.019 D5F7.1    stand  255  150    0      0     80   123  
17:48:05.019 D5F7.0    sit    None 30     30     0     80   123  
17:48:06.035 D5F7.1    stand  255  150    0      0     80   123  
17:48:06.035 D5F7.0    sit    None 30     30     0     80   123  
17:48:07.048 D5F7.0    sit    None 30     30     0     80   0    
17:48:07.048 D5F7.1    stand  255  150    0      0     80   123  
17:48:08.054 D5F7.0    sit    None 30     30     0     80   123  
17:48:08.054 D5F7.1    stand  255  150    0      0     80   123  
17:48:08.964 D5F7.0    sit    None 30     30     0     80   123  
17:48:08.964 D5F7.1    stand  255  150    0      0     80   123  
17:48:09.975 D5F7.0    sit    None 30     30     0     80   123  
17:48:09.975 D5F7.1    stand  255  150    0      0     80   123  
17:48:10.987 D5F7.0    sit    None 30     30     0     80   123  
17:48:10.987 D5F7.1    stand  255  150    0      0     80   123  
17:48:12.004 D5F7.1    stand  255  150    0      0     80   0    
17:48:12.004 D5F7.0    sit    None 30     30     0     80   123  
17:48:13.015 D5F7.1    stand  255  150    0      0     80   123  
17:48:13.015 D5F7.0    sit    None 30     30     0     80   123  
17:48:13.930 D5F7.1    stand  255  150    0      0     80   123  
17:48:13.930 D5F7.0    sit    None 30     30     0     80   123  
17:48:14.937 D5F7.0    sit    None 30     30     0     80   0    
17:48:14.937 D5F7.1    stand  255  150    0      0     80   123  
17:48:15.952 D5F7.1    stand  255  150    0      0     80   0    
17:48:15.952 D5F7.0    sit    None 30     30     0     80   123  
17:48:16.963 D5F7.0    sit    None 30     30     0     80   0    
17:48:16.963 D5F7.1    stand  255  150    0      0     80   123  
17:48:17.979 D5F7.1    stand  255  150    0      0     80   0    
17:48:17.979 D5F7.0    sit    None 30     30     0     80   123  
17:48:18.881 D5F7.1    stand  255  150    0      0     80   123  
17:48:18.881 D5F7.0    sit    None 30     30     0     80   123  
17:48:19.895 D5F7.1    stand  255  150    0      0     80   123  
17:48:19.895 D5F7.0    sit    None 30     30     0     80   123  
17:48:20.988 D5F7.1    stand  255  150    0      0     80   123  
17:48:20.988 D5F7.0    sit    None 30     30     0     80   123  
17:48:21.946 D5F7.1    stand  255  150    0      0     80   123  
17:48:21.946 D5F7.0    sit    None 30     30     0     80   123  
17:48:22.849 D5F7.1    stand  255  150    0      0     80   123  
17:48:22.849 D5F7.0    sit    None 30     30     0     80   123  
17:48:23.866 D5F7.1    stand  255  150    0      0     80   123  
17:48:23.866 D5F7.0    sit    None 30     30     0     80   123  
17:48:24.876 D5F7.1    stand  255  150    0      0     80   123  
17:48:24.876 D5F7.0    sit    None 30     20     59    80   121  
17:48:25.899 D5F7.1    stand  255  150    -10    0     80   123  
17:48:25.899 D5F7.0    sit    None 40     0      77    80   110  
17:48:26.908 D5F7.1    stand  255  160    -60    0     80   134  
17:48:26.908 D5F7.0    sit    None 30     -20    79    80   136  
17:48:27.812 D5F7.0    sit    None 0      -30    64    80   31   
17:48:27.812 D5F7.1    stand  255  160    -70    0     80   164  
17:48:28.826 D5F7.0    sit    None 10     -40    83    80   152  
17:48:28.826 D5F7.1    stand  255  150    -70    0     80   143  
17:48:29.841 D5F7.0    sit    None 0      -50    71    80   151  
17:48:29.841 D5F7.1    stand  255  150    -50    0     80   150  
17:48:30.853 D5F7.0    sit    None 20     -10    67    80   136  
17:48:30.853 D5F7.1    stand  255  170    -20    0     80   150  
17:48:31.866 D5F7.1    stand  255  170    -10    0     80   10   
17:48:31.866 D5F7.0    sit    None 30     0      57    80   140  
17:48:32.776 D5F7.0    sit    None 30     -10    52    80   10   
17:48:32.776 D5F7.1    stand  255  170    -10    0     80   140  
17:48:33.795 D5F7.0    sit    None 20     -10    68    80   150  
17:48:33.795 D5F7.1    stand  255  160    0      0     80   140  
17:48:34.800 D5F7.0    sit    None 10     -20    67    80   151  
17:48:34.800 D5F7.1    stand  255  140    -30    0     80   130  
17:48:35.812 D5F7.0    sit    None 0      -30    79    80   140  
17:48:35.812 D5F7.1    stand  255  140    -60    0     80   143  
17:48:36.824 D5F7.0    sit    None 20     0      0     80   134  
17:48:36.824 D5F7.1    stand  255  160    -10    0     80   140  
17:48:37.736 D5F7.0    sit    None 30     0      0     80   130  
17:48:37.736 D5F7.1    stand  255  160    -10    0     80   130  
17:48:38.748 D5F7.1    stand  255  150    -20    0     80   14   
17:48:38.748 D5F7.0    sit    None 30     -10    69    80   120  
17:48:39.761 D5F7.1    stand  255  150    -20    0     80   120  
17:48:39.761 D5F7.0    sit    None 50     0      0     80   101  
17:48:40.775 D5F7.1    stand  255  150    -20    0     80   101  
17:48:40.775 D5F7.0    sit    None 50     0      82    80   101  
17:48:41.789 D5F7.0    sit    None 40     -10    63    80   14   
17:48:41.789 D5F7.1    stand  255  150    -20    0     80   110  
17:48:42.703 D5F7.0    sit    None 40     -10    0     80   110  
17:48:42.703 D5F7.1    stand  255  160    -50    0     80   126  
17:48:43.705 D5F7.0    sit    None 50     0      0     80   120  
17:48:43.705 D5F7.1    stand  255  160    -50    0     80   120  
17:48:44.717 D5F7.0    sit    None 50     0      0     80   120  
17:48:44.717 D5F7.1    stand  255  160    -50    0     80   120  
17:48:45.728 D5F7.1    stand  255  160    -50    0     80   0    
17:48:45.728 D5F7.0    sit    None 50     0      0     80   120  
17:48:46.742 D5F7.0    sit    None 50     0      0     80   0    
17:48:46.742 D5F7.1    stand  255  160    -50    0     80   120  
17:48:47.652 D5F7.1    stand  255  160    -50    0     80   0    
17:48:47.652 D5F7.0    sit    None 50     0      0     80   120  
17:48:48.666 D5F7.0    sit    None 50     0      0     80   0    
17:48:48.666 D5F7.1    stand  255  160    -50    0     80   120  
17:48:49.676 D5F7.0    sit    None 50     0      0     80   120  
17:48:49.676 D5F7.1    stand  255  160    -50    0     80   120  
17:48:50.721 D5F7.0    sit    None 50     0      0     80   120  
17:48:50.721 D5F7.1    stand  255  160    -50    0     80   120  
17:48:51.646 D5F7.0    sit    None 50     0      0     80   120  
17:48:51.646 D5F7.1    stand  255  160    -50    0     80   120  
17:48:52.653 D5F7.0    sit    None 50     -10    67    80   117  
17:48:52.653 D5F7.1    stand  255  160    -50    0     80   117  
17:48:53.688 D5F7.1    stand  255  160    -20    0     80   30   
17:48:53.688 D5F7.0    sit    None 30     10     79    80   133  
17:48:54.676 D5F7.0    sit    None 30     10     0     80   0    
17:48:54.676 D5F7.1    stand  255  160    -20    0     80   133  
17:48:55.586 D5F7.0    sit    None 30     10     0     80   133  
17:48:55.586 D5F7.1    stand  255  160    -20    0     80   133  
17:48:56.605 D5F7.0    sit    None 30     10     0     80   133  
17:48:56.605 D5F7.1    stand  255  160    -20    0     80   133  
17:48:57.615 D5F7.0    sit    None 30     0      0     80   131  
17:48:57.615 D5F7.1    stand  255  160    -20    0     80   131  
17:48:58.628 D5F7.0    sit    None 30     0      68    80   131  
17:48:58.628 D5F7.1    stand  255  160    -20    0     80   131  
17:48:59.640 D5F7.0    sit    None 40     -20    84    80   120  
17:48:59.640 D5F7.1    stand  255  160    -20    0     80   120  
17:49:00.550 D5F7.0    sit    None 40     -10    0     80   120  
17:49:00.550 D5F7.1    stand  255  160    -20    0     80   120  
17:49:01.562 D5F7.0    sit    None 40     -10    0     80   120  
17:49:01.562 D5F7.1    stand  255  160    -20    0     80   120  
17:49:02.574 D5F7.0    sit    None 40     -10    0     80   120  
17:49:02.574 D5F7.1    stand  255  160    -20    0     80   120  
17:49:03.602 D5F7.0    sit    None 40     -20    87    80   120  
17:49:03.602 D5F7.1    stand  255  160    -20    0     80   120  
17:49:04.605 D5F7.0    sit    None 40     -20    81    80   120  
17:49:04.605 D5F7.1    stand  255  160    -20    0     80   120  
17:49:05.506 D5F7.0    sit    None 40     -10    103   80   120  
17:49:05.506 D5F7.1    stand  255  160    -20    0     80   120  
17:49:06.531 D5F7.1    stand  255  160    -20    0     80   0    
17:49:06.531 D5F7.0    sit    None 40     -10    0     80   120  
17:49:07.546 D5F7.1    stand  255  160    -20    0     80   120  
17:49:07.546 D5F7.0    sit    None 40     -10    0     80   120  
17:49:08.563 D5F7.1    stand  255  160    -20    0     80   120  
17:49:08.563 D5F7.0    sit    None 40     -10    0     80   120  
17:49:09.568 D5F7.0    sit    None 40     -10    0     80   0    
17:49:09.568 D5F7.1    stand  255  160    -20    0     80   120  
17:49:10.468 D5F7.0    sit    None 40     -10    0     80   120  
17:49:10.468 D5F7.1    stand  255  160    -20    0     80   120  
17:49:11.483 D5F7.0    sit    None 40     -10    0     80   120  
17:49:11.483 D5F7.1    stand  255  160    -20    0     80   120  
17:49:12.493 D5F7.0    sit    None 30     -10    0     80   130  
17:49:12.493 D5F7.1    stand  255  160    -20    0     80   130  
17:49:13.506 D5F7.1    stand  255  150    -10    0     80   14   
17:49:13.506 D5F7.0    sit    None 40     -20    77    80   110  
17:49:14.518 D5F7.0    sit    None 10     -30    87    80   31   
17:49:14.518 D5F7.1    stand  255  130    -50    0     80   121  
17:49:15.431 D5F7.0    sit    None 20     -10    0     80   117  
17:49:15.431 D5F7.1    stand  255  140    -50    0     80   126  
17:49:16.441 D5F7.1    stand  255  140    -50    0     80   0    
17:49:16.441 D5F7.0    sit    None 20     0      61    80   130  
17:49:17.458 D5F7.1    stand  255  150    -20    0     80   131  
17:49:17.458 D5F7.0    sit    None 30     -10    82    80   120  
17:49:18.477 D5F7.0    sit    None 40     0      0     80   14   
17:49:18.477 D5F7.1    stand  255  150    -20    0     80   111  
17:49:19.573 D5F7.0    sit    None 40     0      0     80   111  
17:49:19.573 D5F7.1    stand  255  150    -20    0     80   111  
17:49:20.398 D5F7.1    stand  255  150    -20    0     80   0    
17:49:20.398 D5F7.0    sit    None 40     0      0     80   111  
17:49:21.401 D5F7.0    sit    None 40     0      0     80   0    
17:49:21.401 D5F7.1    stand  255  150    -30    0     80   114  
17:49:22.426 D5F7.1    stand  255  150    -30    0     80   0    
17:49:22.426 D5F7.0    sit    None 40     0      0     80   114  
17:49:23.432 D5F7.0    sit    None 40     0      0     80   0    
17:49:23.432 D5F7.1    stand  255  150    -30    0     80   114  
17:49:24.449 D5F7.0    sit    None 40     0      0     80   114  
17:49:24.449 D5F7.1    stand  255  150    -30    0     80   114  
17:49:25.345 D5F7.1    stand  255  150    -30    0     80   0    
17:49:25.345 D5F7.0    sit    None 40     0      0     80   114  
17:49:26.361 D5F7.0    sit    None 40     0      0     80   0    
17:49:26.361 D5F7.1    stand  255  150    -30    0     80   114  
17:49:27.376 D5F7.0    sit    None 40     0      0     80   114  
17:49:27.376 D5F7.1    stand  255  150    -30    0     80   114  
17:49:28.389 D5F7.1    stand  255  150    -30    0     80   0    
17:49:28.389 D5F7.0    sit    None 40     0      56    80   114  
17:49:29.399 D5F7.0    sit    None 30     0      62    80   10   
17:49:29.399 D5F7.1    stand  255  160    0      0     80   130  
17:49:30.311 D5F7.1    stand  255  150    -40    0     80   41   
17:49:30.311 D5F7.0    sit    None 50     -10    0     80   104  
17:49:31.319 D5F7.0    sit    None 50     -10    0     80   0    
17:49:31.319 D5F7.1    stand  255  150    -40    0     80   104  
17:49:32.332 D5F7.1    stand  255  150    -40    0     80   0    
17:49:32.332 D5F7.0    sit    None 50     -10    0     80   104  
17:49:33.348 D5F7.0    sit    None 50     -10    0     80   0    
17:49:33.348 D5F7.1    stand  255  150    -40    0     80   104  
17:49:34.359 D5F7.1    stand  255  150    -40    0     80   0    
17:49:34.359 D5F7.0    sit    None 50     -10    0     80   104  
17:49:35.268 D5F7.1    stand  255  150    -40    0     80   104  
17:49:35.268 D5F7.0    sit    None 50     -10    0     80   104  
17:49:36.288 D5F7.1    stand  255  150    -40    0     80   104  
17:49:36.288 D5F7.0    sit    None 50     -10    0     80   104  
17:49:37.292 D5F7.1    stand  255  150    -40    0     80   104  
17:49:37.292 D5F7.0    sit    None 50     -10    0     80   104  
17:49:38.314 D5F7.0    sit    None 40     0      73    80   14   
17:49:38.314 D5F7.1    stand  255  150    -40    0     80   117  
17:49:39.245 D5F7.1    stand  255  150    -40    0     80   0    
17:49:39.245 D5F7.0    sit    None 30     -10    0     80   123  
17:49:40.248 D5F7.0    sit    None 30     -10    0     80   0    
17:49:40.248 D5F7.1    stand  255  150    -40    0     80   123  
17:49:41.265 D5F7.0    sit    None 30     -10    0     80   123  
17:49:41.265 D5F7.1    stand  255  150    -40    0     80   123  
17:49:42.274 D5F7.0    sit    None 30     -10    0     80   123  
17:49:42.274 D5F7.1    stand  255  150    -40    0     80   123  
17:49:43.288 D5F7.0    sit    None 30     -10    0     80   123  
17:49:43.288 D5F7.1    stand  255  150    -40    0     80   123  
17:49:44.194 D5F7.1    stand  255  150    -40    0     80   0    
17:49:44.194 D5F7.0    sit    None 30     -10    0     80   123  
17:49:45.209 D5F7.0    sit    None 30     -10    0     80   0    
17:49:45.209 D5F7.1    stand  255  150    -40    0     80   123  
17:49:46.224 D5F7.1    stand  255  150    -40    0     80   0    
17:49:46.224 D5F7.0    sit    None 30     -10    0     80   123  
17:49:47.250 D5F7.1    stand  255  150    -40    0     80   123  
17:49:47.250 D5F7.0    sit    None 30     -10    0     80   123  
17:49:48.256 D5F7.0    sit    None 30     -10    0     80   0    
17:49:48.256 D5F7.1    stand  255  150    -40    0     80   123  
17:49:49.158 D5F7.0    sit    None 30     -10    0     80   123  
17:49:49.158 D5F7.1    stand  255  150    -40    0     80   123  
17:49:50.174 D5F7.1    stand  255  150    -40    0     80   0    
17:49:50.174 D5F7.0    sit    None 30     -10    0     80   123  
17:49:51.183 D5F7.0    sit    None 30     -10    82    80   0    
17:49:51.183 D5F7.1    stand  255  150    -40    0     80   123  
17:49:52.193 D5F7.1    stand  255  160    -10    0     80   31   
17:49:52.193 D5F7.0    sit    None 30     -10    74    80   130  
17:49:53.214 D5F7.0    sit    None 20     -10    0     80   10   
17:49:53.214 D5F7.1    stand  255  140    -10    0     80   120  
17:49:54.120 D5F7.1    stand  255  140    -20    0     80   10   
17:49:54.120 D5F7.0    sit    None 10     -10    70    80   130  
17:49:55.166 D5F7.0    sit    None 10     -10    0     80   0    
17:49:55.166 D5F7.1    stand  255  140    -20    0     80   130  
17:49:56.176 D5F7.1    stand  255  140    -20    0     80   0    
17:49:56.176 D5F7.0    sit    None 10     -40    61    80   131  
17:49:57.191 D5F7.0    sit    None 30     -20    53    80   28   
17:49:57.191 D5F7.1    stand  255  150    -20    0     80   120  
17:49:58.085 D5F7.1    stand  255  140    -40    0     80   22   
17:49:58.085 D5F7.0    sit    None 30     0      82    80   117  
17:49:59.099 D5F7.1    stand  255  150    -40    0     80   126  
17:49:59.099 D5F7.0    sit    None 40     0      0     80   117  
17:50:00.109 D5F7.0    sit    None 60     0      0     80   20   
17:50:00.109 D5F7.1    stand  255  150    -40    0     80   98   
17:50:01.124 D5F7.1    stand  255  150    -40    0     80   0    
17:50:01.124 D5F7.0    sit    None 60     0      68    80   98   
17:50:02.147 D5F7.1    stand  255  150    -40    0     80   98   
17:50:02.147 D5F7.0    sit    None 30     0      73    80   126  
17:50:03.044 D5F7.0    sit    None 40     0      49    80   10   
17:50:03.044 D5F7.1    stand  255  150    -40    0     80   117  
17:50:04.060 D5F7.1    stand  255  160    -50    0     80   14   
17:50:04.060 D5F7.0    stand  None 40     0      59    80   130  
17:50:05.069 D5F7.1    stand  255  150    -20    0     80   111  
17:50:05.069 D5F7.0    stand  None 20     0      67    80   131  
17:50:06.087 D5F7.0    stand  None 10     0      80    80   10   
17:50:06.087 D5F7.1    stand  255  150    0      0     80   140  
17:50:07.102 D5F7.1    stand  255  140    -10    0     80   14   
17:50:07.102 D5F7.0    sit    None 20     0      66    80   120  
17:50:08.008 D5F7.0    sit    None 30     -10    0     80   14   
17:50:08.008 D5F7.1    stand  255  140    -10    0     80   110  
17:50:09.021 D5F7.1    stand  255  140    -10    0     80   0    
17:50:09.021 D5F7.0    sit    None 30     -10    0     80   110  
17:50:10.042 D5F7.1    stand  255  150    -20    0     80   120  
17:50:10.042 D5F7.0    sit    None 30     -10    94    80   120  
17:50:11.045 D5F7.1    stand  255  160    -30    0     80   131  
17:50:11.045 D5F7.0    sit    None 30     0      56    80   133  
17:50:12.058 D5F7.1    stand  255  170    -10    0     80   140  
17:50:12.058 D5F7.0    sit    None 40     0      53    80   130  
17:50:12.971 D5F7.0    sit    None 40     0      68    80   0    
17:50:12.971 D5F7.1    stand  255  160    -60    0     80   134  
17:50:13.982 D5F7.1    stand  255  160    -70    0     80   10   
17:50:13.982 D5F7.0    sit    None 40     0      81    80   138  
17:50:14.994 D5F7.0    sit    None 40     0      103   80   0    
17:50:14.994 D5F7.1    stand  255  160    -70    0     80   138  
17:50:16.004 D5F7.1    stand  255  160    -50    0     80   20   
17:50:16.004 D5F7.0    sit    None 40     0      61    80   130  
17:50:17.024 D5F7.1    stand  255  160    -10    0     80   120  
17:50:17.024 D5F7.0    sit    None 40     0      81    80   120  
17:50:17.929 D5F7.0    sit    None 50     0      86    80   10   
17:50:17.929 D5F7.1    stand  255  180    -40    0     80   136  
17:50:18.940 D5F7.1    stand  255  180    -50    0     80   10   
17:50:18.940 D5F7.0    sit    None 50     0      94    80   139  
17:50:20.033 D5F7.0    sit    None 30     0      105   80   20   
17:50:20.033 D5F7.1    stand  255  180    -50    0     80   158  
17:50:20.989 D5F7.0    sit    None 30     0      56    80   158  
17:50:20.989 D5F7.1    stand  255  160    -30    0     80   133  
17:50:21.893 D5F7.1    stand  255  160    20     0     80   50   
17:50:21.893 D5F7.0    sit    None 60     0      49    80   101  
17:50:22.906 D5F7.1    stand  255  180    0      0     80   120  
17:50:22.906 D5F7.0    sit    None 70     0      70    80   110  
17:50:23.920 D5F7.0    sit    None 60     0      65    80   10   
17:50:23.920 D5F7.1    stand  255  180    0      0     80   120  
17:50:24.933 D5F7.1    stand  255  180    0      0     80   0    
17:50:24.933 D5F7.0    sit    None 50     -10    22    80   130  
17:50:25.953 D5F7.0    stand  None 50     -10    51    80   0    
17:50:25.953 D5F7.1    stand  255  180    0      0     80   130  
17:50:26.856 D5F7.1    stand  255  180    0      0     80   0    
17:50:26.856 D5F7.0    stand  None 60     -10    63    80   120  
17:50:27.870 D5F7.1    stand  255  170    0      0     80   110  
17:50:27.870 D5F7.0    stand  None 40     0      70    80   130  
17:50:28.884 D5F7.0    walk   None 0      0      69    80   40   
17:50:28.884 D5F7.1    stand  255  210    10     0     80   210  
17:50:29.897 D5F7.0    walk   None -30    0      88    80   240  
17:50:29.897 D5F7.1    stand  255  210    10     0     80   240  
17:50:30.908 D5F7.0    walk   None -30    -20    93    80   241  
17:50:30.908 D5F7.1    stand  255  200    10     0     80   231  
17:50:31.819 D5F7.1    stand  255  200    10     0     80   0    
17:50:31.819 D5F7.0    walk   None -40    -50    82    80   247  
17:50:32.827 D5F7.0    walk   None -50    -40    98    80   14   
17:50:32.827 D5F7.1    stand  255  200    10     0     80   254  
17:50:33.841 D5F7.0    walk   None -40    -30    92    80   243  
17:50:33.841 D5F7.1    stand  255  200    10     0     80   243  
17:50:34.858 D5F7.1    stand  255  200    10     0     80   0    
17:50:34.858 D5F7.0    walk   None -40    0      0     80   240  
17:50:35.870 D5F7.1    stand  255  200    10     0     80   240  
17:50:35.870 D5F7.0    walk   None -30    0      83    80   230  
17:50:36.782 D5F7.0    walk   None -20    -10    72    80   14   
17:50:36.782 D5F7.1    stand  255  200    20     0     80   222  
17:50:37.795 D5F7.1    stand  255  200    20     0     80   0    
17:50:37.795 D5F7.0    walk   None 0      -20    69    80   203  
17:50:38.812 D5F7.0    walk   None 0      -10    107   80   10   
17:50:38.812 D5F7.1    stand  255  200    20     0     80   202  
17:50:39.817 D5F7.1    stand  255  160    -20    0     80   56   
17:50:39.817 D5F7.0    walk   None -20    30     77    80   186  
17:50:40.828 D5F7.0    walk   None -60    50     76    80   44   
17:50:40.828 D5F7.1    stand  255  160    -20    0     80   230  
17:50:41.732 D5F7.1    stand  255  160    -20    0     80   0    
17:50:41.732 D5F7.0    walk   None -80    70     0     80   256  
17:50:42.806 D5F7.1    stand  255  160    -20    0     80   256  
17:50:42.806 D5F7.0    walk   None -80    80     0     80   260  
17:50:43.716 D5F7.0    stand  None -80    80     0     80   0    
17:50:43.716 D5F7.1    stand  255  160    -20    0     80   260  
17:50:44.732 D5F7.0    stand  None -90    80     0     80   269  
17:50:44.732 D5F7.1    stand  255  160    -20    0     80   269  
17:50:45.741 D5F7.0    stand  None -90    80     0     80   269  
17:50:45.741 D5F7.1    stand  255  160    -20    0     80   269  
17:50:46.818 D5F7.1    stand  255  160    -20    0     80   0    
17:50:46.818 D5F7.0    stand  None -90    80     0     80   269  
17:50:47.824 D5F7.1    stand  255  160    -20    0     80   269  
17:50:48.674 D5F7.1    stand  255  160    -20    0     80   0    
17:50:49.693 D5F7.1    stand  255  160    -20    0     80   0    
17:50:50.728 D5F7.1    stand  255  160    -20    0     80   0    
17:50:51.717 D5F7.1    stand  255  160    -20    0     80   0    
17:50:52.729 D5F7.1    stand  255  160    -20    0     80   0    
17:50:53.644 D5F7.1    stand  255  160    -20    0     80   0    
17:50:54.651 D5F7.1    stand  255  160    -20    0     80   0    
17:50:55.661 D5F7.1    stand  255  160    -20    0     80   0    
17:50:56.673 D5F7.1    stand  255  160    -20    0     80   0    
17:50:57.686 D5F7.1    stand  255  160    -20    0     80   0    
17:50:58.594 D5F7.1    stand  255  160    -20    0     80   0    
17:50:59.612 D5F7.1    stand  255  160    -20    0     80   0    
17:51:00.619 D5F7.1    stand  255  160    -20    0     80   0    
17:51:01.633 D5F7.1    stand  255  160    -20    0     80   0    
17:51:02.645 D5F7.1    stand  255  160    -20    0     80   0    
17:51:03.554 D5F7.1    stand  255  160    -20    0     80   0    
17:51:04.568 D5F7.1    stand  255  160    -20    0     80   0    
17:51:05.580 D5F7.1    stand  255  160    -20    0     80   0    
17:51:06.594 D5F7.1    stand  255  160    -20    0     80   0    
17:51:07.610 D5F7.1    stand  255  160    -20    0     80   0    
17:51:08.516 D5F7.1    stand  255  160    -20    0     80   0    
17:51:09.527 D5F7.1    stand  255  160    -20    0     80   0    
17:51:10.544 D5F7.1    stand  255  160    -20    0     80   0    
17:51:11.556 D5F7.1    stand  255  160    -20    0     80   0    
17:51:12.565 D5F7.1    stand  255  160    -20    0     80   0    
17:51:13.509 D5F7.88   88     -    -      -      -     -    -    
17:51:14.466 D5F7.88   88     -    -      -      -     -    -    
17:51:15.480 D5F7.88   88     -    -      -      -     -    -    
17:51:28.733 D5F7.88   88     -    -      -      -     -    -    
17:52:00.102 D5F7.88   88     -    -      -      -     -    -    
17:52:32.016 D5F7.88   88     -    -      -      -     -    -    

```

**汇总**: xray tick 637 | fire 0 | Fall 事件 0 () | 结论 = 无 Fall 无 fire

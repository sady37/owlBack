case-cd2b-0620-1123113
11:22 A enter 201BedRoom, ExitRoom , EnterRoom  201Bathroom , exitRomm 卫生间更衣室，333B不可见
11:23  B Enter 201BedRoom , InBed
11:24  B sleepad detect  B, Sleepad inBed ,HR/RR
11:25  B fall from Bed，p0, 但仍显示在bedside
11:30  B p0 stillbox,  
11:30  A 从Bathroom 走到Bedroom, 发生track 跳跃， 可能radar 在有real track出现时，坐标直接路到A
11:30  A exitRoom
11:31  B 起身，ExitRoom, 但Enter处 金属把手被显示为一个静止的人

vent_log 拿到了(权威)。关键发现:全程无 firmware Fall 事件——固件没判这次摔(cd2b 本质:固件漏,sensor 必须自己判)。关键时间点(bedroom = CD2B 雷达 + 1641 sleepad,剔除 activity/number_people 噪声):

本地时间	设备	事件	含义
11:23:04	CD2B	EnterRoom	A 进卧室
11:23:16	333B	EnterRoom	A→卫生间
11:23:22	CD2B	ExitRoom	A 离卧室
11:23:45	CD2B	InBed	雷达:有人在床(B)
11:23:56	1641	InBed(bed_status=0)	sleepad:B 在床
11:24:08	CD2B+1641	LeftBed(bed_status=1)	B 短暂离床
11:24:21	CD2B	ExitRoom	
11:24:34	CD2B	InBed	B 回床
~11:24:50–11:25:19	CD2B	radar pose=6 Lying 持续(area=1)	躺着
11:25:12	1641	InBed(bed_status=0)	sleepad 仍判 B 在床
~11:25	—	🔴 真摔(床边)— event_log 无 Fall	固件漏
11:25:36	1641	LeftBed(bed_status=1)	sleepad 离床(摔后 ~30s)
11:30:12	CD2B	EnterRoom	A 从卫生间→卧室(track 跳)
11:30:40	CD2B	ExitRoom	A 离
11:31:12	CD2B	InBed / 1641 InBed	B 起身后又回床
11:31:25→39	CD2B	EnterRoom→ExitRoom



11:23:38-46	 11:24:39 6    None InBed     | 0.01  0.85  0.04  0.05    0.02   0.01  | Bed   
11:23:55-58  sleepad  inbed
11:24:07-22 	

room 201 全部 device event(按时间,3 设备:CD2B 卧室雷达 / 333B 卫生间雷达 / 1641 床 sleepad):

时间	设备	事件	关键字段
11:22:33	CD2B	activity	trk9
11:22:54	CD2B	number_people	trk10 start
11:23:04	CD2B	EnterRoom	A 进卧室
11:23:16	333B	EnterRoom	A→卫生间
11:23:22	CD2B	ExitRoom	
11:23:34	333B	ExitRoom	
11:23:37	333B	activity	trk9
11:23:38	CD2B	activity	trk9
11:23:45	CD2B	InBed	雷达
11:23:56	1641	sleep-stage / InBed	bed=0
11:24:08	CD2B	LeftBed	
11:24:08	1641	LeftBed	bed=1
11:24:10	333B	EnterRoom	
11:24:17	333B	ExitRoom	
11:24:21	CD2B	ExitRoom	
11:24:33	CD2B	activity	trk9
11:24:34	CD2B	InBed	B 回床
11:24:41	333B	activity	trk9
11:25:12	1641	InBed / sleep-stage	bed=0
~11:25	—	🔴 真摔(无任何 event)	radar pose=6 Lying 持续
11:25:32	CD2B	activity	trk9
11:25:36	1641	LeftBed	bed=1
11:25:45	333B	activity	trk9
11:26:31	CD2B	activity	trk9
11:26:48	333B	activity	trk9
11:27:31	CD2B	activity	trk9
11:27:35	CD2B	number_people	trk10
11:27:52	333B	activity	trk9
11:28:55	333B	activity	trk9
11:29:00	CD2B	activity	trk9
11:29:32	CD2B	activity	trk9
11:29:59	333B	activity	trk9
11:30:08	333B	EnterRoom	
11:30:12	CD2B	EnterRoom	A 卫生间→卧室(track 跳)
11:30:16	333B	ExitRoom	
11:30:29	CD2B	activity	trk9
11:30:40	CD2B	ExitRoom	A 离
11:31:02	333B	activity	trk9
11:31:12	CD2B	InBed	B 又回床
11:31:21	1641	InBed / sleep-stage	bed=0
11:31:25	CD2B	EnterRoom	trk1
11:31:30	CD2B	activity	trk9
11:31:31	1641	LeftBed	bed=1
11:31:39	CD2B	ExitRoom	trk1
11:31:53	CD2B	EnterRoom	trk1
11:32:12	CD2B	number_people	trk10
11:32:29	CD2B	activity	trk9
11:32:37	333B	activity	trk9
(deviceStatus 心跳两条 11:22:48/11:23:06 333B 省略。)

两个观察:

整段无 Fall 事件——固件全程没判摔(cd2b 本质)。
11:25 摔倒窗只有 activity(trk9),无 Enter/Exit/Fall;sleepad 11:25:12 还报 InBed(bed=0),摔后 ~30s(11:25:36)才 LeftBed。
trk9

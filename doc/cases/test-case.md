replay集
# * case-d5f7-0617-23252357 replay  false stand, lost, 
# * case-d5f7-0627-17431755 bashroom, ghost,exitRoom, ghost lsot


# * case-cabb-0616-17441802	replay  hunzi, bathroom lost,ghost,
# * case-cd2b-0617-23302345-curtain  bedroom 窗帘误报  这个基本无解  replay  z:有高有低
# * case-cd2b-0620-11231131  hand-off
# * case-cd2b-0627-04270442  curtain出生，才豁免，
# * case-cd2b-0628-02280232  fire fall in bed, 

# * case-09e7-0620-22402242   
# * case-09e7-6021-22162229   1room2radar,reisktime,fall2,9min fall     **success  同房在床，可能睡，不抑制fire报警

# * case-d523-0622-13341336   D523,椅子后80s lost, 120s 101guest sleepad inbed handoff  **success
# * case-d523-0627-15351540   D523，重启， stillbox learn sit 持久化，避免重启丢失。 stillbox >5min后 平均数，tFloor=media*1.5
# * case-d523-0629-17021705   03:30 04:01 瞬间移动

# * case-B197-0624-09180932  livingRoom 2人， 有Ghost, 有2次fall. 
# * case-B197-0624-08430856  livingRoom  ExitRoom,  BlindArea, 12min  handle->Enter/Door  Exit zone: add permanently  
# * case-b197-0626-15371540  livingRoom   镜面/反射 ghost 签名(一条假轨长期挂在固定小区,跟真人并存）
# * casse-B197-0627-09280932  livingRoom  openArea  lost  有一边是开放空间
# * casse-B197-0627-15311535  livingRoom  openArea  lost  有一边是开放空间
#* case-b197-0630-08020806   livingRoom  openArea  lost  有一边是开放空间


# * case-1797-0627-13261328  mom bathroom ,  stillbox 整体30秒走了500， 要视为stillbox break
# * case-1797-0627-20282033  mom bathroom, 



# * case-B17F-0626-13321344  kitchen, 1个静止interfer, real split 2 track,  其中一个在island still



# * case-0777-0624-12001201  sandiego  learning area,  重复sitonground fire


# * case-d5f7-0617-23252357 replay
case1  bathroom  still+D
E598A2ACD5F7 20260617 23:25  
101 bathroom    

放置2个metal canbin 作干扰

23：26进入
缓慢倒倒，Radar 显示站wall side
中途爬行， position有移动，但仍在still-box
23:35  
爬出bathroom  p0 仍保持静止状态
23:36:xx   约30秒，人在外，
23:37  track消失

20min

# case-d5f7-0618-08500855
进入bathroom,快速跌倒，被Radar检测，站起，再fall, 爬出room,在pose=2时， 人已出bathroom, 


# * case-def7-0618-1138-1142  replay

11:38  进入， 
11：40  缓慢爬出，很快firmware就过滤掉，应该在30s左右
11:40:50 左右，  101GuestRoom Bed   inBed
handoff  neighbor event取消报警才对

11:38:55  Walking          人走动(进 bathroom)
11:39:04  Initialization
…人在 bathroom，StillSec 累积，pose 多为坐/站…
11:40:34  track: pose=3(坐) area=6(shower) xyz=(-70,0,0) z=0  ← 蹲/坐地(可能在爬)
11:40:35  ExitRoom ★       firmware 判人出门(爬出 bathroom)
11:40:36  track → tid=88(空占位)  ← 人消失 = LOST
11:40:57  case 结束



项	结果
fire	0（峰值 stillbox 234s=3.9min ≪ 18min 阈，符合预期；5540f07 的 0.979 确是累加吹的）
Sit 折扣 ×0.8	✅ ss/sb 全程 ≈0.80（z 在 0-67，Sit 的 0.8 经 min 压过 z 的 0.9/1.0）
Stand@z=0 → 1.0	✅ ss=sb（无折扣，正确）
path 闸 + 30cm 抖动地板	✅ 2 次破盒均 disp=60>50, path=0——破在位置范围超盒，抖动 path 恒 0（每段<30cm 全滤掉）= 抖动没误破盒
但 d5f7 没有"walking/z>80 + 还在 still"的帧：人一站起/走动就移动 → 破盒 sb=0 → 没有 stillsec 可折扣。Walk(6)/z>80(2) 帧的 sb 都=0。所以 d5f7 验不了 z-band×0.5 和 walking-while-still。

那个场景只在 cabb 出现（lost-carry 冻结帧：站姿 z=101 但 sb 还在跑）。stillsec 是虚拟时钟算的、与 speed 无关（你点过），我用 8x 快取一份（35s，stillsec 跟 speed=1 完全一致）



# case2  bedroom 窗帘误报  这个基本无解  replay

case-cd2b-0617-23302345-curtain
2026-06-17 23:31:03(America/Denver)
2026-06-17 23:42:24(America/Denver)
Denver-201	9D8A32A1CD2B	Fall	S




case-333b-0618-13301333
201bathroom  有跌倒，有ghost


# * case-cabb-0616-17441802	replay
	case (SH)	firmware	Xsensor	雷达姿态    bathroom 无ExitRoom
1	0612-05440555	报了 Fall	不报	Sitting×330 为主
7	0616-09250933	报了 Fall	不报	Standing×499 / Sitting×415
8	0616-17441802	报了 Fall	不报	Sitting×33 / Walking×17
16	0618-05180523	报了 Fall	不报	Sitting×351 为主

case-cabb-0612-05440555	报了 Fall	不报	Sitting×330 为主
case-cabb-0616-09250933	报了 Fall	不报	Standing×499 / Sitting×415
case-cabb-0616-17441802	报了 Fall	不报	Sitting×33 / Walking×17
case-cabb-0618-05180523	报了 Fall	不报	Sitting×351 为主


case-cabb-0616-17441802	报了 Fall	不报	Sitting×33 / Walking×17
 2个fall, 
 Z增高 

时间	tid	pose	z	effStillSec	SFallen	top	fire
09:44:37	0	Standing	95	0	0.002	Empty	❌
09:44:43	0	Walking	96	0	0.013	Empty	❌
09:44:52	0	Walking	92	0	0.035	Empty	❌
09:44:53	0	Walking	47	0	0.038	Empty	❌
…							
10:02:48	0	Standing	101	539	0.296	Fallen	❌
10:02:48	0	Standing	84	971	0.346	Fallen	❌
10:02:49	0	Standing	71	972	0.398	Fallen	❌
10:02:51	0	Walking	70	0	0.509	Fallen	❌
10:02:55	0	Walking	62	0	0.712	Fallen	❌


# * case-09e7-6021-22162229

22:16-17   d523*09e7 均可见track, 近60s, n_r 融合成1
22:17-18   先radar inbed, 有HR/RR
22:18      fall在床边，仍显示在sit模式，
22:21       still-box 有动，但不知 break still, 但后续仍是stillbox
22:28      27-18=9min, risktime , tFloor=8minn吧，应该报警
22:29      fall_1 不足10s, SFall 保持， 起身， fall_2  数秒， 此时Xsensor SFall应快速增加


# *  case-d523-0622-11141121
11:15 
D523可见， 9e7可见
走到L型，
D523--chair--person,，因椅背是实木+少许空隙，+2层纸箱+1个枕头  只能检出人
9e7--铁架+2张毯子， 仍能检测人体活动
11：21 出来

# * case-d523-0622-11571200

11:15 
D523可见， 9e7可见
走到L型，
D523--chair--person,，因椅背是实木+少许空隙，+2层纸箱+1个枕头  
9e7--铁架+2张毯子， 仍能检测人体活动
11:58 D523 lost track, FE仍显示

# * case-d523-0622-12381244
9e7 掉线，同room仅 D523+sleepad
走进D523, fall，forzen,  sleepad inbed

# * case-d523-0622-13341336
13:34  go to 桌子后，雷达被挡，frozen
80s  lost, tid=88
120s  101GuestRoom sleepad inBed  

导出完成，跨房场景确实捕到了。这个 /80 unit 有 3 个房：

设备	房	角色
D523（radar）	:100 Bedroom	人走桌后被挡 → frozen → lost(tid=88)
D5F7（radar）	:300（卫浴）	—
BM87224601903（sleepad）	:200 GuestRoom	+120s InBed
（注：你说的"101 GuestRoom"，实际 GuestRoom 是 room :200；导出按 D523 的 /80 unit 抓，GuestRoom 那台 sleepad 1903 在同 unit，已包进来。另一台 Bedroom sleepad 978 本窗口无数据没进。）

这是个跨房 hand-off 测试：人在 Bedroom(100) 被 D523 跟丢（桌后冻结），~120s 后在 GuestRoom(200) sleepad InBed = 人其实走去隔壁睡了。机制看点：

Bedroom 的 lost-fall 二义，能不能被 GuestRoom 的 sleepad InBed 解析成"人去隔壁了"（hand-off 宿）？
⚠️ 预判：现 hand-off（rhoFor）只认兄弟房 radar GainedReal，sleepad InBed 在别房算不算 hand-off 落点没接 → 大概率 Bedroom lost-fall 留着空兜底、而人其实安全在 GuestRoom 床上 = 跨房 sleepad hand-off 缺口。


# * case-cd2b-0628-02280232  fire fall in bed, 
一、全夜画像:90 分钟里系统 89 分钟是对的

p_fallen 全程 ≈ 0,直到 02:30:25 才第一次抬头
bed_reading 几乎全程 InBed(中间几次秒级抖动)
n_r 多数时间 1~2
关键结论:稳态睡眠被稳稳保护住了——人在床躺了一个半小时,p_fallen 一直贴地。所以"还是超过90min"在生产里其实没成立:床信用在稳态扛住了 90min。问题只发生在最后一个~24 秒的瞬态。这把根因从"久卧阈值"挪到了"瞬态 churn"。

二、转折点:02:26–02:30 轨道暴涨
时刻	事件
02:26–02:31 radar + sleepad 合并时间线

time      源    状态                                            标记
─────────────────────────────────────────────────────────────────────
02:26:00  雷达  t0:躺@床(a1)  t1:躺(a5)                         t0在床
02:26:13  垫    InBed  hr=72                                    ← 有心跳在床
02:26:42  雷达  t0→a255(离床区)  t1:站(a5)
02:26:42  垫    LeftBed                                         ← 秒级抖动
02:26:48  雷达  t0:站@a255(冻住)  t1:躺@床(a1)                  床信用换到 t1
··· 02:26:42–02:28:51  垫 静默 2 分钟(无上报) ···
02:28:23  雷达  t0:冻@a255  t1@床(a1)  ★t2 出生(无area)         t2 从此无 area_id,固件Bug, 在床但area_id=空
02:28:51  垫    InBed  hr=68                                    ← 人仍在床有心跳
02:29:01  垫    InBed  hr=73
02:29:13  垫    InBed  hr=69                                    ← 直到这秒都 InBed+心跳
02:29:15  垫    LeftBed  (保持到 fire 不再回)                   ★保护被撤
02:29:31  雷达  t1@床(a1)  t2:躺(无area)                        t2 第一次躺
02:29:51  雷达  t2:站                                          (起来)
02:30:17  雷达  t1@床(a1)  t2:躺(无area)                        t2 再躺→将settle
02:30:29  雷达  t1@床(a1)  t2:躺  ★t3 出生
02:30:42  雷达  ☠ t0、t1 同时丢轨    只剩 t2:躺 t3:站           床占用轨 t1 没了!
02:30:44  雷达  ☠ t2 丢轨  → ▶▶ FIRE (band=lost pose=6 still=23)
三、fire 解剖(与 alarm evidence 逐字吻合)

02:30:44 CD2B22822912 band=lost cell=AreaBed tfloor=720 still=23 pose=6 (60,160)
present 段 belief 已爬到 0.99(02:30:40),但只有丢轨那刻 lost 腿才开火 → 印证"loss 时已确认摔的 carry-over"
同帧 n_r=4:房里 4 条轨,至少 tid=0 是冻死 ghost
四、根因(被生产数据修正/深化)
之前我说"床信用只走固件 area_id/sleepad,学习 cell 不喂 SBed"。生产数据补上了真正撑住全夜的那条腿:

全夜的床保护 = sleepad InBed(接触轴)。它撑住了 90min,p_fallen≈0。
02:29:16 sleepad 翻 LeftBed,把这条唯一有效的保护腿撤了。
同期 churn:tid=2 这条轨在床区躺定,而它拿不到任何在床信用——sleepad 说 LeftBed、固件床信用此刻在别的轨上、学习 cell=Bed 不喂 SBed。
于是"躺着不动"被判成"倒在开阔地",23s 冲到 Fallen → lost fire。
一句话:不是"在床被误判",而是"sleepad 在关键时刻撤了在床信用 + 一条无信用的 churn 轨在床区躺下"。

五、必须查清的两个问题(决定怎么修)
02:29:16 sleepad 为什么翻 LeftBed? 人真的起床了(那 tid=2 躺床区是 ghost/反射),还是人没走、sleepad 丢接触(经典 [[bed_stale_leftbed_vetoes_radar_inbed]] 陈旧 LeftBed 否决真在床)?——这决定是该信 sleepad 还是信几何。
tid=0 冻在 (-80,230) 3+ 分钟的 ghost 哪来的? 这房 ghost/反射明显活跃(01:10、01:14 有 ghost_left_suppress),churn 是不是反射体没学进 AreaReflector?
要我接着拉 sleepace 日志看 02:26–02:31 这张垫子(1641?)的原始 InBed/LeftBed + 心率体动,定位它为什么在 02:29:16 翻 LeftBed 吗?这是解开"信 sleepad 还是信几何"的钥匙。

其实tid=0 站立卡死，实际上走进了bathroom  过一会，以tid=2的身份返回。
radar性能： fall 场景，只能跟踪2个tick的area_id,

02:28:22.912  radar.track  [{"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 320, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:22.912  radar.track  [{"pose": 4, "area_id": 1, "track_id": 1, "position_x": -10, "position_y": 150, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:22.912  radar.track  [{"pose": 4, "event": 1, "track_id": 2, "position_x": -130, "position_y": 0, "position_z": 70, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:23.869  radar.track  [{"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 320, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:23.869  radar.track  [{"pose": 4, "area_id": 1, "track_id": 1, "position_x": -10, "position_y": 150, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:23.869  radar.track  [{"pose": 4, "track_id": 2, "position_x": -140, "position_y": 50, "position_z": 98, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:24.876  radar.track  [{"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 320, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:24.876  radar.track  [{"pose": 4, "area_id": 1, "track_id": 1, "position_x": -50, "position_y": 200, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:24.876  radar.track  [{"pose": 1, "track_id": 2, "position_x": -140, "position_y": 100, "position_z": 81, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:25.881  radar.track  [{"pose": 4, "area_id": 255, "track_id": 0, "position_x": -170, "position_y": 320, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:25.881  radar.track  [{"pose": 4, "area_id": 1, "track_id": 1, "position_x": -90, "position_y": 220, "position_z": 0, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
02:28:25.881  radar.track  [{"pose": 1, "track_id": 2, "position_x": -140, "position_y": 100, "position_z": 54, "track_count": 3, "remaining_time": 0, "track_confidence": 80}]
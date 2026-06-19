# case-d5f7-0617-23252357
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


# case-def7-0618-1138-1142

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



# case2  bedroom 窗帘误报  这个基本无解

case-cd2b-0617-23302345-curtain
2026-06-17 23:31:03(America/Denver)
2026-06-17 23:42:24(America/Denver)
Denver-201	9D8A32A1CD2B	Fall	S




case-333b-0618-13301333
201bathroom  有跌倒，有ghost


#	case (SH)	firmware	Xsensor	雷达姿态    bathroom 无ExitRoom
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

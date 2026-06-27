#!/usr/bin/env python3
# timeline_from_xray.py — case 重放后,合并 [Xsensor.log xsensor_xray 房级 belief] + [window.json 双雷达 raw track/事件]
# + [sleepad 床事件] → track_event_timeline.md(每 tick 全景,格式同 doc/cases/*/track_event_timeline.md)。
#
# 用法: timeline_from_xray.py <case 目录> <room_prefix(如 fd00:0:3:111:3:100)> [xsensor.log]
#
# 多雷达同房:窗口里**每个雷达的 raw track 帧都出一行**(dev.tid=uid后4.track_id),belief 列(top/SFall..)取房级
# xray 最近 tick(房只按引擎实际用的 base track 跑信念)。于是"某雷达看到 FALL 但房信念没动"的 FN 一眼可见。
import os, sys, json, bisect, datetime, zoneinfo
from collections import defaultdict

case_dir = sys.argv[1].rstrip('/')
room_pfx = sys.argv[2]
log_path = sys.argv[3] if len(sys.argv) > 3 else '/home/wisefido/owl/log/Xsensor.log'
# 第 4 参=本房设备 uid 后 4(逗号分隔),过滤掉同 unit 其它房的雷达;省略=全收。
allow = set(s.upper() for s in sys.argv[4].split(',')) if len(sys.argv) > 4 else None
PLACEHOLDER_TIDS = {88, 11}  # firmware 心跳/无坐标占位 track,非真人
# 显示时区=设备所在地(epoch ms 不变,只换标签);跨时区 case 必传 CASE_TZ,否则 HH:MM:SS 错位。
TZ = zoneinfo.ZoneInfo(os.environ.get('CASE_TZ', 'America/Denver'))

POSE = {0:'init',1:'walk',2:'susfall',3:'sit',4:'stand',5:'FALL',6:'lying',
        7:'sitgnd',8:'sitgnd',9:'bedsit',10:'bedsit',11:'bedsit',12:'run'}
def pose_name(p): return POSE.get(p, f'p{p}')
def hhmmss(ms): return datetime.datetime.fromtimestamp(ms/1000, TZ).strftime('%H:%M:%S')

# ── 1) xsensor_xray 房级 belief tick ───────────────────────────────────────
ticks = []  # {ts, bed, top, sdist, still, fire, tidlid:{tid->lid}}
for line in open(log_path):
    if '"xsensor_xray"' not in line: continue
    try: d = json.loads(line)
    except json.JSONDecodeError: continue
    if not d.get('room','').startswith(room_pfx): continue
    sd = d.get('s_dist', {})
    # target(tid,pose,z,stillbox,x,y) ↔ dbn(lid,still_sec,x,y) 按 (x,y) 关联
    dbn_by_xy = {(t['x'], t['y']): t for t in d.get('dbn', [])}
    dbn_by_lid = {t['lid']: t for t in d.get('dbn', [])}  # per-track 信念(每条轨自己的 s_marg/p_real)
    tidlid, stills = {}, []
    for t in d.get('target', []):
        dd = dbn_by_xy.get((t['x'], t['y']))
        if dd and t.get('present'):
            tidlid[t['tid']] = dd['lid']
            stills.append(dd.get('still_sec', t.get('stillbox', 0)))
    ticks.append({'ts': d['ts'], 'bed': d.get('bed_reading','NoReport'),
                  'top': d.get('top_s','?'), 'sdist': sd,
                  'still': max(stills) if stills else 0, 'fire': d.get('fire', False),
                  'n_r': d.get('n_r'),  # 房内真人数(census 折叠后；2 雷达 1 人 → 1)
                  'tidlid': tidlid, 'dbn_by_lid': dbn_by_lid})
ticks.sort(key=lambda r: r['ts'])
tick_ts = [r['ts'] for r in ticks]

# 全局 (dev,track_id)->lid(引擎只给被采用的 base track 发 lid;未被采用的雷达 → 无 lid)
dev_tid_lid = {}
for r in ticks:
    for tid, lid in r['tidlid'].items():
        dev_tid_lid[(lid[:4], tid)] = lid  # lid 前 4 = uid 后 4(出生设备)

def belief_at(ts):
    i = bisect.bisect_right(tick_ts, ts) - 1
    return ticks[i] if i >= 0 else (ticks[0] if ticks else None)

# ── 2) window.json 双雷达 raw track 帧 + 离散事件 ────────────────────────────
win = json.load(open(f'{case_dir}/window.json'))
meta = {d['device_uid']: d['device_type'] for d in json.load(open(f'{case_dir}/meta.json'))['devices']}
# 只取本房雷达(room_pfx 下);用 spatial 无法在此判,改用:出现在 xray lid 的设备 + 发事件的设备都收,
# 但 raw track 行只收 belief tick 覆盖时段内、且该设备确属本房 → 用设备清单过滤(调用方保证 case=本 unit)。
EVT_RDR = {'EnterRoom','ExitRoom','Fall','InBed','LeftBed','Walking'}
rows = []  # (ts, prio, dev, tid, lid, pose, z, bed, event)
for r in win:
    u4 = r['device_uid'][-4:].upper(); dtype = meta.get(r['device_uid'],'?')
    cat = r['category']; ts = r['timestamp']
    if allow is not None and u4 not in allow: continue
    if cat == 'track' and dtype == 'Radar':
        for t in (r['data_value'] or []):
            if not isinstance(t, dict) or 'track_id' not in t: continue
            tid = t['track_id']
            if tid in PLACEHOLDER_TIDS:
                # 不滤:firmware 无目标/心跳帧(88/11)也出行(标 no-target),否则 lost coast 整段在报告里消失。
                rows.append([ts, 1, u4, tid, '-', '88', '-', None, 'no-target(88)'])
                continue
            rows.append([ts, 1, u4, tid, dev_tid_lid.get((u4, tid), '-'),
                         pose_name(t.get('pose',0)), t.get('position_z',0), None,
                         pose_name(t.get('pose',0))])
    elif cat in EVT_RDR and dtype == 'Radar':
        evtid = next((t['track_id'] for t in (r['data_value'] or [])
                      if isinstance(t, dict) and t.get('track_id') is not None), None)
        evlid = dev_tid_lid.get((u4, evtid), '-') if evtid is not None else '-'
        rows.append([ts, 0, u4, f'E{evtid}' if evtid is not None else 'E', evlid, '-', 0, None, f'{cat}(rdr)'])
    elif cat == 'number_people' and dtype == 'Radar':
        for t in (r['data_value'] or []):
            npv = t.get('number_people', t.get('count', '?'))
            rows.append([ts, 0, u4, 'E', '-', '-', 0, None, f'np={npv}' + ('  ★0' if npv == 0 else '')])
    elif cat in ('InBed','LeftBed') and dtype == 'Sleepad':
        rows.append([ts, 0, u4, 'E', '-', '-', 0, None, f'{cat}(pad)'])

# sleepad monitor 逐帧(window_sleepad.json;脚本原只读 window.json → sleepad 流完全不可见)：
#   bed_status/HR/RR/body_move 出行,与 radar/belief 同时间轴交织(看 InBed 接力 vs 本房 lost)。
try:
    sp = json.load(open(f'{case_dir}/window_sleepad.json'))
except (FileNotFoundError, json.JSONDecodeError):
    sp = []
for r in sp:
    if r.get('category') != 'track': continue
    u4 = r['device_uid'][-4:].upper()
    if allow is not None and u4 not in allow: continue
    for t in (r['data_value'] or []):
        bs = t.get('bed_status')
        bedlab = 'InBed' if bs == 0 else ('LeftBed' if bs == 1 else f'bs{bs}')
        ev = f'pad {bedlab} HR={t.get("heart_rate")} RR={t.get("respiratory_rate")} mv={t.get("body_move")} turn={t.get("turn_over")}'
        rows.append([r['timestamp'], 1, u4, t.get('track_id'), '-', 'pad', '-', None, ev])

# alarm.json（设备直发告警 + evidence；X 光第三块 monitor/event/alarm 齐）：出行。
try:
    alarms = json.load(open(f'{case_dir}/alarm.json'))
except (FileNotFoundError, json.JSONDecodeError):
    alarms = []
for r in alarms:
    u4 = r.get('device_uid', '????')[-4:].upper()
    if allow is not None and u4 not in allow: continue
    rows.append([r['timestamp'], 0, u4, 'A', '-', '-', 0, None, f"{r.get('category','alarm')}(ALARM)"])

# 不漏任何一秒：覆盖秒之外,每秒补一行 held(no frame),belief 取最近 tick(冻结态)。lost 期固件无帧的空档也成行。
if ticks:
    secs = [int(r[0] // 1000) for r in rows] + [ticks[0]['ts'] // 1000, ticks[-1]['ts'] // 1000]
    covered = {int(r[0] // 1000) for r in rows}
    for s in range(int(min(secs)), int(max(secs)) + 1):
        if s not in covered:
            rows.append([s * 1000, 9, '-', '-', '-', '-', '-', None, '(no frame, held)'])

rows.sort(key=lambda x: (x[0], x[1]))

# ── 3) 渲染 ────────────────────────────────────────────────────────────────
# s_marg 9 态序: [Empty0 Bed1 Sit2 OpenFloor3 Bath4 Fallen5 BlindRest6 BlindOpen7 Left8]
SM = {'SFall':5, 'SBed':1, 'SOpen':3, 'SBliR':6, 'SEmpt':0, 'SLeft':8}
ROOMKEY = {'SFall':'Fallen', 'SBed':'Bed', 'SOpen':'OpenFloor', 'SBliR':'BlindRest', 'SEmpt':'Empty', 'SLeft':'Left'}

def belief_cols(b, lid):
    """该行信念: 有 lid 且本 tick 该轨在 dbn 里 → 用 per-track s_marg(src=trk);否则回退房级(src=room)。
    返回 (src, preal, {SFall..SLeft})。"""
    dd = b['dbn_by_lid'].get(lid) if lid and lid != '-' else None
    if dd is not None:
        sm = dd['s_marg']
        return 'trk', dd.get('p_real', 0.0), {k: sm[i] for k, i in SM.items()}
    sd = b['sdist']
    return 'room', None, {k: sd.get(ROOMKEY[k], 0) for k in SM}

hdr = (f"{'time':8} {'dev.tid':8} {'lid':13} {'pose':7} {'z':4} {'bed':8} "
       f"{'event':18} {'src':4} {'pR':4} {'top':10} {'nr':3} {'still':5} {'SFall':5} {'SBed':5} {'SOpen':5} "
       f"{'SBliR':5} {'SEmpt':5} {'SLeft':5}")
out = []
fall_ts = [r['timestamp'] for r in win if r['category']=='Fall']
out.append(f"# {case_dir.split('/')[-1]} — 每 tick belief 时间线 (room {room_pfx}, TZ {TZ.key})")
out.append("")
out.append("dev.tid=uid后4.track_id(雷达 raw track 帧,**两台雷达都出行**)。lid=引擎采用的 base track 出生身份。")
out.append("**belief 列现为 per-track**:src=trk → 该 lid 自己的 s_marg(pR=该轨 p_real);src=room → 该行无 lid,回退房级 s_dist。")
out.append("于是「摔的那条轨自己的 SFall 是否起来」一眼可见(对照 top=房级裁决态)。")
out.append("")
out.append("```")
out.append(hdr)
for row in rows:
    ts, prio, dev, tid, lid, pose, z, _bed, event = row
    b = belief_at(ts)
    if b is None: continue
    src, preal, sm = belief_cols(b, lid)
    pr = f"{preal:.2f}" if preal is not None else '-'
    devtid = f"{dev}.{tid}"
    nr = b.get('n_r')
    nrs = str(nr) if nr is not None else '-'
    line = (f"{hhmmss(ts):8} {devtid:8} {lid:13} {pose:7} {z:<4} {b['bed']:8} "
            f"{event:18} {src:4} {pr:4} {b['top']:10} {nrs:3} {b['still']:<5} "
            f"{sm['SFall']:.2f}  {sm['SBed']:.2f}  {sm['SOpen']:.2f}  "
            f"{sm['SBliR']:.2f}  {sm['SEmpt']:.2f}  {sm['SLeft']:.2f}")
    out.append(line)
out.append("```")

# ── 原始 stream 子表（每雷达逐帧 raw track;x/y/Δ 对照上方 belief 的 stillbox;tid=88 标 no-target）──
out.append("")
out.append("## 原始 stream（每雷达逐帧 raw track；x/y/Δ 对照上方 belief 的 stillbox）")
out.append("```")
out.append(f"{'time':12} {'dev.tid':9} {'pose':6} {'area':4} {'x':6} {'y':6} {'z':5} {'conf':4} {'Δcm':5}")
def hmsms(ms): return datetime.datetime.fromtimestamp(ms/1000, TZ).strftime('%H:%M:%S.%f')[:-3]
raw_by_dev = {}
for r in win:
    if r['category'] != 'track' or meta.get(r['device_uid'],'?') != 'Radar': continue
    u4 = r['device_uid'][-4:].upper()
    if allow is not None and u4 not in allow: continue
    raw_by_dev.setdefault(u4, []).append((r['timestamp'], r.get('data_value') or []))
for u4 in sorted(raw_by_dev):
    last = None
    for ts, dv in sorted(raw_by_dev[u4], key=lambda z: z[0]):
        d = dv[0] if dv else None
        tid = d.get('track_id') if d else 88
        if not d or tid in PLACEHOLDER_TIDS:
            out.append(f"{hmsms(ts):12} {u4+'.'+str(tid):9} {'88':6} {'-':4} {'-':6} {'-':6} {'-':5} {'-':4} {'-':5}")
            continue
        x, y = d.get('position_x'), d.get('position_y')
        dd = str(int(((x-last[0])**2+(y-last[1])**2)**0.5)) if last and isinstance(x,int) else ''
        if isinstance(x,int): last = (x,y)
        out.append(f"{hmsms(ts):12} {u4+'.'+str(tid):9} {pose_name(d.get('pose',0)):6} "
                   f"{str(d.get('area_id')):4} {str(x):6} {str(y):6} {str(d.get('position_z')):5} "
                   f"{str(d.get('track_confidence')):4} {dd:5}")
    out.append("")
out.append("```")
out.append("")
fired = sum(1 for t in ticks if t['fire'])
out.append(f"**汇总**: xray tick {len(ticks)} | fire {fired} | Fall 事件 {len(fall_ts)} "
           f"({', '.join(hhmmss(t) for t in fall_ts)}) | 结论 = "
           f"{'FN(看到 Fall 但 fire=0)' if fired==0 and fall_ts else ('fire 命中' if fired else '无 Fall 无 fire')}")

dst = f'{case_dir}/track_event_timeline.md'
open(dst, 'w').write('\n'.join(out) + '\n')
print(f"写出 {dst} | {len(rows)} 行 | xray tick {len(ticks)} fire {fired} | Fall {len(fall_ts)}")

#!/usr/bin/env python3
# timeline_from_xray.py — case 重放后,合并 [Xsensor.log xsensor_xray 房级 belief] + [window.json 双雷达 raw track/事件]
# + [sleepad 床事件] → track_event_timeline.md(每 tick 全景,格式同 doc/cases/*/track_event_timeline.md)。
#
# 用法: timeline_from_xray.py <case 目录> <room_prefix(如 fd00:0:3:111:3:100)> [xsensor.log]
#
# 多雷达同房:窗口里**每个雷达的 raw track 帧都出一行**(dev.tid=uid后4.track_id),belief 列(top/SFall..)取房级
# xray 最近 tick(房只按引擎实际用的 base track 跑信念)。于是"某雷达看到 FALL 但房信念没动"的 FN 一眼可见。
import sys, json, bisect, datetime, zoneinfo
from collections import defaultdict

case_dir = sys.argv[1].rstrip('/')
room_pfx = sys.argv[2]
log_path = sys.argv[3] if len(sys.argv) > 3 else '/home/wisefido/owl/log/Xsensor.log'
# 第 4 参=本房设备 uid 后 4(逗号分隔),过滤掉同 unit 其它房的雷达;省略=全收。
allow = set(s.upper() for s in sys.argv[4].split(',')) if len(sys.argv) > 4 else None
PLACEHOLDER_TIDS = {88, 11}  # firmware 心跳/无坐标占位 track,非真人
TZ = zoneinfo.ZoneInfo('America/Denver')

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
            if tid in PLACEHOLDER_TIDS: continue
            rows.append([ts, 1, u4, tid, dev_tid_lid.get((u4, tid), '-'),
                         pose_name(t.get('pose',0)), t.get('position_z',0), None,
                         pose_name(t.get('pose',0))])
    elif cat in EVT_RDR and dtype == 'Radar':
        rows.append([ts, 0, u4, 'E', '-', '-', 0, None, f'{cat}(rdr)'])
    elif cat in ('InBed','LeftBed') and dtype == 'Sleepad':
        rows.append([ts, 0, u4, 'E', '-', '-', 0, None, f'{cat}(pad)'])

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
       f"{'event':18} {'src':4} {'pR':4} {'top':10} {'still':5} {'SFall':5} {'SBed':5} {'SOpen':5} "
       f"{'SBliR':5} {'SEmpt':5} {'SLeft':5}")
out = []
fall_ts = [r['timestamp'] for r in win if r['category']=='Fall']
out.append(f"# {case_dir.split('/')[-1]} — 卧室(09E7+D523 双雷达同房) 每 tick belief 时间线")
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
    line = (f"{hhmmss(ts):8} {devtid:8} {lid:13} {pose:7} {z:<4} {b['bed']:8} "
            f"{event:18} {src:4} {pr:4} {b['top']:10} {b['still']:<5} "
            f"{sm['SFall']:.2f}  {sm['SBed']:.2f}  {sm['SOpen']:.2f}  "
            f"{sm['SBliR']:.2f}  {sm['SEmpt']:.2f}  {sm['SLeft']:.2f}")
    out.append(line)
out.append("```")
out.append("")
fired = sum(1 for t in ticks if t['fire'])
out.append(f"**汇总**: xray tick {len(ticks)} | fire {fired} | Fall 事件 {len(fall_ts)} "
           f"({', '.join(hhmmss(t) for t in fall_ts)}) | 结论 = "
           f"{'FN(看到 Fall 但 fire=0)' if fired==0 and fall_ts else ('fire 命中' if fired else '无 Fall 无 fire')}")

dst = f'{case_dir}/track_event_timeline.md'
open(dst, 'w').write('\n'.join(out) + '\n')
print(f"写出 {dst} | {len(rows)} 行 | xray tick {len(ticks)} fire {fired} | Fall {len(fall_ts)}")

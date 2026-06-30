#!/usr/bin/env python3
# case_pg_timeline.py — 把 case 的 window.json(=export 从 PG 拉的 event+monitor) 按时间顺序展平成
# 一份可复核的 pg_event_monitor.md(留 case 目录)。读标准产物,不再敲 PG(规则 #5)。
#
# 用法: case_pg_timeline.py <case 目录> [CASE_TZ 默认 America/Denver]
import os, sys, json, datetime, zoneinfo

case = sys.argv[1].rstrip('/')
TZ = zoneinfo.ZoneInfo(sys.argv[2] if len(sys.argv) > 2 else 'America/Denver')
POSE = {0:'init',1:'walk',2:'susfall',3:'sit',4:'stand',5:'FALL',6:'lying',7:'sitgnd',8:'sitgnd',9:'bedsit',12:'run'}

def hms(ms): return datetime.datetime.fromtimestamp(ms/1000, TZ).strftime('%H:%M:%S.%f')[:-3]

rows = []
for fn in ('window.json', 'window_sleepad.json'):
    p = os.path.join(case, fn)
    if not os.path.exists(p): continue
    for rec in json.load(open(p)):
        rows.append(rec)
rows.sort(key=lambda r: r['timestamp'])

out = [f"# {os.path.basename(case)} — PG event+monitor 时间线 (window.json 展平, TZ {TZ.key})", "",
       "time         cat            uid4 detail", "```"]
for r in rows:
    t = hms(r['timestamp']); cat = r['category']; u4 = r['device_uid'][-4:]
    for e in (r['data_value'] if isinstance(r['data_value'], list) else [r['data_value']]):
        if cat == 'track':
            out.append(f"{t} track          {u4} tid={e.get('track_id')} {POSE.get(e.get('pose'),'p'+str(e.get('pose')))} "
                       f"({e.get('position_x')},{e.get('position_y')}) z={e.get('position_z')} area={e.get('area_id')} "
                       f"cnt={e.get('track_count')} conf={e.get('track_confidence')}")
        elif cat == 'heart':
            out.append(f"{t} heart          {u4} tid={e.get('track_id')} hr={e.get('heart_rate')} rr={e.get('respiratory_rate')} pose={e.get('pose')}")
        elif cat == 'number_people':
            out.append(f"{t} number_people  {u4} np={e.get('number_people')} status={e.get('event_status')} tid={e.get('track_id')}")
        elif cat == 'activity':
            out.append(f"{t} activity       {u4} status={e.get('event_status')} lie={e.get('lie_duration')} "
                       f"walk={e.get('walk_duration')} stand={e.get('stand_duration')} dist={e.get('walk_distance')}")
        else:
            out.append(f"{t} {cat:<14} {u4} {json.dumps(e, ensure_ascii=False)}")
out.append("```")
dst = os.path.join(case, 'pg_event_monitor.md')
open(dst, 'w').write('\n'.join(out) + '\n')
print(f"写出 {dst} | {len(rows)} 条记录")

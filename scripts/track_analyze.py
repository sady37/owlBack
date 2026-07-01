#!/usr/bin/env python3
# track_analyze.py — 按 device_uid + 本地起止时间，直接查 PG（monitor_stream + event_log），
# 把 track/heart/事件按时间展平成一份可复核的 track 时间线，并附「track 分析」小结。
#
# 这是把 tools/case_pg_timeline.py 的 track 分析功能抽出来做成的独立脚本：不再依赖 case 导出
# 的 window.json，直接敲 PG（复用 tools/pg_dump_raw.sh 的查询），一条命令即可复盘任意设备/窗口。
#
# 用法:
#   scripts/track_analyze.py <device_uid> '<本地起>' '<本地止>' [tz=America/Denver] [--md 输出文件]
# 例:
#   scripts/track_analyze.py E598A2ACD523 '2026-06-30 18:40:00' '2026-06-30 18:57:00'
#
# DB 连接单源 owlBack/.env（DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME）。
import os, sys, json, subprocess, datetime, zoneinfo
from collections import defaultdict

# ── 参数 ──────────────────────────────────────────────────────────────────
raw = sys.argv[1:]
md_out = None
if '--md' in raw:
    i = raw.index('--md'); md_out = raw[i + 1]; del raw[i:i + 2]
args = [a for a in raw if not a.startswith('--')]
if len(args) < 3:
    sys.exit("用法: track_analyze.py <device_uid> '<本地起>' '<本地止>' [tz] [--md 文件]")
UID, START_LOCAL, END_LOCAL = args[0], args[1], args[2]
TZ = zoneinfo.ZoneInfo(args[3] if len(args) > 3 else 'America/Denver')

# firmware pose 语义（与 tools/*.py 一致）
POSE = {0: 'init', 1: 'walk', 2: 'susfall', 3: 'sit', 4: 'stand', 5: 'FALL', 6: 'lying',
        7: 'sitgnd', 8: 'sitgnd', 9: 'bedsit', 10: 'bedsit', 11: 'bedsit', 12: 'run'}
def pose_name(p): return POSE.get(p, f'p{p}')
TELEPORT_CM = 200   # 单帧位移 ≥ 此值 = 疑似 firmware 瞬移干扰轨（见记忆 teleport_interference_purge）

# ── 1) 读 .env 取 DB 连接 ───────────────────────────────────────────────────
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
env = {}
envfile = os.path.join(ROOT, '.env')
if os.path.exists(envfile):
    for line in open(envfile):
        line = line.strip()
        if line and not line.startswith('#') and '=' in line:
            k, v = line.split('=', 1); env[k] = v
DB = dict(host=env.get('DB_HOST', '127.0.0.1'), port=env.get('DB_PORT', '5432'),
          user=env.get('DB_USER', 'postgres'), name=env.get('DB_NAME', 'owl_v2'),
          pw=env.get('DB_PASSWORD', 'postgres'))

def to_ms(local):
    dt = datetime.datetime.strptime(local, '%Y-%m-%d %H:%M:%S').replace(tzinfo=TZ)
    return int(dt.timestamp() * 1000)
START_MS, END_MS = to_ms(START_LOCAL), to_ms(END_LOCAL)

# ── 2) 查 PG（monitor_stream + event_log 合并，按 ts 升序）─────────────────────
SQL = f"""
WITH mon AS (
  SELECT ts, jsonb_build_object('_src','MON') || to_jsonb(m.*) AS row
  FROM monitor_stream m
  WHERE device_uid='{UID}'
    AND ts BETWEEN to_timestamp({START_MS}/1000.0) AND to_timestamp({END_MS}/1000.0)
), evt AS (
  SELECT ts, jsonb_build_object('_src','EVT') || to_jsonb(e.*) AS row
  FROM event_log e
  WHERE device_uid='{UID}'
    AND ts BETWEEN to_timestamp({START_MS}/1000.0) AND to_timestamp({END_MS}/1000.0)
)
SELECT row::text FROM (SELECT ts,row FROM mon UNION ALL SELECT ts,row FROM evt) u ORDER BY ts;
"""
proc = subprocess.run(
    ['psql', '-h', DB['host'], '-p', DB['port'], '-U', DB['user'], '-d', DB['name'], '-tA', '-c', SQL],
    env={**os.environ, 'PGPASSWORD': DB['pw']}, capture_output=True, text=True)
if proc.returncode != 0:
    sys.exit(f"psql 失败: {proc.stderr}")
recs = [json.loads(l) for l in proc.stdout.splitlines() if l.strip()]

def ms_of(r):
    # ts 形如 2026-07-01T00:40:00.325+00:00
    return int(datetime.datetime.fromisoformat(r['ts']).timestamp() * 1000)
def hms(ms): return datetime.datetime.fromtimestamp(ms / 1000, TZ).strftime('%H:%M:%S.%f')[:-3]

# ── 3) 展平成时间线 + 收集 per-track 帧 ─────────────────────────────────────
lines = []                       # (ms, prio, text)
tracks = defaultdict(list)       # track_id -> [(ms,x,y,z,pose,area,conf)]
for r in recs:
    ms = ms_of(r); u4 = r['device_uid'][-4:]
    pl = r.get('payload')
    items = pl if isinstance(pl, list) else ([pl] if pl else [])
    if r['_src'] == 'MON' and r.get('stream_type') == 'radar.track':
        for e in items:
            tid = e.get('track_id')
            x, y, z = e.get('position_x'), e.get('position_y'), e.get('position_z')
            pose, area, conf = e.get('pose'), e.get('area_id'), e.get('track_confidence')
            tracks[tid].append((ms, x, y, z, pose, area, conf))
            lines.append((ms, 1, f"{hms(ms)} track   {u4} tid={tid} {pose_name(pose):<7} "
                                 f"({x},{y}) z={z} area={area} cnt={e.get('track_count')} conf={conf}"))
    elif r['_src'] == 'MON' and r.get('stream_type') == 'radar.heart':
        for e in items:
            lines.append((ms, 1, f"{hms(ms)} heart   {u4} tid={e.get('track_id')} "
                                 f"hr={e.get('heart_rate')} rr={e.get('respiratory_rate')} pose={e.get('pose')}"))
    elif r['_src'] == 'EVT':
        kind = r.get('event_kind', '?')
        for e in items:
            det = ''
            if kind == 'number_people':
                np = e.get('number_people', '?'); det = f"np={np}" + ('  ★0' if np == 0 else '')
            elif kind == 'activity':
                det = (f"status={e.get('event_status')} lie={e.get('lie_duration')} "
                       f"walk={e.get('walk_duration')} stand={e.get('stand_duration')} dist={e.get('walk_distance')}")
            else:
                det = f"status={e.get('event_status')} tid={e.get('track_id')} np={e.get('number_people','')}"
            star = '  ◀━━ ' + kind.upper() if kind in ('EnterRoom', 'ExitRoom', 'Fall') else ''
            lines.append((ms, 0, f"{hms(ms)} EVENT   {u4} {kind:<13} {det}{star}"))
lines.sort(key=lambda x: (x[0], x[1]))

# ── 4) track 分析：每条轨的存活段、瞬移跳变、pose 迁移、静止段 ─────────────────
def dist(a, b):
    if None in (a[1], a[2], b[1], b[2]): return 0
    return ((a[1] - b[1]) ** 2 + (a[2] - b[2]) ** 2) ** 0.5

analysis = []
for tid in sorted(tracks, key=lambda t: (t is None, t)):
    fr = tracks[tid]
    if not fr: continue
    dur = (fr[-1][0] - fr[0][0]) / 1000.0
    poses = [pose_name(f[4]) for f in fr]
    # pose 迁移压缩
    trans, prev = [], None
    for p in poses:
        if p != prev: trans.append(p); prev = p
    # 瞬移跳变
    jumps = []
    for i in range(1, len(fr)):
        d = dist(fr[i - 1], fr[i])
        if d >= TELEPORT_CM:
            jumps.append(f"{hms(fr[i][0])} Δ{d:.0f}cm ({fr[i-1][1]},{fr[i-1][2]})→({fr[i][1]},{fr[i][2]})")
    # 最长静止段（连续帧位移 <10cm）
    still_run = still_max = 0
    for i in range(1, len(fr)):
        if dist(fr[i - 1], fr[i]) < 10:
            still_run += (fr[i][0] - fr[i - 1][0]) / 1000.0
            still_max = max(still_max, still_run)
        else:
            still_run = 0
    fell = any(f[4] == 5 for f in fr)
    analysis.append((tid, fr, dur, trans, jumps, still_max, fell))

# ── 5) 输出 ────────────────────────────────────────────────────────────────
out = [f"# track 分析 — {UID}  {START_LOCAL} .. {END_LOCAL}  (TZ {TZ.key})",
       f"# PG: {DB['name']}@{DB['host']}  帧数={sum(len(v) for v in tracks.values())}  轨数={len(tracks)}  行={len(lines)}",
       "", "## 时间线", "```",
       "time         cat     uid4 detail"]
out += [t for _, _, t in lines]
out += ["```", "", "## per-track 分析", "```"]
out.append("tid  帧数  存活s  最长静止s  跌倒?  pose迁移")
for tid, fr, dur, trans, jumps, still_max, fell in analysis:
    out.append(f"{str(tid):<4} {len(fr):<5} {dur:<6.0f} {still_max:<10.0f} "
               f"{'FALL' if fell else '-':<6} {' → '.join(trans[:12])}")
    for j in jumps:
        out.append(f"      ⚡瞬移 {j}")
out.append("```")
text = '\n'.join(out) + '\n'

if md_out:
    open(md_out, 'w').write(text)
    print(f"写出 {md_out} | 行={len(lines)} 轨={len(tracks)}")
else:
    sys.stdout.write(text)

package playback

import (
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
)

// WriteHTML 把 snapshots 渲染成单文件 HTML 滑动查看器写入 w。
// 与原 cmd/roomengine-playback 的 writeHTML 行为一致。
func WriteHTML(w io.Writer, roomID string, snaps []Snapshot) error {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>RoomEngine Playback: `)
	sb.WriteString(html.EscapeString(roomID))
	sb.WriteString(`</title><style>
body{margin:0;background:#1d1f23;color:#ddd;font-family:system-ui,-apple-system,sans-serif}
header{padding:8px 14px;background:#26282d;display:flex;align-items:center;gap:14px;flex-wrap:wrap}
h1{font-size:14px;margin:0;font-weight:600}
button{background:#3a8;color:#fff;border:0;padding:6px 14px;border-radius:4px;cursor:pointer;font-size:13px}
button:hover{background:#4b9}
.label{font-size:13px;font-variant-numeric:tabular-nums;color:#aac}
input[type=range]{flex:1;min-width:300px;accent-color:#3a8}
#stage{display:flex;justify-content:center;padding:10px 0;background:#1d1f23}
#stage svg{max-width:96vw;max-height:88vh}
.hint{font-size:11px;color:#888}
</style></head><body>
<header>
  <h1>Playback: `)
	sb.WriteString(html.EscapeString(roomID))
	sb.WriteString(`</h1>
  <button id="play">▶ Play</button>
  <input type="range" id="slider" min="0" max="`)
	if len(snaps) > 0 {
		fmt.Fprintf(&sb, "%d", len(snaps)-1)
	} else {
		sb.WriteString("0")
	}
	sb.WriteString(`" value="0"/>
  <div class="label" id="counter"></div>
  <div class="label" id="ts"></div>
  <div class="hint">键盘 ← → 步进；空格 播放/暂停</div>
</header>
<div id="stage"></div>
<script>
const FRAMES = [`)
	for i, f := range snaps {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(strconv.Quote(f.SVG))
	}
	sb.WriteString(`];
const LABELS = [`)
	for i, f := range snaps {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Quote(f.Label))
	}
	sb.WriteString(`];
const stage=document.getElementById('stage');
const slider=document.getElementById('slider');
const counter=document.getElementById('counter');
const tsEl=document.getElementById('ts');
const playBtn=document.getElementById('play');
let playing=false;
let timer=null;

function render(i){
  i=Math.max(0,Math.min(FRAMES.length-1,i|0));
  stage.innerHTML=FRAMES[i];
  slider.value=i;
  counter.textContent='frame '+(i+1)+' / '+FRAMES.length;
  tsEl.textContent=LABELS[i];
}
slider.addEventListener('input',e=>render(+e.target.value));
playBtn.addEventListener('click',()=>{
  playing=!playing;
  playBtn.textContent=playing?'⏸ Pause':'▶ Play';
  if(playing){
    timer=setInterval(()=>{
      let v=+slider.value+1;
      if(v>=FRAMES.length){playing=false;playBtn.textContent='▶ Play';clearInterval(timer);return;}
      render(v);
    },200);
  } else { clearInterval(timer); }
});
document.addEventListener('keydown',e=>{
  if(e.key==='ArrowLeft'){render(+slider.value-1);}
  else if(e.key==='ArrowRight'){render(+slider.value+1);}
  else if(e.key===' '){playBtn.click();e.preventDefault();}
});
render(0);
</script>
</body></html>`)

	_, err := io.WriteString(w, sb.String())
	return err
}

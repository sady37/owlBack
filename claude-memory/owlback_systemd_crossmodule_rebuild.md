---
name: owlback-systemd-crossmodule-rebuild
description: owlback-run-service.sh 已支持跨 module stale-check（go.mod replace 本地路径），owl-common/wisefido-sensor 等共享 module 改动会触发依赖服务 rebuild
metadata: 
  node_type: memory
  type: project
  originSessionId: 5c181f1d-cf48-440b-bd17-e5ae711213b7
---

2026-05-16 修：[owlback-run-service.sh](../../../owl/owlBack/scripts/systemd/owlback-run-service.sh) `owlback_go_exec` 的 stale-check 原来只看 service 自己目录里的 `.go` 文件，**跨 module 改动（如 owl-common）不触发 rebuild**。表现：改完 owl-common 后 `sudo systemctl restart owlback.<svc>` 跑的还是旧二进制，bug 看似不修。

## 现在的行为

stale-check 额外解析 `go.mod` 里 `replace xxx => ../yyy` 形式的本地依赖，对每个 target dir 也跑 `find -newer` 判定。

```bash
# pseudo:
for target in $(awk '/^replace.*=>/' go.mod | extract right-of-=>); do
  [[ $target starts with . or / ]] || continue   # 跳过 module path replace
  find "$dir/$target" -name '*.go' -newer "$bin" → stale=1
done
```

## 涉及的 replace 链（2026-05-16 当前）

- wisefido-sensor → ../owl-common
- wisefido-cardagg → ../owl-common
- wisefido-data → ../owl-common + ../wisefido-sensor + ../wisefido-qinglan
- wisefido-iot / wisefido-qinglan / wisefido-sleepace → ../owl-common

## 验证

```bash
touch /home/wisefido/owl/owlBack/owl-common/card/alarm_db.go
sudo systemctl restart owlback.cardagg.service  # 应触发 rebuild
ls -la /home/wisefido/owl/owlBack/wisefido-cardagg/.bin/wisefido-cardagg  # 时间应是 just-now
```

## 历史教训

2026-05-15 platform-agent-addressing 战役里改了 `owl-common/card/alarm_db.go` 的 SQL 但 cardagg `.bin/wisefido-cardagg` 没自动重编，导致 fan-out fix 看似失败；手动 `go build -o .bin/wisefido-cardagg .` 才生效。

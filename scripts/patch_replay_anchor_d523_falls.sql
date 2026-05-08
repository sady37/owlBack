-- 一次性 patch：给两条历史 D523 engine_lost_fall 在 trigger_data.evidence
-- 注入 replay_anchor_ms，验证前端 RePlay 是否会用 anchor 而非 triggered_at 作中心。
--
-- 07:35 fall (PDT) — frozen_start_ms=0 → anchor = ts_ms - wait_ms - 30s
--   ts_ms = 1778164527640 (07:35:27.640)
--   wait_ms = 300000
--   anchor = 1778164197640 (07:29:57.640 PDT)
--
-- 06:22 fall (PDT) — frozen_start_ms 非零 → 直接用
--   anchor = 1778159721420 (06:15:21.420 PDT)
--
-- jsonb_set: 若 evidence 已存在则只插 replay_anchor_ms；createIfMissing=true。

UPDATE alarm_events
SET trigger_data = jsonb_set(
  COALESCE(trigger_data, '{}'::jsonb),
  '{evidence,replay_anchor_ms}',
  to_jsonb(1778164197640::bigint),
  true
)
WHERE event_id = 'c89c45c9-907a-4b29-8741-19c0fc32a86a';

UPDATE alarm_events
SET trigger_data = jsonb_set(
  COALESCE(trigger_data, '{}'::jsonb),
  '{evidence,replay_anchor_ms}',
  to_jsonb(1778159721420::bigint),
  true
)
WHERE event_id = 'fcf378f6-93b0-49d3-96e4-d2d42d28da7d';

-- 验证
SELECT
  event_id,
  triggered_at,
  trigger_data->'evidence'->>'replay_anchor_ms' AS replay_anchor_ms,
  trigger_data->'evidence'->>'frozen_start_ms' AS frozen_start_ms,
  trigger_data->'evidence'->>'wait_ms' AS wait_ms
FROM alarm_events
WHERE event_id IN (
  'c89c45c9-907a-4b29-8741-19c0fc32a86a',
  'fcf378f6-93b0-49d3-96e4-d2d42d28da7d'
);

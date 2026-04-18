# Card Overview UI Specification

## Card Types

| Type | Description | Binding |
|------|-------------|---------|
| ActiveBedCard | Bed-level monitoring card | bed_id + unit_id |
| UnitCard | Unit-level card (unbound devices in unit) | unit_id only |
| DeviceCard | Standalone device card (card_id = device_id) | No unit/bed binding |

- DeviceCard is created automatically for devices not bound to any unit
- When device binds to a unit, DeviceCard is cleaned up and device joins unit card
- cardagg manages all device status via `card.device.[status]` key structure

## Card Layout

- Card size: 270 x 240px + 15px margin = 300 x 270px per slot
- Grid: flex-wrap, left-aligned
- Page size: dynamically calculated from viewport (cols x rows)
- Recalculates on window resize and sidebar collapse/expand (ResizeObserver)

### Card Sections

```
+------------------------------------------+
| S1: card_name    [offline] [AlarmBell]   |
|     card_address                          |
|------------------------------------------|
| S2: Main content                         |
|   ActiveBedCard: bed_status + HR/RR      |
|   UnitCard: person_count + postures      |
|   DeviceCard: rendered as UnitCard       |
|------------------------------------------|
| S3: status lines + visitor info          |
|------------------------------------------|
| PopAlarm bar (red/orange/yellow)         |
+------------------------------------------+
```

## Pagination

### Controls

Inline after Focus button: `|<< [current_page] >>| total_cards`

- `|<<` jump to first page
- `>>|` jump to last page
- Scroll wheel: up/down page
- Arrow keys: Left/Up/PageUp = prev, Right/Down/PageDown = next, Home/End = first/last

### Dynamic Page Size

- Calculated from available viewport: `cols = floor(width / 300)`, `rows = floor(height / 270)`
- Adapts to sidebar collapse/expand via ResizeObserver
- No fixed page size constant

## Auto-Scroll (Auto-Rotate)

| Parameter | Value |
|-----------|-------|
| Page interval | 6 seconds |
| Enter delay | 120 seconds after page load or last interaction |
| Resume delay | 120 seconds after any interaction |

### Interaction Detection

- Pagination buttons, scroll wheel, arrow keys
- Mouse click or movement in card area (2s throttle)
- Any interaction resets the 120s timer

### Alarm Interruption

- New alarm immediately pauses auto-scroll
- Jumps to alarm card's page
- Stays paused for 120 seconds, then resumes

## Alarm Display

### Filter Button

- Label: **Alarm** (unified with APP, previously "Unhandled")
- Sorted by `triggered_at` descending (newest alarm first)

### AlarmBell (S1 Icon)

- Blinks on **new** pop alarm only (not historical unhandled)
- Duration: **15 seconds**, then auto-stops
- Also stops immediately when alarm is handled (pop_alarm cleared)
- CSS animation: 0.8s cycle, opacity 1 <-> 0.2, scale 1 <-> 1.3
- Triggered via same path as alarm sound (fresh window check: 15s)

### Alarm Sound

- Plays on new pop alarm arrival
- Duration: ~3 seconds
- L1 (level 0-2: EMERG/ALERT/CRITICAL): urgent tone
- L2 (level 3-4: ERROR/WARNING): standard tone
- Deduplication: per card per event_id, sessionStorage backup

### Pop Alarm Bar

- Bottom of card, absolute positioned (z-index: 999)
- Level 0-2 (Red): white "Handle" button
- Level 3 (Deep Orange): dynamic theme
- Level 4 (Light Orange): orange "Handle" button

## Lock Screen (Three-Level State Machine)

```
Active ──(inactivity)──> Passive ──(long inactivity)──> Locked
  ^                        |                               |
  |     (PIN verify)       |                               |
  +────────────────────────+         (re-login)            |
  +────────────────────────────────────────────────────────+
```

### State Definitions

| State | Timeout | UI | Operations | Unlock |
|-------|---------|-----|------------|--------|
| Active | - | Normal | Full access | - |
| Passive | 5 min prod / 60s test | Read-only, opacity 0.7, cursor not-allowed | View only, no click/edit | 4-digit PIN |
| Locked | 4 hours prod / 5 min test | Full lock overlay | Nothing | Re-login |

### Passive Trigger Conditions (any one triggers)

1. **Inactivity timeout**: no mouse/keyboard activity for `passiveTimeout`
2. **Window blur timeout**: browser tab hidden or window lost focus for `passiveTimeout`
3. Both use countdown mechanism (`passiveTargetTime = now + timeout`)

### Passive Behavior

- Card display continues rendering (cards visible but dimmed)
- **Auto-scroll continues** (not affected)
- **Alarm sound continues** (not affected)
- **Alarm jump continues** (not affected)
- **SSE stream continues** (realtime data keeps flowing)
- Click anywhere shows PIN modal
- Correct PIN returns to Active, resets all timers

### Locked Behavior

- Full-screen lock overlay, all content hidden
- SSE stream closed
- Requires full re-login (username + password)

### Configuration

| Env Variable | Default (prod) | Default (test) |
|-------------|----------------|----------------|
| `VITE_LOCK_SCREEN_PASSIVE_TIMEOUT_MS` | 300000 (5 min) | 60000 (60s) |
| `VITE_LOCK_SCREEN_LOCKED_TIMEOUT_MS` | 14400000 (4 hr) | 300000 (5 min) |
| `VITE_LOCK_SCREEN_TEST_MODE` | - | `true` enables test timeouts |

### State Persistence

- Saved to `sessionStorage` on every state change
- Restored on page reload (if target times haven't expired)
- `logout` clears all lock state

### Interaction with Auto-Scroll

| Event | Active | Passive | Locked |
|-------|--------|---------|--------|
| Auto-scroll runs | Yes (after 120s idle) | Yes | No |
| Alarm sound plays | Yes | Yes | No |
| Alarm jumps page | Yes | Yes | No |
| Mouse resets scroll timer | Yes | No (dimmed) | No |
| PIN unlock | - | Returns to Active | - |

## SSE (Server-Sent Events)

- Endpoint: `/data/api/v1/data/vital-focus/cards/stream`
- Data rate: 0.5 Hz (realtime vitals)
- Event types: message (vitals), card_status (state changes), card_change (add/delete), connected, ready
- Retry: max 5, exponential backoff (1s -> 16s)
- Watch IDs: Admin = all cards, Focus = selected, Non-admin = first branch
- View IDs: current page cards only (optimized bandwidth)

## Data Architecture

```
Frontend (3 reactive Maps):
  cardMap      : card_id -> CardStatic     (name, address, devices, residents)
  realtimeMap  : card_id -> CardRealTime   (HR, RR, sleep_stage @ 0.5Hz)
  statusMap    : card_id -> CardStatus     (bed_state, room_state, alarm_state, device_status)

Backend:
  cards table  -> CardStatic (SQL query, paginated)
  cardagg      -> CardRealTime + CardStatus (SSE stream via Redis)
```

## Future: iPad Monitor Mode (TODO)

For elder care facilities using iPad as wall-mounted monitor:

- Manual toggle button to enter/exit Monitor Mode
- Enter: hide sidebar, full-screen cards, auto-scroll
- Alarm: jump to alarm page + bell blink 15s + sound 3s
- Alarm list sorted by triggered_at (same as web)
- Exit: tap button or PIN unlock

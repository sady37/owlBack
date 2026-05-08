# OwlCare 2.0
## AI-Assisted Elder Care Monitoring System

OwlCare 2.0 is a next-generation AI care-assistance platform built for senior living and elder care facilities.

While Owl Monitor 1.0 delivered "device-state monitoring," 2.0 has been redesigned as:

> **A Care Decision Support + Risk Prevention + Compliance system**

Through multi-sensor fusion, a self-learning room engine, and behavioral time-series analysis, the system helps care teams identify risk earlier, intervene sooner, and reduce both care omissions and regulatory exposure.

---

# Core Upgrade — From 1.0 to 2.0

## Owl Monitor 1.0

Provided foundational safety monitoring:

- Radar fall / sit-on-floor detection
- Long-stay bathroom alerts
- SleepPad HR / RR vital monitoring
- Out-of-bed alerts
- Turn / reposition reminders
- In-bed seizure alerts

Capabilities centered on **single-device, single-event triggers**.

---

# OwlCare 2.0

## From "Device Alerts" to "Behavior Understanding"

### Multi-Sensor Fusion AI

The system fuses:

- Radar (spatial behavior)
- SleepPad (bed-surface pressure)
- Vitals (HR / RR)
- Activity-window time-series model

Rather than relying on any single sensor, 2.0 evaluates real risk across multiple dimensions.

For example:

- Did the resident actually leave the bed?
- Is there sustained activity?
- Is this a sensor artifact?
- Does this fit the resident's normal routine?

---

## Self-Learning Spatial Cognition (RoomEngine)

Every room is automatically learned over time:

- Room geometry
- Common activity zones
- Ghost-reflection / interference areas
- Daily movement patterns

No complex commissioning or manual calibration required.

The longer the system runs, the better it understands each room.

---

## Cell-Level Adaptive Thresholds

The system learns:

> "What activity is normal in this specific spot?"

For example:

- Long lying on the sofa
- Long standing in the kitchen
- HVAC-induced micro-motion
- Oxygen concentrator vibration
- Wall-reflection ghost tracks

False-alert probability is automatically reduced for these zones.

---

# Key Capability Upgrades

---

# Bathroom Risk Detection
## From "Status Monitoring" to "High-Risk Behavior Recognition"

The bathroom is the highest-risk and hardest-to-monitor area in any senior living facility.

Privacy rules out conventional cameras.

OwlCare 2.0 introduces a new **StayFSM** behavior model that recognizes:

- Prolonged stay
- Prolonged standing
- Lack of meaningful activity response
- Abnormal stay trends

The system automatically generates:

- A timeline of the stay
- A risk-state assessment
- A response/intervention record

Helping facilities catch potential incidents before they escalate.

---

# Nighttime Fall Detection
## Continuous Behavior Analysis via the Activity-Window Model

2.0 no longer relies on a single "fall motion" trigger.

The system reconstructs the full activity arc by combining:

- Out-of-bed
- Walking
- Prolonged standing
- Posture changes
- Vital-sign changes

For example:

- When did the resident leave the bed?
- How did they move?
- Where did they stop?
- When did the anomaly occur?
- When did staff respond?

It also fuses:

- SleepPad pressure data
- Radar spatial data
- Anti-interference engine

Effectively filtering:

- HVAC disturbance
- Oxygen concentrators
- Environmental vibration
- Multipath ghost reflections

Substantially eliminating ghost-track false positives in single-occupant rooms, and significantly reducing missed falls.

---

# AI-Assisted Care Interface
## Caregivers See "States," Not "Data"

The interface surfaces directly:

- In room?
- In bed?
- In bathroom?
- Inactive for an extended period?
- Visitor present?
- Last activity timestamp
- Risk-level changes

Helping care staff:

- Get a fast read on the whole floor
- Reduce rounding pressure
- Cut care omissions

---

# Mobile Care Support

Supports:

- iPhone app
- Apple Watch real-time alerts
- iPad multi-room monitoring

With:

- Auto-rotation views
- Auto-jump to the alerting room
- Real-time risk visibility

---

# Compliance & Operational Value

## Incident Review and SOP Support

The system supports:

- Fall-trajectory replay
- Rounding records
- Response-process records
- Alert-handling records

Helping facilities:

- Establish standardized care SOPs
- Pass regulatory surveys
- Reduce legal exposure

---

# Longitudinal Care Intelligence
## From "Recording Events" to "Observing Change"

A traditional alert system answers:

> "Is something happening right now?"

OwlCare 2.0 also answers:

> "Has this resident been quietly declining over the past 30 days?"

---

## Design Principle

Passive radar + bed sensor, observing 24×365 — no resident cooperation required, no extra staff workflow.

What the system accumulates is not raw data, but:

> **A personal baseline for every resident**

Any deviation from that baseline can then be identified as an **early signal**.

---

## Six Longitudinal Trend Capabilities

### 1. Mobility Trend

- Morning gait speed (the most sensitive frailty indicator)
- Sit-to-stand time
- Daily cumulative distance
- Nighttime bathroom visits

> Annualized gait-speed decline > 5% = active frailty progression — an objective basis for care-level escalation.

---

### 2. Stability Trend

Detects when a resident drifts from a "stable waveform" into a "fluctuating waveform":

- Nighttime HR / RR stability
- Sleep-efficiency consistency
- Gait-speed consistency

> The stable-to-drifting inflection point typically precedes a clinical event by 2–4 weeks.

---

### 3. Recovery Pattern

How long does it take a resident to return to resting state after getting up at night, using the bathroom, or short bursts of activity?

> A steadily lengthening recovery time = early signal of declining cardiovascular reserve.
> A home-setting exclusive — hospitals can only measure point-in-time.

---

### 4. Sleep & Restlessness

- Sleep efficiency and stages
- Pre-sleep restlessness index (posture-transition frequency)
- **Prolonged immobility without turning (NoTurnOver)** — a direct, actionable cue for pressure-injury prevention
- Abnormal body movement / restless-legs candidate signals

> Lack of repositioning is the #1 facility liability event — the system provides an independent, actionable signal.
> Pre-sleep restlessness = earliest-stage signal for delirium / cognitive fluctuation.

---

### 5. Behavior Pattern Trend

The system builds a 30-day personal-routine vector for every resident:

- Wake / sleep times
- Bathroom frequency and timing
- Room-stay distribution
- Kitchen presence at meal times

A sudden routine deviation = candidate signal for an acute event, depression, or cognitive decline.

---

### 6. Cardiovascular Vital Trend

- Nighttime HR / RR median and rhythm
- HR-RR coupling
- HR response during ambulation (chronotropic competence)
- Postprandial nap HR (autonomic-neuropathy proxy)

> Leading indicators (6–12 months) for heart failure, arrhythmia, and cardiovascular events.

---

## Trends vs. Alerts — Complementary, Not Substitutable

| Dimension | Alert System | Trend System |
|---|---|---|
| Time scale | Seconds — minutes | Weeks — months |
| Output | Events / interventions | Deviations / assessments |
| Primary user | Care staff (on shift) | Director of Nursing / Administrator (review) |
| Purpose | Reduce incidents | Justify care-level changes; communicate with families & regulators |

---

## Three-Tier Value to the Facility

| Audience | What they get |
|---|---|
| **Director of Nursing** | Continuous in-home clinical signals unavailable in hospital settings; evidence for care planning, assessment, and incident review |
| **Administrator** | Objective, signed monthly trend reports for care-level escalation, contract renewal, and board review |
| **Family Members** | Readable, downloadable monthly health-trend reports that build trust and reduce complaints and litigation risk |

---

## Important Boundary

The system **does not diagnose**. It outputs only:

- A personal baseline
- The magnitude of deviation
- Nurse-readable, explainable indicators

All clinical judgment and intervention remain with the care team.

---

# Damage-Free Installation

A new mounting base:

- No drilling
- No wall damage
- 10-second deployment
- Easy relocation

Built for the operational realities of senior living facilities.

---

# OwlCare 2.0 — Core Goal

Not a replacement for caregivers.

Rather:

> Helping caregivers identify risk earlier, reduce omissions, and produce traceable evidence of care.

---

# Summary

OwlCare 2.0 has evolved from:

> "A device-alert system"

into:

> **"An AI-Assisted Elder Care Risk Monitoring Platform"**

Through:

- Multi-sensor AI
- Behavioral spatial learning
- Time-series behavior analysis
- Care decision support

Helping senior living facilities:

- Improve care efficiency
- Reduce incident risk
- Strengthen compliance posture
- Increase resident safety and family trust

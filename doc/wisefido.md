# **WiseFido 2.0: Dignity Through Technology**
### **The Only Sensor Allowed in Every Room of an Aging American Home**

---

## **1. Executive Summary**

WiseFido 2.0 is an institutional-grade, **camera-free, wearable-free** monitoring system for the $100B+ US elder care market. We solved the two technical barriers that have kept mmWave radar out of senior living for a decade — **multipath ghosting** and **angular-resolution-induced fall miss** — through **multi-device fusion + self-learning AI**.

> **Cameras own home security. No one owns radar elder care.**
> A multi-billion-dollar market with **zero incumbent** — because three barriers stack: radar hardware, signal processing, and elder-scene data. We've crossed all three.

- **Status**: $200K self-funded R&D completed; 2.0 hardware in production
- **Traction**: Lighthouse pilot live at **WeCare (Denver)**; free OTA upgrade rolling out to 4 partner sites
- **Compliance**: HIPAA-grade PHI encryption (AES-256-GCM, dual-factor KMS) deployed; HITRUST / SOC 2 / CMS / state-level on roadmap
- **Objective**: Seeking **$1M seed** to convert Denver lighthouse into 10+ paid institutional contracts and lock in the 6–8 month time-to-data moat before any competitor can enter

---

## **2. The Problem: The "Accountability Gap"**

Senior living faces a brutal paradox: **regulators demand oversight, residents demand privacy, and insurers demand evidence**.

- **The Liability Black Box**: ~80% of falls happen in bedrooms and bathrooms — the exact spaces where cameras are ethically and legally prohibited
- **Insurance Crisis**: Without objective trajectory evidence, facilities lose "he-said-she-said" lawsuits and face skyrocketing liability premiums
- **Labor Shortage**: Manual night rounds are unreliable, invasive, and a leading cause of caregiver burnout
- **Wearable Failure**: Pendants and watches have ~30% true compliance among dementia residents — the population that needs them most

---

## **3. The Solution: Owl Monitor 2.0**

ICU-grade oversight with **no cameras, no microphones, no wearables, no resident behavior change**.

### **A. The Technical Breakthrough — Why 2.0 Wins Where 1.0 Couldn't**

mmWave radar has been "almost there" for a decade. Two failure modes have blocked every competitor:

| Industry Failure | WiseFido 2.0 Solution |
|---|---|
| **Multipath ghosting** — mirrors, glass, furniture create phantom targets → false fall alarms | **Multi-device fusion** + **per-cell self-learning** identifies and suppresses ghost signatures unique to each room |
| **Angular resolution gap** — single radar misses falls behind furniture or in oblique geometry | **Dual-radar fusion** + Sleepace pressure-pad cross-validation closes the geometric blind spots |

**The result**: ghost suppression and fall recall **both improve with every install**, because the AI learns each home's unique geometry rather than relying on a one-size-fits-all model.

### **B. AI Stack — Precision Terminology**

- **Multi-Modal Sensor Fusion AI** — radar (spatial) + Sleepace (pressure) + activity-window temporal model
- **Self-Learning Spatial Cognition (RoomEngine)** — per-home unsupervised learning of room geometry, ghost zones, and behavior cells
- **Cell-Level Adaptive Thresholds** — every grid cell in every room learns its own false-alarm signature; couch-lying and kitchen-standing stop triggering false falls
- **Fleet-Learning Algorithms** — accuracy compounds across the install base; this is the engine behind the time-to-data moat (§5C)

### **C. StayFSM — Bathroom Black-Box, Solved**

Proprietary finite-state-machine logic optimized for the highest-risk, most-unmonitorable space in any facility. Distinguishes bathing, toileting, and prolonged-no-response — turning the bathroom from a liability black box into auditable, explainable evidence.

### **D. AI-VoIP Active Inquiry**

On detected fall, the system initiates a two-way VoIP call: *"Are you OK? Do you need help?"* Human staff are paged only on negative response or silence — **reducing manual workload by ~90%** while preserving 100% safety coverage.

---

## **4. Market Opportunity**

### **No Incumbent in Radar Elder Care**

Cameras are a red ocean (Wyze, Ring, Nest, ADT). Radar elder care is **empty** — three stacked barriers (hardware, signal processing, elder-scene data) have kept every potential entrant out.

### **TAM Pyramid**

| Stage | Customer | Size | Timing |
|---|---|---|---|
| **SOM** | Denver-triangle PACE / SNF pilot | $10–50M | 2026 |
| **SAM** | US senior living chains + state Medicaid HCBS Waiver | **$3–5B / yr** | Post Series A |
| **TAM** | Aging-in-place residents (B2C / adult-child paid) | **$20–40B** | Post Series B |

**Three customer arcs already mapped**:
1. Large senior living chains + PACE operators (Denver-first: InnovAge)
2. State Medicaid HCBS Waiver programs (Colorado HCPF first)
3. Commercial health insurance / Medicare Advantage / re-insurance / LTC (post-VC)

---

## **5. Competitive Moats — Four Buckets**

### **A. Sensing Moat — *We see what no one else can***
1. Zero-wearable, zero-camera, zero-behavior-change — the **only modality permitted in bedrooms and bathrooms 24×365**
2. **Bathroom black-box solved** via StayFSM
3. **Zero-damage install** — magnetic mounts, 10-second redeploy, no drilling, ADA-friendly

### **B. AI Moat — *Every install makes the model smarter***
4. **Multi-modal fusion AI** with continuous activity-window analysis (not single-event triggers)
5. **Self-learning spatial cognition** — per-home, per-cell unsupervised learning
6. **Five Tier-S long-trend signals** — recovery curve, frailty velocity, stability fingerprint, multi-signal resonance, solo-living vitality (signals hospitals and wearables literally cannot capture)

### **C. Architecture & Data Moat — *Contrarian bet that compounds***
7. **All-cloud architecture** — we skipped MCU edge AI entirely. Result: **5× iteration speed, 80% lower power, $40 BoM**, and the ability to deploy any model size without hardware changes
8. **Raw point-cloud data lake** — every install permanently deepens the only proprietary elder-radar dataset in existence
9. **6–8 month time-to-data moat** — capital cannot compress fleet rollout (3 mo) + per-home baseline convergence (3–5 mo). **A competitor with 10× our funding still arrives 6–8 months behind, and the gap widens monthly.**

### **D. Business & Regulatory Moat — *We own the data exit***
10. **Care-Not-Treatment positioning** — we monitor, we do not diagnose. This deliberately sidesteps FDA 510(k), saving 5–10× R&D cost and 2–3 years to market versus medical-device competitors
11. **B2B-only signed-report monopoly** — residents and families see nothing; institutions cannot export raw data; **WiseFido is the sole HIPAA-compliant data exit** for an entire generation of aging-in-place Americans

### **The Three Contrarian Bets That Define Us**
- **Skipped edge computing** — industry went all-in on MCU/edge AI; we went all-cloud
- **Skipped consumer / skipped medical** — no family app (avoids FTC), no diagnosis (avoids FDA)
- **Skipped cameras** — no fight in the red ocean; own the radar blue ocean instead

---

## **6. Unit Economics**

Targeting 2.7M+ institutional beds in the US, transitioning from Hardware-as-a-Service to high-margin **Hybrid SaaS**.

### **Subscription Tiers**

| Tier | Configuration | Target Market | Monthly |
| :--- | :--- | :--- | :--- |
| **Lite** | 1× SleepPad | Low-acuity / Independent Living | $19.9 / bed |
| **Pro** | 1× Radar + 1× SleepPad | **Assisted Living (mainstream)** | **$69.9 / room** |
| **Pro Max** | 2× Radar + 1× SleepPad | Memory Care / VIP Units | $99.9 / room |

- **Expected ARPU**: ~$68 / room / month
- **Hardware Revenue**: $299–$499 upfront per room → immediate hardware break-even
- **SaaS Gross Margin**: ~85%

---

## **7. Why WiseFido**

- **Founder–Product Fit**: Founded by a software / former network engineer with deep personal experience caring for a stroke-recovering parent
- **Speed of Execution**: 3-month iteration cycle between US design and global supply chain — outpacing ADT, Best Buy Health, and every legacy player
- **Deployment Reality**: 2.0 hardware shipping; lighthouse pilot live at WeCare Denver; free OTA upgrade currently rolling out to four partner sites

---

## **8. Use of Funds ($1M Seed Round)**

- **45% Sales & Field Deployment** — scale Denver team to convert WeCare lighthouse into 10+ paid institutional contracts; this directly extends the time-to-data moat
- **35% Product & Compliance** — HIPAA / HITRUST / SOC 2 / state-level audit completion; FHIR data interoperability for Medicaid / PACE integration
- **20% AI-VoIP Backend & Data Infrastructure** — scale the cloud signal-processing pipeline and data lake that powers fleet learning

> **No FDA 510(k) spend** — by design. Care-Not-Treatment positioning saves 5–10× R&D cost and 2–3 years to market.

---

## **9. The Closing Pitch**

> *Classical signal processing alone cannot solve radar multipath in elder homes. We layered a self-learning AI on top — every cell in every room learns its own signature. Ghost suppression and fall recall both improve with every install.*
>
> *We don't sell health monitoring. We sell **the only data exit** for an entire generation of aging-in-place Americans — and **time-to-data is the one moat capital cannot compress**.*

---

> *"Privacy is Freedom. Safety should be as invisible as air."*

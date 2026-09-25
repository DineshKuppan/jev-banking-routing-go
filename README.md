# TypeSafe Jev: Real-World AI Decision-Making

A Medium blog post + complete Go implementation demonstrating how bounded, typed decision-making changes production AI systems.

## 🎯 Quick Summary

This project shows why "let the AI decide" fails—and what works instead: **deterministic policy + probabilistic judgment + explicit orchestration = safe AI in production.**

- ✅ **2,900-word Medium article** (ready to publish)
- ✅ **Complete Go implementation** (banking transaction router)
- ✅ **Five real-world domains** (coffee, banking, finance, e-commerce, insurance)
- ✅ **Measurement & rollout strategy** (shadow mode, calibration, gradual scaling)

**Status:** Phases 1–3 complete. **Ready for publication (Phase 4).**

---

## 📁 What's Here

### Blog & Publishing
- **`MEDIUM-BLOG-DRAFT.md`** — Final blog post (2,900 words, 10–11 min read)
  - Hook → Problem → Solution → Pattern → Worked Example → Applications → Measurement → Closing
  - Bridges concepts and code seamlessly
  - Medium-optimized (short paragraphs, scannable, code snippets)

- **`MEDIUM-PUBLICATION-GUIDE.md`** — Complete publication guide
  - Formatting, tags, visual suggestions
  - Step-by-step publication process
  - Sharing & promotion templates

- **`PUBLISH-NOW.md`** — Quick-start checklist
  - Three options: Publish today, preview first, or add visuals
  - GitHub setup, Medium posting, sharing templates
  - Estimated time: 45 min to live

### Code Implementation
- **`jev-banking-routing-go/`** — Production-grade Go project
  - Zero external dependencies (stdlib only)
  - Demonstrates safe AI decision workflow
  - Includes tests, API, and runnable example

  **Key components:**
  - `internal/domain/` — Data types (Transaction, Customer, RiskSignals, RoutingDecision)
  - `internal/jev/` — HTTP client for Jev API (3 typed questions)
  - `internal/routing/` — Deterministic policy (hard stops, confidence gates)
  - `internal/service/` — Router with fallback logic
  - `cmd/api/` — HTTP API (GET /healthz, POST /v1/transactions/route)
  - `testdata/` — Worked scenario (₹48.5K transaction, multiple risk signals)

### Planning & Documentation
- **`PROJECT-SUMMARY.md`** — Full project overview
  - All three phases broken down
  - Achievements, next steps, long-term roadmap

- **`PLAN-DRAFT-BLOG.md`** — Original blog concepts (source material)
- **`jev-banking-routing-go.md`** — Original Go guide (source material)
- **`TASKS.md`** — Task breakdown by phase

---

## 🚀 Go Live in 45 Minutes

### Step 1: Push Code to GitHub (5 min)
```bash
cd jev-banking-routing-go
git init
git add .
git commit -m "Initial commit: Jev banking transaction router"
git remote add origin https://github.com/YOUR_USERNAME/jev-banking-routing-go.git
git push -u origin main
```

### Step 2: Update Blog URL (2 min)
Replace `example` with your GitHub username in `MEDIUM-BLOG-DRAFT.md`:
```markdown
git clone https://github.com/YOUR_USERNAME/jev-banking-routing-go.git
```

### Step 3: Publish on Medium (30+ min)
1. Go to https://medium.com (sign up if needed)
2. Click "Write"
3. Copy entire `MEDIUM-BLOG-DRAFT.md` content
4. Paste into editor
5. Add title, tags, publish
6. Share the link

**See `PUBLISH-NOW.md` for detailed steps.**

---

## 💡 Key Insights from the Project

### The Problem
Generative AI in production decisions fails because:
- ❌ Model can invent fields, break enums, omit required data
- ❌ No schema enforcement; you must parse and validate output
- ❌ Confidence scores are uncalibrated
- ❌ Latency spikes when parsing/validating/retrying
- ❌ Hard to audit decisions; explanations are generated text, not evidence

### The Solution: Jev's Approach
✅ **Bounded, typed outputs:** Model returns schema-matched answers (no parsing)
✅ **Three typed questions:** risk_band (choice), recommended_handling (choice), account_takeover_suspected (probability)
✅ **Confidence + probability distribution:** Full calibration info for thresholds
✅ **Explicit policy layer:** Hard stops + confidence gates separate policy from judgment

### The Architecture: Five Layers
1. **Data integrity** — Inputs are complete, valid, authorized
2. **Deterministic policy** — Non-negotiable rules (sanctions, balance, age)
3. **Jev assessment** — Probabilistic judgment (fraud risk, recommended route)
4. **Decision orchestration** — Combine rules + thresholds + policy logic
5. **Governance** — Audit, monitor, measure, improve

**Result:** The model is **one signal** inside a **system humans designed**, not the authority.

---

## 📖 The Blog Post

### Structure (2,900 words)
1. **Hook** — Why AI decisions fail in production
2. **Problem** — Generative LLMs aren't safe for decision workflows
3. **Shift** — What Jev changes about the architecture
4. **Pattern** — Five-layer design framework
5. **Example** — Banking transaction router (complete walkthrough)
6. **Applications** — Five real-world domains (decision contracts shown for each)
7. **Measurement** — Shadow mode, calibration, rollout strategy
8. **Closing** — The standard for production AI

### Tone
- Accessible to broad audience (engineers, product, business)
- Emphasizes patterns over code depth
- Practical examples throughout
- Clear call-to-action (GitHub repo link)

### Medium Optimization
- ✅ Short paragraphs (2–3 sentences)
- ✅ Frequent subheadings (every 200–300 words)
- ✅ Tables for comparison
- ✅ Code snippets (short, focused)
- ✅ Curl example for reproducibility
- ✅ Links to GitHub and TypeSafe docs

---

## 💻 The Go Implementation

### What It Demonstrates
✅ **Safe decision routing:** Hard stops guard against autonomous decisions
✅ **Confidence-gated automation:** Only automate high-confidence cases
✅ **Explicit fallback:** Uncertain cases route to human review
✅ **Auditable decisions:** Full trace of inputs, model output, policy decision

### How to Run
```bash
cd jev-banking-routing-go

# Run tests (verify policy logic)
go test ./...

# Build
go build ./cmd/api

# Start API (requires TYPESAFE_API_KEY)
export TYPESAFE_API_KEY="your-key"
go run ./cmd/api

# Call from another terminal
curl -sS http://localhost:8080/v1/transactions/route \
  -H 'Content-Type: application/json' \
  --data-binary @testdata/scenarios.json | jq
```

### Key Scenario
Transaction: ₹48,500 (23× customer average)
Signals:
- ✓ Clean customer history
- ✓ Sufficient balance
- ✓ Address verification matches
- ✗ Unrecognized device
- ✗ Recent password reset
- ✗ New merchant
- ✗ High velocity (5 txns in 10 min)

**Result:** STEP_UP_AUTHENTICATION
- Model: HIGH risk (0.91 confidence)
- Model: Recommend DECLINE (0.86 confidence)
- Account takeover probability: 0.82
- **Policy decision:** Account takeover signal exceeds threshold → require re-authentication
- **Why not auto-decline?** Amount exceeds auto-decline ceiling (₹10,000); high-value cases require analyst

---

## 📚 Real-World Applications Covered

The pattern works across five domains (all in the blog):

1. **Coffee Shop** — Offer personalization without hurting margin
2. **Banking** — Transaction triage with step-up and manual review
3. **Finance** — Invoice exception routing (PO mismatch, duplicate risk, vendor changes)
4. **E-Commerce** — Returns routing (instant refund vs. evidence vs. manual review)
5. **Insurance** — Claims intake (urgency, routing, completeness, fraud flags)

Each shows the same pattern:
- Deterministic rules enforce constraints
- Jev provides probabilistic judgment
- Thresholds gate automation
- Humans handle low-confidence or high-impact cases

---

## 📊 Measurement Strategy (in Blog)

**Don't measure:** Model accuracy vs. analyst agreement
**Do measure:**
- Fraud capture rate (recall)
- False-positive challenge rate (specificity)
- Manual-review precision
- Business impact (fraud loss, operational cost)
- Latency (p50, p95, p99)
- Fallback rate (when model fails)

**Calibration:** Bucket decisions by confidence and track real outcomes
- 0.50–0.69 confidence: keep in manual review
- 0.70–0.84 confidence: conservative automation
- 0.85–1.00 confidence: candidate for high-priority review

**Rollout:** Shadow mode → low-risk subset → expand → measure

---

## 🎓 Why This Matters

TypeSafe Jev represents a paradigm shift in production AI:

**Before:** Model as oracle
```
State → Prompt → Model output → Parse → Hope → Decide
```

**After:** Model as signal supplier
```
State → Typed questions → Model output (schema-matched) → Policy → Decide
```

This changes everything:
- ✅ No parsing failures (schema is guaranteed)
- ✅ No invented fields (model chooses from your enum)
- ✅ No tail-latency surprises (fixed response shape)
- ✅ Auditable decisions (full trace of policy + model + logic)
- ✅ Testable thresholds (confidence gates, amounts, rules)

The real opportunity: **Fast, safe, probabilistic judgments inside normal software systems.**

---

## 📝 Files at a Glance

| File | Purpose | Status |
|------|---------|--------|
| `MEDIUM-BLOG-DRAFT.md` | Blog post | ✅ Ready |
| `MEDIUM-PUBLICATION-GUIDE.md` | Publication guide | ✅ Ready |
| `PUBLISH-NOW.md` | Quick-start checklist | ✅ Ready |
| `PROJECT-SUMMARY.md` | Full overview | ✅ Complete |
| `jev-banking-routing-go/` | Go implementation | ✅ Tested |
| `go.mod`, `internal/`, `cmd/`, `testdata/`, `Dockerfile`, `Makefile` | Go project files | ✅ In place |
| `.claude/plans/vast-wobbling-turtle.md` | Implementation plan | ✅ Complete |
| `.claude/projects/.../project_jev_medium_blog.md` | Project memory | ✅ Saved |

---

## 🚀 Next Steps

### Immediate (Today)
1. Choose: Publish now, preview first, or add visuals
2. Follow `PUBLISH-NOW.md` (45 min to live)
3. Push code to GitHub
4. Update blog URL
5. Publish on Medium
6. Share on Twitter, Slack, Dev.to, HN

### Short-term (This Week)
1. Monitor Medium engagement
2. Reply to comments and questions
3. Gather feedback from readers

### Long-term (This Month)
1. Follow-up article: "Measuring AI Decisions in Production"
2. Extended walkthrough: "Building a Fraud Router with Jev"
3. Comparative analysis: "AI Decision Systems: A Benchmark"

---

## 📖 Additional Resources

- **TypeSafe Jev Docs:** https://typesafe.ai/blog/introducing-system-one-models-and-jev
- **Go Source Code:** `jev-banking-routing-go/`
- **Implementation Plan:** `.claude/plans/vast-wobbling-turtle.md`
- **Publication Guide:** `MEDIUM-PUBLICATION-GUIDE.md`

---

## ✨ Ready to Ship

All phases complete. Everything is in place. **Pick an option in `PUBLISH-NOW.md` and go live.** 🚀

**Questions?** See the docs, dive into the code, or follow the publication guide step-by-step.

**Ready?** Let's make this live today.

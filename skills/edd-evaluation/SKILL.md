---
name: edd-evaluation
description: >
  Evaluation-Driven Development for agents. Before writing code, define a
  verifiable evaluation contract that specifies, with objective indicators,
  when the goal is achieved. Use this skill when the human asks to implement
  a feature, fix a bug, refactor, or build anything — before writing any
  production code.
---

# EDD — Evaluation-Driven Development

## What EDD is

EDD gives the agent one thing it does not have: an **objective, verifiable
definition of "done"** that exists before implementation and survives the
conversation.

The agent already generates goals, translates intent, writes code, and runs
tests. What it lacks is a permanent, structured contract that says: *this is
what success looks like, these are the indicators that prove it, and this
evidence determines whether we succeeded or not.*

EDD is not a ceremony. It is not a framework. It is a **single YAML file**
per change — the evaluation contract — and a binary that validates it,
runs its evaluators, and computes the decision from evidence.

## When to activate

Activate this skill when the human asks to:
- Implement a feature, fix a bug, or refactor
- Build anything that will be reviewed or merged
- "Just do it" or "quick" — propose a minimal contract, do not skip EDD

Do NOT activate for:
- Informational questions, read-only analysis
- Changes to documentation only
- Chat about the codebase without implementing

## The contract format

Every change gets one contract file at `.edd/contracts/<slug>.yaml`.
The schema is `edd/contract/v2`. Required sections:

### intent — what and why (product language)

\`\`\`yaml
intent:
  actor: "who is affected"
  problem: "the pain, with data if available"
  outcome:
    observable: "what success looks like from outside"
    baseline: "current state — or 'no baseline, measure in Phase 0'"
    threshold: "minimum acceptable value after the change"
    instrument: "how it is measured (tool, query, dashboard)"
    sample: "what is measured (all requests, 7-day window, etc.)"
    window: "observation period (pre-merge, 14 days post-deploy, etc.)"
\`\`\``

**Rule**: every outcome field must be filled. An honest declaration of
uncertainty is valid (`baseline: "unknown — will measure"`). An empty
field is not. `edd validate` enforces this.

### hypothesis — the falsifiable claim

\`\`\`yaml
hypothesis:
  statement: "what this change asserts will happen"
  based_on: "prior evidence — PAT-NNN, LEARNINGS.md date, issue URL, or 'none — first attempt'"
  falsified_when: "concrete, measurable condition that would disprove it"
\`\`\``

### claims — what must be true, and how to verify each one

\`\`\`yaml
claims:
  - id: C1
    type: functional      # functional | regression | product | safety | performance | security
    phase: premerge        # premerge | live
    blocking: true         # when true, this claim must pass for merge/close
    statement: "what this claim asserts"
    evaluators:
      - id: E1
        type: fail_to_pass  # fail_to_pass | pass_to_pass | measurement | structural | human
        run: "pytest tests/test_feature.py::test_case -q"
        expect: "exit 0"
\`\`\``

**Evaluator types**:
- `fail_to_pass`: new behavior that must work (run the test, expect exit 0)
- `pass_to_pass`: existing behavior that must not break (regression guard)
- `measurement`: quantifiable threshold (query a metric, check the value)
- `structural`: code properties (lint, format, secret scan)
- `human`: requires a person's judgment (leave `run` empty, approved out-of-band)

### gates — human approval checkpoints (optional but recommended)

\`\`\`yaml
gates:
  - id: HG-RELEASE
    type: release
    mandatory: true
    reason: "why human approval is required"
\`\`\``

## The workflow

### Step 1: Generate the contract from the human's intent

The human expresses the goal in their own words — a user story, a feature
request, a bug report, a one-line message. You, the agent, translate that
into the contract format above.

**Inverse valuation**: you propose the metrics, thresholds, instruments, and
evaluators. The human reacts to your proposal — accept, correct, or reject.
The human does not need to invent anything. You do the heavy lifting.

### Step 2: Validate the contract

\`\`\`bash
edd validate .edd/contracts/<slug>.yaml
\`\`\`

If `edd` is not in PATH, use the full path to the binary or run the
validation yourself against the schema. The contract MUST pass validation
before implementation begins. Empty fields are blocking errors.

### Step 3: Get human agreement on the contract

Show the human the claims and evaluators you proposed. Let them correct
anything. Once agreed, the contract is the shared definition of "done."

### Step 4: Implement against the contract

Write code that makes the `fail_to_pass` evaluators pass while keeping
the `pass_to_pass` evaluators green. The contract tells you exactly what
"done" looks like.

### Step 5: Verify with evidence

\`\`\`bash
edd verify .edd/contracts/<slug>.yaml [--phase live]
\`\`\`

This runs every evaluator and produces a JSON evidence bundle. Each
evaluator reports `passed`, `failed`, or `skipped`. A blocking claim
is only satisfied when ALL its evaluators are `selected + executed + passed`.

### Step 6: Present the decision

The evidence bundle includes a computed decision:

| Field | Value | Meaning |
|-------|-------|---------|
| merge | allow | All blocking premerge claims pass + all mandatory gates approved |
| merge | deny | A blocking claim failed, or a mandatory gate is pending |
| close | allow | merge=allow + all blocking live claims also pass |
| close | deny | Live claims are pending — the contract stays open |
| status | failed | A blocking premerge claim did not pass |
| status | ready | Claims pass but a mandatory gate awaits human approval |
| status | verified_premerge | Premerge done, live not yet checked |
| status | validated_live | Live claims pass — ready to close |

## Important rules

1. **The contract exists before code.** Never write production code without
   a validated contract. The contract is the agreement; the evidence is the proof.

2. **Empty fields are blocking.** `edd validate` rejects them. If you do
   not know a value, declare that honestly (`"unknown — will measure"`).

3. **You propose, the human corrects.** Do not ask the human to invent
   metrics or thresholds. Propose them yourself, guided by this skill.

4. **The human reviews evidence, not code.** After `edd verify`, the human
   sees which claims passed and which failed. The code is the means, not
   the proof.

5. **No ambiguous PASS.** The decision engine never says "looks good."
   It computes merge/close/status deterministically from structured evidence.

## Minimal contract example

\`\`\`yaml
schema: edd/contract/v2
contract_id: "ADD-HEALTH-ENDPOINT"
profile: express
status: draft

intent:
  actor: "platform engineers"
  problem: "No health check; dead backends received traffic 2x in Q3"
  outcome:
    observable: "GET /health returns 200 with JSON status"
    baseline: "no health endpoint exists"
    threshold: "p95 latency < 100ms under normal load"
    instrument: "prometheus:http_request_duration_seconds"
    sample: "all /health requests during 7-day observation"
    window: "7 days post-deploy"

hypothesis:
  statement: "A health endpoint will let the LB detect and drain unhealthy instances in < 10s"
  based_on: "incident RCA Q3-002"
  falsified_when: "Any incident of traffic to dead backend > 10s after deploy"

claims:
  - id: C1
    type: functional
    phase: premerge
    blocking: true
    statement: "GET /health returns 200 with status ok"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "curl -s http://localhost:8080/health | python3 -c 'import sys,json; d=json.load(sys.stdin); assert d.get("status")=="ok"'"
        expect: "exit 0"

  - id: C2
    type: regression
    phase: premerge
    blocking: true
    statement: "Existing endpoints still return expected responses"
    evaluators:
      - id: E2
        type: pass_to_pass
        run: "make test-integration"
        expect: "exit 0"
\`\`\``

## Using edd-core binary

If `edd` is installed:
- `edd validate <file>` — check contract structure
- `edd verify <file> [--phase live]` — run evaluators, output evidence JSON
- `edd decide <evidence.json>` — compute merge/close/status from a bundle

Install:
\`\`\`bash
curl -sSL https://github.com/edd-framework/edd-core/releases/download/v0.1.0/edd -o /usr/local/bin/edd
chmod +x /usr/local/bin/edd
\`\`\``

If `edd` is not available, validate the contract yourself against the
schema and run evaluators manually. The contract format is the value —
the binary is the reference implementation.

## Proven patterns

These rules emerged from real use across 4 production repos and 740+ eval
specs. They are not theoretical — each one prevented a real failure.

### 1. Every quantitative claim needs a re-executable command

Never cite a file:line as evidence. Always provide a command that anyone
can run to reproduce the result. (Source: LEARNINGS #79 — fabricated evidence
passed two human gates.)

### 2. Measure the baseline before writing the contract

Run a measurement command before filling in intent.outcome.baseline.
If impossible, declare: "no baseline — will measure in Phase 0".
(Source: LEARNINGS #82 — stale baseline caused rework.)

### 3. One test per claim, traceable and reproducible

Each functional claim names a specific test. Never use "make test" as
a catch-all evaluator. (Source: PAT-004 — 1116 tests mapping to claims.)

### 4. Destructive actions require a mandatory human gate

If the change could delete or recreate production resources, add a gate
of type destructive_action with mandatory: true. (Source: AWS/Kiro incident.)

### 5. Start advisory, progress to enforced

After 5 successful cycles, move from advisory to enforced.
(Source: ADR-004 — teams starting enforced often abandon the practice.)

## Full pattern catalog

See [docs/patterns/INDEX.md](docs/patterns/INDEX.md) for the complete list.

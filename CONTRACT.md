# EDD Contract — Canonical Reference

A **contract** is a single YAML file that defines "done correctly" for a
change before any code is written. It is the source of truth against which
an AI agent's work is judged.

## Why a contract?

- **Before code**: you and the agent agree on what "success" looks like
- **During code**: the agent works toward passing evaluators, not guessing
- **After code**: the decision engine computes merge/close from evidence,
  never from prose or vibes

## The file

One YAML file per change. Place it anywhere; the convention is
`.edd/contracts/<slug>.yaml`.

## Schema

```yaml
schema: edd/contract/v2
contract_id: "SLUG"          # kebab-case, unique within the repo
profile: standard            # express | standard | governed
status: draft                # draft | ready | verified_premerge | merged | deployed | validated_live | closed | failed | falsified | superseded | waived | inconclusive

intent:                      # WHAT and WHY (product language)
  actor: ""                  # Who is affected
  problem: ""                # The pain, with data if available
  outcome:
    observable: ""           # What success looks like from outside
    baseline: ""             # Current state (or "no baseline — measure in Phase 0")
    threshold: ""            # Minimum acceptable post-change value
    instrument: ""           # How it is measured
    sample: ""               # What is measured (e.g. "all requests in 7-day window")
    window: ""               # Observation period
  constraints: []            # Non-negotiable boundaries
  approval:                  # Hash-bound: if intent changes, approval is invalidated
    approver: ""
    state: pending           # pending | approved | rejected
    approved_section_sha: "" # Set by the approver after review

hypothesis:                  # Falsifiable claim
  statement: ""              # What this change asserts
  based_on: ""               # Prior evidence (PAT-NNN, LEARNINGS.md date, issue URL, or "none — first attempt")
  falsified_when: ""         # Concrete, measurable condition that would disprove it

claims:                      # HOW we verify (at least one required)
  - id: ""                   # Unique within this contract (e.g. C1, C2)
    type: functional         # functional | regression | product | safety | performance | security
    phase: premerge          # premerge | live
    blocking: true           # When true: this claim must pass for merge/close
    statement: ""            # What this claim asserts
    evaluators:              # At least one per claim
      - id: ""               # Unique within this contract (e.g. E1, E2)
        type: fail_to_pass   # fail_to_pass | pass_to_pass | measurement | structural | human
        run: ""              # Shell command. "human" evaluators: leave empty (approved out-of-band)
        expect: "exit 0"     # exit 0 | output contains "..." | value < N

gates:                       # Human gates (optional but recommended)
  - id: ""                   # e.g. HG-RELEASE
    type: release            # release | architecture | security | destructive_action
    mandatory: true
    reason: ""

technical_approval:          # Hash-bound approval for claims+gates
  approver: ""
  state: pending
  approved_section_sha: ""

decision:                    # Optional: override policy defaults
  merge_requires:
    - premerge_blocking_claims_pass
    - mandatory_gates_approved
  live_close_requires:
    - live_blocking_claims_pass
```

## Evaluator types

| Type | What it checks | run | expect |
|---|---|---|---|
| fail_to_pass | New behavior that must work | Test command | exit 0 |
| pass_to_pass | Existing behavior that must NOT break | Test command | exit 0 |
| measurement | Quantifiable threshold | Query/meter command | output contains "value", value < N |
| structural | Code properties (lint, format, secrets) | Lint/scan command | exit 0 |
| human | Requires a person's judgment | (empty) | Out-of-band approval |

## Decision rules

The engine computes merge/close/status deterministically from structured
evidence. Never from prose or a single ambiguous PASS.

### merge

`allow` when:
- Every claim with blocking:true + phase:premerge has ALL its evaluators selected, executed, and passed
- No gate with mandatory:true is pending (unapproved)

`deny` otherwise.

### close

`allow` when:
- merge is allow
- Every claim with blocking:true + phase:live has ALL its evaluators selected, executed, and passed

`deny` otherwise. A pending live claim keeps the contract open even when every premerge claim is green.

### status

| State | Meaning |
|---|---|
| draft | Contract is being written |
| ready | Premerge claims pass but a mandatory gate is pending |
| verified_premerge | Premerge claims pass, gates approved, live not yet checked |
| merged | Code is in main, waiting for live evidence |
| deployed | Running in production, collecting live signals |
| validated_live | Live claims pass — ready to close |
| closed | close=allow confirmed, contract archived |
| failed | A blocking claim did not pass |
| falsified | The hypothesis was disproven |
| superseded | Replaced by a newer contract |
| waived | Explicitly waived by a human with justification |
| inconclusive | Evidence is insufficient to decide |

## Hash-bound approvals

Intent approval and technical approval are bound to the content they
approved via SHA-256. If the approved section changes, the approval is
automatically invalidated (approved_section_sha no longer matches the
canonical hash of the current content). This prevents silent drift between
"what was approved" and "what was implemented."

## Evidence bundle

Running `edd verify` produces an evidence bundle (JSON). The bundle is the
single source of truth for what passed and failed. The Markdown view next to
it is generated, never the authority.

See [decision-engine.md](decision-engine.md) for the bundle schema and
decision computation details.

## v1 Migration

If you have v1 eval specs in `evals/`, run `edd migrate <path>` for a
best-effort conversion. v1 specs are still recognized by `edd check` and
reported as "legacy — consider migrating."

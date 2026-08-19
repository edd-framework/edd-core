# EDD Core

**EDD Core specifies, once, how success is measured — before any code
is written. The agent generates the goal and its implementation. The
agent evaluates the result against the indicators. EDD Core provides
the one thing the agent does not: an objective, verifiable definition
of "done."**

EDD Core is a single, harness-agnostic focus: the **evaluation spec** —
a structured contract that defines, with objective and verifiable
indicators, when the goal has been achieved. Nothing else.

What EDD Core is **not**:

- It does not generate the goal. The agent does — from a user story,
  a feature request, an epic, a bug report, a one-line message.
- It does not translate intent into the spec. The agent does that,
  guided by the contract format.
- It does not write the code. The agent's harness already does that well.
- It does not evaluate the result. The agent runs the evaluators and
  reports the evidence.

EDD Core provides the format, the validation, and the deterministic
decision. It is the objective mirror the agent holds its work up to.

## What EDD Core does — and does not do

EDD Core's focus is narrow and single: the **evaluation spec**. That is
the artifact that makes "done" objective and verifiable, by defining
indicators for the goal before implementation.

| Concern | Who owns it |
|---------|-------------|
| Generating the goal (from a story, epic, bug, message) | The agent, guided by the human |
| Translating that goal into the evaluation spec | The agent |
| Implementing the code | The agent (its harness already does this well) |
| Evaluating the result against the indicators | The agent, running the declared evaluators |
| **The evaluation spec format** — the contract structure | **EDD Core** |
| **Validating the spec is complete and concrete** | **EDD Core** (`edd validate`) |
| **The deterministic "done" decision from evidence** | **EDD Core** (`edd decide`) |

EDD Core does not compete with the agent or its harness. It fills the one
gap: a permanent, objective definition of success that survives the
conversation and can be checked objectively, by anyone, at any time.

The human does not need to invent metrics, thresholds, or test commands.
The agent proposes them, guided by the spec format. The human reacts to
the proposal — accept, correct, or reject. `edd validate` makes sure
nothing is left vague. The agent then implements and evaluates, and
`edd decide` converts the evidence into an unambiguous decision.

### The cycle

```
Human: "Improve the onboarding flow"
        └─ Intent expressed. Vague. No metrics. No tests.

Agent:  Creates .edd/contracts/improve-onboarding.yaml
        └─ Proposes: completion rate as the metric
        └─ Proposes: 60% threshold within 14 days
        └─ Proposes: making step 3 optional as the intervention
        └─ Proposes: PostHog as the measurement instrument
        └─ Proposes: pytest tests/test_onboarding.py as the evaluator
        └─ Runs edd validate → VALID

        The agent turned "improve onboarding" into precise,
        verifiable instructions. Before writing a single line of code.

Human:  "60% is too aggressive. We use Mixpanel, not PostHog."
        └─ Reacts to the proposal. Corrects two things.

Agent:  Adjusts threshold to 45%. Switches instrument to Mixpanel.
        └─ Runs edd validate → VALID

Human:  "Approved. Build it."
        └─ Contract is now a shared agreement on what "done" means.

Agent:  Implements the skip button. Runs edd verify.
        └─ C1 (skip button works):      passed
        └─ C2 (existing onboarding ok): passed
        └─ C3 (live metric):            pending (needs production data)
        └─ Decision: merge=allow, close=deny

Human:  Reviews the evidence — not the code. Merges. Deploys.

14 days later:

Agent:  Runs edd verify --phase live
        └─ C3 (completion rate >= 45%): passed
        └─ Decision: close=allow

        The contract opened with intent. It closed with evidence.
```

Before the evaluation spec existed, the cycle was: "Write code. Review
code. Hope it works." With it, the cycle is: the agent generates the goal,
translates it into a verifiable spec, the human agrees, the agent
implements and evaluates against the indicators, and the decision is
computed from evidence — not from hope.

## Install

```bash
# Direct download (Linux, amd64)
curl -sSL https://github.com/edd-framework/edd-core/releases/download/v0.1.0/edd -o /usr/local/bin/edd
chmod +x /usr/local/bin/edd

# Go install
go install github.com/edd-framework/edd-core/cmd/edd@v0.1.0

# Build from source
git clone https://github.com/edd-framework/edd-core.git
cd edd-core && go build -o edd ./cmd/edd/
```

## Commands

| Command | Who runs it | What it does |
|---------|------------|--------------|
| `edd validate <contract>` | Agent or human | Structural check. Errors are blocking. |
| `edd verify <contract> [--phase live]` | Agent | Runs evaluators, outputs evidence JSON. |
| `edd decide <evidence.json>` | CI or human | Computes merge/close/status from evidence. |
| `edd help` | Anyone | Shows usage. |

## The contract

A YAML file that the agent writes, guided by this format, to define the
objective and its verifiable indicators. See [CONTRACT.md](CONTRACT.md)
for the full reference.

```yaml
schema: edd/contract/v2
contract_id: "ADD-OAUTH2-LOGIN"
profile: standard

intent:
  actor: "billing API consumers"
  problem: "No auth on billing endpoints; PII exposed"
  outcome:
    observable: "POST /auth/token returns JWT; billing endpoints require it"
    baseline: "0 endpoints secured"
    threshold: "auth covers all /api/billing/* routes"
    instrument: "curl + integration test suite"
    sample: "all billing endpoints"
    window: "pre-merge"

hypothesis:
  statement: "Adding OAuth2 will secure billing endpoints without breaking existing integrations"
  based_on: "security audit Q3-001"
  falsified_when: "Any billing endpoint accepts requests without a valid token"

claims:
  - id: C1
    type: functional
    phase: premerge
    blocking: true
    statement: "POST /auth/token returns a signed JWT for valid credentials"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "pytest tests/test_auth.py::test_token_issued -q"
        expect: "exit 0"

  - id: C2
    type: regression
    phase: premerge
    blocking: true
    statement: "Existing billing integration tests still pass"
    evaluators:
      - id: E2
        type: pass_to_pass
        run: "pytest tests/test_billing.py -q"
        expect: "exit 0"

  - id: C3
    type: safety
    phase: premerge
    blocking: true
    statement: "No secrets or credentials in source or logs"
    evaluators:
      - id: E3
        type: structural
        run: "grep -rE 'TODO-remove|temp-token|sk-.*test' src/ && exit 1 || exit 0"
        expect: "exit 0"

gates:
  - id: HG-RELEASE
    type: release
    mandatory: true
    reason: "Auth changes require security review"
```

## Decision engine

Deterministic. Never a single ambiguous PASS.

| Decision | Rule |
|----------|------|
| merge=allow | Every blocking premerge claim: all its evaluators selected, executed, and passed. No mandatory gate pending. |
| merge=deny | Any blocking claim fails, or a mandatory gate is unapproved. |
| close=allow | merge=allow + every blocking live claim also satisfied. |
| close=deny | Otherwise. A pending live claim blocks close even if premerge is green. |

| Status | Meaning |
|--------|---------|
| failed | A blocking premerge claim did not pass |
| ready | Claims pass, a mandatory gate awaits approval |
| verified_premerge | Premerge done, live not yet checked |
| validated_live | Live claims pass |
| closed | Close confirmed |

Evaluator fine print:
- `skipped`, `xfail`, `executed=false`, or absent from evidence → claim NOT satisfied
- A blocking claim needs _all_ its evaluators selected+executed+passed

## Harness adapters

EDD Core is a CLI. **Any agent, harness, or coding IDE can use it.**
The contract format and decision engine are identical everywhere.
Only the invocation wrapper differs — and it is always a thin adapter
with zero business logic.

| Agent / Harness | How it calls EDD Core |
|-----------------|----------------------|
| **DeepSeek Harness (DSH)** | `tools.bash("edd verify contract.yaml")` |
| **pi** | 40-line TypeScript wrapper calling `edd verify` |
| **Claude Code** | Terminal: `edd verify contract.yaml` |
| **GitHub Copilot** | Terminal or extension invoking `edd` CLI |
| **Cursor** | Terminal: `edd verify contract.yaml` |
| **Windsurf** | Terminal: `edd verify contract.yaml` |
| **Gemini CLI** | Terminal: `edd verify contract.yaml` |
| **VSCode** | Extension exec: `child_process.exec("edd", ["verify", path])` |
| **Any CI pipeline** | `edd verify contract.yaml --phase premerge` |

The contract defines what to verify. The binary runs it. The evidence
is JSON that any tool can consume. The harness brings its own strengths —
conversation, tool integration, code generation — and EDD Core adds the
evaluation loop: define success before coding, verify with evidence,
decide deterministically.

## v1 Migration

Have v1 eval specs in `evals/`? See [docs/migration-v1-to-v2.md](docs/migration-v1-to-v2.md).
Migrate at your own pace. No flag day.

## License

Business Source License 1.1 → Apache 2.0 on 2030-03-02.

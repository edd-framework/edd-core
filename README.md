# EDD Core

**You describe the goal. The agent writes the evaluation contract. Then
implements against it. You verify the evidence. Done.**

EDD flips AI-assisted development: the agent doesn't start coding. It
starts by defining _how you will know the code is correct_ — a structured
contract with falsifiable claims and executable evaluators. Then it writes
code that satisfies them. You review the evidence, not the code.

It works from any input: a user story, a feature request, an epic, a bug
report, a one-line Slack message. The agent turns intent into a verifiable
contract. The contract becomes the source of truth for "done."

EDD Core is the minimal engine: one binary, one contract format, one
decision engine. No model selection, no ceremony, no harness lock-in.
Any agent, any IDE, any stack.

## The loop

```
Human: "Add OAuth2 login for the billing API"

Agent:  Creates .edd/contracts/add-oauth2-login.yaml
        └─ intent: what problem, for whom, what outcome
        └─ claims: what must be true
        └─ evaluators: shell commands that check each claim

Agent:  Runs edd verify → evidence bundle (JSON)
        └─ C1 (token issued): passed
        └─ C2 (existing tests): passed
        └─ C3 (no secrets leaked): passed
        └─ Decision: merge=allow, close=deny (live pending)

Human:  Reviews the evidence. Approves. Merges.
```

The agent defined how success is measured _before_ writing code. The
contract is the agreement. The evidence is the proof.

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

A YAML file the agent creates from your intent. See
[CONTRACT.md](CONTRACT.md) for the full reference.

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

EDD Core is a CLI. Any agent or IDE can use it:

- **pi**: 40-line TypeScript wrapper calling `edd verify`
- **DSH**: `tools.bash("edd verify contract.yaml")`
- **Claude Code / Cursor / Windsurf / Copilot**: terminal command
- **VSCode**: exec from extension

No harness contains EDD business logic. The CLI is the single source of truth.

## v1 Migration

Have v1 eval specs in `evals/`? See [docs/migration-v1-to-v2.md](docs/migration-v1-to-v2.md).
Migrate at your own pace. No flag day.

## License

Business Source License 1.1 → Apache 2.0 on 2030-03-02.

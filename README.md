# EDD Core

**Define how you will verify code works — before an AI agent writes it.**

EDD Core is the minimal, harness-agnostic engine of Evaluation-Driven
Development. One binary, one contract format, one decision engine.

## Install

### Homebrew (macOS/Linux)

```bash
brew install edd-framework/tap/edd
```

### Go install

```bash
go install github.com/edd-framework/edd-core/cmd/edd@latest
```

### Direct download

```bash
curl -sSL https://edd-framework.github.io/edd-core/install.sh | bash
```

## Quickstart (3 minutes)

```bash
# Initialize EDD in your repo
edd init

# Create a contract for your change
edd new "Add health check endpoint" --profile standard

# Edit .edd/contracts/add-health-check-endpoint.yaml
# Fill in intent, claims, and evaluators

# Run evaluators and produce evidence
edd verify .edd/contracts/add-health-check-endpoint.yaml

# See claims, gates, and the computed decision
edd status .edd/contracts/add-health-check-endpoint.yaml
```

## How it works

1. **Contract** — a YAML file that defines "done correctly": claims
   (falsifiable statements) and evaluators (shell commands that check them)
2. **Verify** — run the evaluators, produce a JSON evidence bundle
3. **Decide** — the engine computes merge/close/status deterministically
   from evidence, never from prose or vibes

## Contract format

See [CONTRACT.md](CONTRACT.md) for the complete reference.

A minimal contract:

```yaml
schema: edd/contract/v2
contract_id: "MY-CHANGE"
profile: standard

intent:
  actor: "developers deploying the service"
  problem: "No health check; 2 incidents of dead backends receiving traffic"
  outcome:
    observable: "GET /health returns 200"
    baseline: "no endpoint exists"
    threshold: "p95 < 100ms"
    instrument: "prometheus:http_request_duration_seconds"
    sample: "all /health requests"
    window: "7 days"

hypothesis:
  statement: "Health endpoint will let LB detect dead instances in < 10s"
  falsified_when: "Any incident of traffic to dead backend > 10s"

claims:
  - id: C1
    type: functional
    phase: premerge
    blocking: true
    statement: "GET /health returns 200 with status ok"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "curl -s http://localhost:8080/health | grep -q ok"
```

## Harness adapters

EDD Core is a CLI tool. Any agent or IDE can use it:

- **pi**: 40-line TypeScript wrapper calling `edd verify`
- **DSH**: `tools.bash("edd verify contract.yaml")`
- **Claude Code / Cursor / Windsurf / Copilot**: terminal command
- **VSCode**: exec from extension

No harness contains EDD business logic. The CLI is the single source of truth.

## Decision engine

The engine is deterministic and never collapses to a single ambiguous PASS:

| Merge | When |
|-------|------|
| allow | All blocking premerge claims pass + all mandatory gates approved |
| deny  | Any blocking premerge claim fails or a mandatory gate is pending |

| Close | When |
|-------|------|
| allow | Merge is allow + all blocking live claims pass |
| deny  | Merge is deny or a blocking live claim is pending |

| Status | Meaning |
|--------|---------|
| failed | A blocking premerge claim did not pass |
| ready | Premerge claims pass but a gate is pending approval |
| verified_premerge | Premerge done, live not yet checked |
| validated_live | Live claims pass — ready to close |
| closed | Close confirmed |

A pending live claim keeps close=deny even when every premerge claim is
green. There is no single ambiguous PASS.

## Differences from v1

EDD v1 used two files per change (objective spec + eval spec) with ~40
YAML fields, many specific to the toolkit's internal framework. EDD Core
uses one contract file with ~15 fields, all focused on the evaluation
itself. Everything else (autonomy level, budget, model selection, ceremony)
is a harness/operator concern, not a contract concern.

v1 specs in `evals/` are still recognized by `edd check` and reported
as "legacy — consider migrating." Use `edd migrate` for best-effort
conversion.

## License

Business Source License 1.1 — permits internal use, consulting,
education, and research. Converts to Apache 2.0 on 2030-03-02.

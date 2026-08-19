# EDD Core

**Define how you will verify code works — before an AI agent writes it.**

EDD Core is the minimal, harness-agnostic engine of Evaluation-Driven
Development. One binary, one contract format, one decision engine.
No model selection, no ceremony levels, no autonomy budgets.
The contract defines what "done" means. The CLI verifies it.
Any agent or IDE can call the CLI.

## Install

### Direct download (Linux, amd64)

```bash
curl -sSL https://github.com/edd-framework/edd-core/releases/download/v0.1.0/edd -o /usr/local/bin/edd
chmod +x /usr/local/bin/edd
edd help
```

### Go install

```bash
go install github.com/edd-framework/edd-core/cmd/edd@v0.1.0
```

### Build from source

```bash
git clone https://github.com/edd-framework/edd-core.git
cd edd-core
go build -o edd ./cmd/edd/
./edd help
```

## Quickstart

```bash
# 1. Create a contract file (anywhere in your repo)
cat > .edd/contracts/my-change.yaml << 'EOF'
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
        run: "curl -s http://localhost:8080/health"
EOF

# 2. Validate the contract structure
edd validate .edd/contracts/my-change.yaml

# 3. Run evaluators, produce evidence JSON
edd verify .edd/contracts/my-change.yaml

# 4. Inspect the evidence
edd decide .edd/contracts/my-change.yaml
```

## How it works

1. **Contract** — a YAML file with claims (falsifiable statements) and
   evaluators (shell commands that check them)
2. **Verify** — run the evaluators, produce a JSON evidence bundle
3. **Decide** — the engine computes merge/close/status deterministically
   from evidence, never from prose or vibes

## Commands

| Command | What it does |
|---------|-------------|
| `edd validate <file>` | Check contract structure. Errors are blocking. |
| `edd verify <file> [--phase live]` | Run evaluators, output evidence JSON. |
| `edd decide <evidence.json>` | Compute merge/close/status from a bundle. |
| `edd help` | Show usage. |

## Decision engine

The engine is deterministic and never collapses to a single ambiguous PASS:

| Decision | When |
|----------|------|
| merge=allow | All blocking premerge claims pass + all mandatory gates approved |
| merge=deny | Any blocking claim fails or a mandatory gate is pending |
| close=allow | Merge is allow + all blocking live claims pass |
| close=deny | Merge is deny or a blocking live claim is pending |

| Status | Meaning |
|--------|---------|
| failed | A blocking premerge claim did not pass |
| ready | Claims pass but a mandatory gate is pending approval |
| verified_premerge | Premerge done, no live claims (or live pending) |
| validated_live | Live claims pass, ready to close |
| closed | Close confirmed |

A pending live claim keeps `close=deny` even when every premerge claim is
green. There is no single ambiguous PASS.

Evaluator semantics:
- `skipped`, `xfail`, `executed=false` → claim NOT satisfied
- Evaluator absent from evidence → claim NOT satisfied
- All evaluators must be `selected + executed + result=passed` for a blocking claim

## Contract format

See [CONTRACT.md](CONTRACT.md) for the complete reference. One file per change.
Required fields: `schema`, `contract_id`, `profile`, `intent`, `hypothesis`, `claims`.
Each claim requires `id`, `type`, `phase`, `statement`, and at least one `evaluator`.

## Harness adapters

EDD Core is a CLI. Any agent or IDE can use it:

- **pi**: 40-line TypeScript wrapper calling `edd verify`
- **DSH**: `tools.bash("edd verify contract.yaml")`
- **Claude Code / Cursor / Windsurf / Copilot**: terminal command
- **VSCode**: exec from extension

No harness contains EDD business logic. The CLI is the single source of truth.

## v1 Migration

If you have v1 eval specs in `evals/`, see [docs/migration-v1-to-v2.md](docs/migration-v1-to-v2.md).
v1 specs are recognized as legacy. Migrate at your own pace — no flag day.

## License

Business Source License 1.1 — permits internal use, consulting, education,
and research. Converts to Apache 2.0 on 2030-03-02.

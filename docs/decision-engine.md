# Decision Engine

The decision engine computes `merge`, `close`, and `status` deterministically
from structured evidence. It never collapses to a single ambiguous PASS and
never depends on key order in the source YAML.

## Rules

### merge

`allow` when:
- Every claim with `blocking: true` and `phase: premerge` has ALL its
  evaluators `selected`, `executed`, and `result: passed`
- No gate with `mandatory: true` is pending (unapproved)

`deny` otherwise.

### close

`allow` when:
- `merge` is `allow`
- Every claim with `blocking: true` and `phase: live` has ALL its
  evaluators `selected`, `executed`, and `result: passed`

`deny` otherwise. A pending live claim keeps `close: deny` even when
every premerge claim is green.

### status

| State | Condition |
|-------|-----------|
| `failed` | A blocking premerge claim did not pass |
| `ready` | Premerge claims pass but a mandatory gate is pending |
| `verified_premerge` | Premerge done, gates approved, no live claims (or live pending) |
| `validated_live` | Live claims also pass — ready to close |
| `closed` | `close: allow` confirmed |

## Evaluator semantics

A blocking claim is satisfied ONLY when ALL its declared evaluators are:
- `selected: true`
- `executed: true`
- `result: "passed"`

These NEVER satisfy a blocking claim:
- `result: "skipped"`
- `result: "xfail"`
- `selected: true, executed: false` (declared but not run)
- Declared in the contract but absent from the evidence bundle

## Gate semantics

- A `mandatory: true` gate with `approved: false` blocks `merge` but
  does NOT set `status: failed`. Status stays `ready` to distinguish
  "automated evidence passed, waiting for human approval" from
  "automated evidence failed."
- Non-mandatory gates never block merge.
- Gate dedup in `combine_close`: if a gate is pending in the premerge
  bundle but approved in the live bundle, the combined view reads it
  as approved (approved-wins).

## Stale bundles

An evidence bundle with `contract_sha` that does not match the current
canonical hash of the contract is **stale**. Stale bundles never count
toward a `close` decision.

## Permission to be pending

A contract with no live claims reaches `verified_premerge` and stays
there. It never reaches `validated_live` or `closed` — and that is
correct. Not every change needs a live observation window. The engine
does not force closure where the contract does not ask for it.

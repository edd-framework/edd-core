# Migration — EDD v1 to EDD Core v2

EDD v1 used two files per change (objective spec + eval spec) with ~40 YAML
fields, many specific to the toolkit's internal framework. EDD Core uses one
contract file with ~15 fields, all focused on the evaluation itself.

## Quick migration

```bash
$ edd migrate evals/eval-spec-my-feature.yaml
```

This reads a v1 eval spec and generates `.edd/contracts/my-feature.yaml`.
Best-effort: fields that have no v2 equivalent are reported as warnings.

## What maps 1:1

| v1 field | v2 field |
|----------|----------|
| `objective.titulo` | `contract_id` (slugified) |
| `objective.descripcion` | `intent.problem` |
| `objective.alcance.incluye` | `intent.constraints` |
| `evaluation.automated_tests.unit[]` | `claims[].evaluators[]` with `phase: premerge` |
| `evaluation.metrics[]` | `claims[].evaluators[]` with `type: measurement` |
| `evaluation.hypothesis.statement` | `hypothesis.statement` |
| `evaluation.hypothesis.based_on` | `hypothesis.based_on` |
| `evaluation.hypothesis.falsifiable_condition` | `hypothesis.falsified_when` |
| `evaluation.human_gates[]` | `gates[]` |

## What is dropped

These v1 fields were internal framework concerns and have no v2 equivalent:

| v1 field | Why dropped |
|----------|-------------|
| `autonomy_level` | Harness/operator decision, not contract concern |
| `orchestration_pattern` | Same |
| `budget.*`, `model_selection`, `preferred_model` | Same |
| `stop_conditions` | Same |
| `static_analysis` (as a block) | Use a claim with `type: structural` |
| `model_graders` (as a block) | Use a claim with `type: human` |
| `security` (as a block) | Use claims with `type: safety` |
| `eval_type`, `graduation_criteria`, `pass_metric` | Redundant with decision engine |
| `scope` | Claims declare their scope individually |
| `objective.riesgos_identificados` | Document in `intent.constraints` or a risk-register claim |
| `objective.adrs_referenciados` | Document in `hypothesis.based_on` |

## Coexistence

v1 specs in `evals/` continue to work. `edd check` recognizes them and
reports them as "legacy — consider migrating." CI gates that check for
eval spec presence accept both v1 and v2 formats.

## No flag day

Migrate at your own pace. New changes use v2 contracts. Existing v1 specs
keep working. When a v1 spec is touched for a real change, migrate it then.

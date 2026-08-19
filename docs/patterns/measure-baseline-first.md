# Measure the baseline before writing the contract

Do not hardcode baselines from memory. Run a measurement command
before drafting the contract.

## Origin

LEARNINGS #82 (2026-06-16): A contract stated "32/32 items" from
the previous cycle. The real number was 36. The agent wrote acceptance
criteria against the stale number.

## Rule

Before filling in intent.outcome.baseline, measure the current state.
If measurement is impossible, declare honestly:

```yaml
baseline: "no baseline — Phase 0 will measure before implementation"
```

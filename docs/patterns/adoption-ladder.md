# Adoption ladder: advisory to enforced

Start with advisory mode. Move to enforced after 5 successful cycles.

## Origin

ADR-004 — enforcement configurable. Forcing strict enforcement on
day one creates friction. But no enforcement makes the contract optional.

## Rule

| Stage | Mode | Behavior |
|-------|------|----------|
| Onboarding | advisory | Reports gaps, never blocks |
| Active | enforced | Blocks merge without validated contract |
| Mature | enforced + adversarial | Independent review of claims |

After 5 successful PRs with validated contracts, progress to the
next stage.

# Destructive action gate

Every contract that could delete, recreate, or overwrite production
resources must include a mandatory human gate.

## Origin

PAT-010 — validated against the AWS/Kiro incident (13h outage).
An agent deleted and recreated a customer-facing system. A mandatory
gate would have prevented it.

## Rule

```yaml
gates:
  - id: HG-DESTRUCTIVE
    type: destructive_action
    mandatory: true
    reason: "This change drops and recreates the users table"
```

The agent must stop before executing and present exactly what will
be destroyed, with a dry-run showing affected rows.

# Tests as executable acceptance criteria

Every functional claim must name the test that verifies it. The test
is the executable form of the acceptance criterion.

## Origin

PAT-004 — proven across 3+ repos, 1116 tests in eod-gestor-emails.
When tests map to claims, evidence is automatic.

## Rule

```yaml
claims:
  - id: C1
    statement: "Login returns JWT on valid credentials"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "pytest tests/test_auth.py::test_login_returns_jwt -q"
        expect: "exit 0"
```

Not:

```yaml
claims:
  - id: C1
    statement: "Login works"
    evaluators:
      - id: E1
        run: "make test"
```

One test per claim. Traceable. Reproducible.

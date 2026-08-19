# Premise verification

Every quantitative claim must be backed by a re-executable command.
Never cite a file:line — cite the command anyone can run to reproduce
the result.

## Origin

LEARNINGS #79 (2026-06-10): A discovery subagent fabricated evidence
by citing a non-existent file:line. The fabrication passed two human
gates. The falsifiable_condition eventually caught it.

## Rule

When a claim asserts a quantitative fact, the evaluator must include
a run command that produces the raw output.

```yaml
claims:
  - id: C1
    statement: "45 emails skipped due to missing config"
    evaluators:
      - id: E1
        type: measurement
        run: "grep -c EML_DATALAKE_PATH /var/log/errors.log"
        expect: "output contains 45"
```

## Also see

LEARNINGS #101: Numbers can be correct but interpretation wrong.
Always verify the mechanism, not just the count.

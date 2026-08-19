package schema_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/edd-framework/edd-core/internal/schema"
)

func mustParse(t *testing.T, yamlStr string) map[string]interface{} {
	t.Helper()
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		t.Fatalf("yaml parse error: %v", err)
	}
	return doc
}

func TestValidContractPasses(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
status: draft
intent:
  actor: "test"
  problem: "test problem"
  outcome:
    observable: "test observable"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7 days"
hypothesis:
  statement: "test hypothesis"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    blocking: true
    statement: "test claim"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "true"
gates:
  - id: HG-RELEASE
    type: release
    mandatory: true
`)

	errs := schema.ValidateContract(doc)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("unexpected error: %v", e)
		}
	}
}

func TestMissingSchema(t *testing.T) {
	doc := mustParse(t, `
schema: wrong/schema
contract_id: "TEST-001"
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    statement: "test"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "true"
`)

	errs := schema.ValidateContract(doc)
	if len(errs) == 0 {
		t.Error("expected errors for wrong schema")
	}
	found := false
	for _, e := range errs {
		if e.Field == "schema" {
			found = true
		}
	}
	if !found {
		t.Error("expected schema field error")
	}
}

func TestMissingContractID(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    statement: "test"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "true"
`)

	errs := schema.ValidateContract(doc)
	if len(errs) == 0 {
		t.Error("expected errors for missing contract_id")
	}
}

func TestMissingIntent(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    statement: "test"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "true"
`)

	errs := schema.ValidateContract(doc)
	if len(errs) == 0 {
		t.Error("expected errors for missing intent")
	}
}

func TestMissingClaims(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
`)

	errs := schema.ValidateContract(doc)
	if len(errs) == 0 {
		t.Error("expected errors for missing claims")
	}
}

func TestDuplicateClaimIDs(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    statement: "claim 1"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "true"
  - id: C1
    type: regression
    phase: premerge
    statement: "claim 2"
    evaluators:
      - id: E2
        type: pass_to_pass
        run: "true"
`)

	errs := schema.ValidateContract(doc)
	found := false
	for _, e := range errs {
		if e.Field == "claims[1].id" {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate claim ID error")
	}
}

func TestInvalidClaimType(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: invalid_type
    phase: premerge
    statement: "test"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "true"
`)

	errs := schema.ValidateContract(doc)
	found := false
	for _, e := range errs {
		if e.Field == "claims[0].type" {
			found = true
		}
	}
	if !found {
		t.Error("expected invalid claim type error")
	}
}

func TestContractWithoutEvaluators(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    statement: "test"
    evaluators: []
`)

	errs := schema.ValidateContract(doc)
	found := false
	for _, e := range errs {
		if e.Field == "claims[0].evaluators" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing evaluators error, got: %v", errs)
	}
}

func TestInvalidEvaluatorType(t *testing.T) {
	doc := mustParse(t, `
schema: edd/contract/v2
contract_id: "TEST-001"
profile: standard
intent:
  actor: "test"
  problem: "test"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
hypothesis:
  statement: "test"
  falsified_when: "test fails"
claims:
  - id: C1
    type: functional
    phase: premerge
    statement: "test"
    evaluators:
      - id: E1
        type: unknown_type
        run: "true"
`)

	errs := schema.ValidateContract(doc)
	found := false
	for _, e := range errs {
		if e.Field == "claims[0].evaluators[0].type" {
			found = true
		}
	}
	if !found {
		t.Error("expected invalid evaluator type error")
	}
}

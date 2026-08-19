package decision_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/edd-framework/edd-core/internal/decision"
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

func TestClaimSatisfiedAllPassed(t *testing.T) {
	claim := map[string]interface{}{
		"id": "C1",
		"evaluators": []interface{}{
			map[string]interface{}{"id": "E1"},
		},
	}
	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "passed"},
	}

	if !decision.ClaimSatisfied(claim, evidence) {
		t.Error("expected satisfied with passed evaluator")
	}
}

func TestClaimSatisfiedSkipped(t *testing.T) {
	claim := map[string]interface{}{
		"id": "C1",
		"evaluators": []interface{}{
			map[string]interface{}{"id": "E1"},
		},
	}
	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "skipped"},
	}

	if decision.ClaimSatisfied(claim, evidence) {
		t.Error("expected NOT satisfied with skipped evaluator")
	}
}

func TestClaimSatisfiedNotExecuted(t *testing.T) {
	claim := map[string]interface{}{
		"id": "C1",
		"evaluators": []interface{}{
			map[string]interface{}{"id": "E1"},
		},
	}
	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: false, Result: ""},
	}

	if decision.ClaimSatisfied(claim, evidence) {
		t.Error("expected NOT satisfied with unexecuted evaluator")
	}
}

func TestClaimSatisfiedAbsentFromEvidence(t *testing.T) {
	claim := map[string]interface{}{
		"id": "C1",
		"evaluators": []interface{}{
			map[string]interface{}{"id": "E1"},
		},
	}

	if decision.ClaimSatisfied(claim, nil) {
		t.Error("expected NOT satisfied with absent evaluator")
	}
}

func TestClaimSatisfiedXFail(t *testing.T) {
	claim := map[string]interface{}{
		"id": "C1",
		"evaluators": []interface{}{
			map[string]interface{}{"id": "E1"},
		},
	}
	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "xfail"},
	}

	if decision.ClaimSatisfied(claim, evidence) {
		t.Error("expected NOT satisfied with xfail evaluator")
	}
}

func TestClaimSatisfiedMultipleAllMustPass(t *testing.T) {
	claim := map[string]interface{}{
		"id": "C1",
		"evaluators": []interface{}{
			map[string]interface{}{"id": "E1"},
			map[string]interface{}{"id": "E2"},
		},
	}
	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "passed"},
		{ID: "E2", ClaimID: "C1", Selected: true, Executed: true, Result: "failed"},
	}

	if decision.ClaimSatisfied(claim, evidence) {
		t.Error("expected NOT satisfied when one evaluator fails")
	}
}

func TestGatePendingMandatoryNotApproved(t *testing.T) {
	gate := map[string]interface{}{
		"id": "HG-RELEASE", "mandatory": true,
	}
	evidence := []schema.EvidenceGate{
		{ID: "HG-RELEASE", Mandatory: true, Approved: false},
	}

	if !decision.GatePending(gate, evidence) {
		t.Error("expected pending for unapproved mandatory gate")
	}
}

func TestGatePendingMandatoryApproved(t *testing.T) {
	gate := map[string]interface{}{
		"id": "HG-RELEASE", "mandatory": true,
	}
	evidence := []schema.EvidenceGate{
		{ID: "HG-RELEASE", Mandatory: true, Approved: true},
	}

	if decision.GatePending(gate, evidence) {
		t.Error("expected NOT pending for approved mandatory gate")
	}
}

func TestGatePendingNotMandatory(t *testing.T) {
	gate := map[string]interface{}{
		"id": "HG-OPTIONAL", "mandatory": false,
	}

	if decision.GatePending(gate, nil) {
		t.Error("expected NOT pending for non-mandatory gate")
	}
}

func TestComputePremergeCompleteLivePending(t *testing.T) {
	contract := mustParse(t, `
claims:
  - id: C1
    blocking: true
    phase: premerge
    evaluators:
      - id: E1
  - id: C2
    blocking: true
    phase: live
    evaluators:
      - id: E2
gates: []
`)

	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "passed"},
	}

	dec := decision.Compute(contract, evidence, nil)
	if dec.Merge != "allow" {
		t.Errorf("expected merge=allow, got %s", dec.Merge)
	}
	if dec.Close != "deny" {
		t.Errorf("expected close=deny, got %s", dec.Close)
	}
	if dec.Status != "verified_premerge" {
		t.Errorf("expected status=verified_premerge, got %s", dec.Status)
	}
}

func TestComputeLivePassedAllowsClose(t *testing.T) {
	contract := mustParse(t, `
claims:
  - id: C1
    blocking: true
    phase: premerge
    evaluators:
      - id: E1
  - id: C2
    blocking: true
    phase: live
    evaluators:
      - id: E2
gates: []
`)

	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "passed"},
		{ID: "E2", ClaimID: "C2", Selected: true, Executed: true, Result: "passed"},
	}

	dec := decision.Compute(contract, evidence, nil)
	if dec.Merge != "allow" {
		t.Errorf("expected merge=allow, got %s", dec.Merge)
	}
	if dec.Close != "allow" {
		t.Errorf("expected close=allow, got %s", dec.Close)
	}
	if dec.Status != "validated_live" {
		t.Errorf("expected status=validated_live, got %s", dec.Status)
	}
}

func TestComputePremergeFailedDeniesMerge(t *testing.T) {
	contract := mustParse(t, `
claims:
  - id: C1
    blocking: true
    phase: premerge
    evaluators:
      - id: E1
  - id: C2
    blocking: true
    phase: live
    evaluators:
      - id: E2
gates: []
`)

	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "failed"},
		{ID: "E2", ClaimID: "C2", Selected: true, Executed: true, Result: "passed"},
	}

	dec := decision.Compute(contract, evidence, nil)
	if dec.Merge != "deny" {
		t.Errorf("expected merge=deny, got %s", dec.Merge)
	}
	if dec.Close != "deny" {
		t.Errorf("expected close=deny, got %s", dec.Close)
	}
}

func TestComputeGatePendingDeniesMerge(t *testing.T) {
	contract := mustParse(t, `
claims:
  - id: C1
    blocking: true
    phase: premerge
    evaluators:
      - id: E1
gates:
  - id: HG-RELEASE
    type: release
    mandatory: true
`)

	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "passed"},
	}
	evidenceGates := []schema.EvidenceGate{
		{ID: "HG-RELEASE", Mandatory: true, Approved: false},
	}

	dec := decision.Compute(contract, evidence, evidenceGates)
	if dec.Merge != "deny" {
		t.Errorf("expected merge=deny with pending gate, got %s", dec.Merge)
	}
	if dec.Status != "ready" {
		t.Errorf("expected status=ready with pending gate, got %s", dec.Status)
	}
}

func TestComputeNoLiveClaimsReachesVerified(t *testing.T) {
	contract := mustParse(t, `
claims:
  - id: C1
    blocking: true
    phase: premerge
    evaluators:
      - id: E1
gates: []
`)

	evidence := []schema.EvidenceEvaluator{
		{ID: "E1", ClaimID: "C1", Selected: true, Executed: true, Result: "passed"},
	}

	dec := decision.Compute(contract, evidence, nil)
	if dec.Merge != "allow" {
		t.Errorf("expected merge=allow, got %s", dec.Merge)
	}
	if dec.Status != "verified_premerge" {
		t.Errorf("expected status=verified_premerge, got %s", dec.Status)
	}
}

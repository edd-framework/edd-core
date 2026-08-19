package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/edd-framework/edd-core/internal/runner"
)

func mustParseContract(t *testing.T, yamlStr string) map[string]interface{} {
	t.Helper()
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		t.Fatalf("yaml parse error: %v", err)
	}
	return doc
}

func TestBuggyImplementationIsCaughtByEvaluator(t *testing.T) {
	dir := t.TempDir()

	checkoutBuggy := "def checkout_total_with_shipping(items):\n    subtotal = sum(item[\"price\"] * item[\"quantity\"] for item in items)\n    if subtotal > 50:\n        shipping = 0\n    else:\n        shipping = 5.99\n    return subtotal + shipping\n"
	if err := os.WriteFile(filepath.Join(dir, "checkout.py"), []byte(checkoutBuggy), 0o644); err != nil {
		t.Fatalf("write checkout.py: %v", err)
	}

	boundaryTest := "import sys\nsys.path.insert(0, \".\")\nfrom checkout import checkout_total_with_shipping\n\nitems = [{\"price\": 50.0, \"quantity\": 1}]\nresult = checkout_total_with_shipping(items)\nassert result == 50.0, f\"FAIL: exactly $50 charged shipping. Got {result}, expected 50.0\"\nprint(\"BOUNDARY PASSED\")\n"
	if err := os.WriteFile(filepath.Join(dir, "test_boundary.py"), []byte(boundaryTest), 0o644); err != nil {
		t.Fatalf("write test_boundary.py: %v", err)
	}

	contract := mustParseContract(t, `schema: edd/contract/v2
contract_id: "FREE-SHIPPING-OVER-50"
profile: express
status: draft
intent:
  actor: "shoppers"
  problem: "shipping complaint"
  outcome:
    observable: "$50 gets free shipping"
    baseline: "all pay"
    threshold: "exactly $50 free"
    instrument: "python3 test_boundary.py"
    sample: "all carts"
    window: "pre-merge"
hypothesis:
  statement: ">= 50 makes shipping free"
  based_on: "support tickets"
  falsified_when: "$50 charged shipping"
claims:
  - id: C1
    type: functional
    phase: premerge
    blocking: true
    statement: "exactly $50 gets free shipping"
    evaluators:
      - id: E1
        type: fail_to_pass
        run: "python3 test_boundary.py"
`)

	results := runner.RunClaims(contract, "premerge", dir, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 evaluator result, got %d", len(results))
	}

	got := results[0]
	if got.Executed != true {
		t.Errorf("evaluator should be executed, got executed=%v", got.Executed)
	}
	if got.Result != "failed" {
		t.Errorf("BUGGY implementation: expected result=failed, got %q (exit %d, err %q)", got.Result, got.ExitCode, got.Error)
	}

	checkoutFixed := "def checkout_total_with_shipping(items):\n    subtotal = sum(item[\"price\"] * item[\"quantity\"] for item in items)\n    if subtotal >= 50:\n        shipping = 0\n    else:\n        shipping = 5.99\n    return subtotal + shipping\n"
	if err := os.WriteFile(filepath.Join(dir, "checkout.py"), []byte(checkoutFixed), 0o644); err != nil {
		t.Fatalf("write fixed checkout.py: %v", err)
	}

	results2 := runner.RunClaims(contract, "premerge", dir, nil)
	if len(results2) != 1 {
		t.Fatalf("expected 1 evaluator result after fix, got %d", len(results2))
	}

	got2 := results2[0]
	if got2.Result != "passed" {
		t.Errorf("FIXED implementation: expected result=passed, got %q (exit %d, err %q)", got2.Result, got2.ExitCode, got2.Error)
	}
}

package hash_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/edd-framework/edd-core/internal/hash"
)

func mustParse(t *testing.T, yamlStr string) map[string]interface{} {
	t.Helper()
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		t.Fatalf("yaml parse error: %v", err)
	}
	return doc
}

func TestCanonicalDeterministic(t *testing.T) {
	a := map[string]interface{}{"b": 1, "a": 2}
	b := map[string]interface{}{"a": 2, "b": 1}

	h1 := hash.Canonical(a)
	h2 := hash.Canonical(b)

	if h1 != h2 {
		t.Errorf("canonical hash must be deterministic regardless of key order: %s != %s", h1, h2)
	}

	if h1 == "" || len(h1) != 64 {
		t.Errorf("canonical hash must be a 64-char hex string, got %q", h1)
	}
}

func TestIntentSectionHashExcludesApproval(t *testing.T) {
	contract := mustParse(t, `
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
  approval:
    approver: "@someone"
    state: approved
    approved_section_sha: "abc123"
`)

	h := hash.IntentSection(contract)
	if h == "" || len(h) != 64 {
		t.Errorf("intent section hash must be a 64-char hex string, got %q", h)
	}

	// Same contract with different approval sha should produce same hash
	contract2 := mustParse(t, `
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
  approval:
    approver: "@someone-else"
    state: approved
    approved_section_sha: "different-sha"
`)

	h2 := hash.IntentSection(contract2)
	if h != h2 {
		t.Error("intent section hash must exclude approval block")
	}
}

func TestIntentSectionHashChangesWhenIntentChanges(t *testing.T) {
	contract1 := mustParse(t, `
intent:
  actor: "test"
  problem: "problem A"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
`)

	contract2 := mustParse(t, `
intent:
  actor: "test"
  problem: "problem B — changed"
  outcome:
    observable: "x"
    baseline: "0"
    threshold: "> 0"
    instrument: "test"
    sample: "all"
    window: "7d"
`)

	h1 := hash.IntentSection(contract1)
	h2 := hash.IntentSection(contract2)

	if h1 == h2 {
		t.Error("intent section hash must change when intent changes")
	}
}

func TestIsIntentApprovalValid(t *testing.T) {
	// Build a contract and approve it
	contract := mustParse(t, `
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
`)

	// Initially no approval = invalid
	if hash.IsIntentApprovalValid(contract) {
		t.Error("expected invalid without approval block")
	}

	// Add approval with correct sha
	if intent, ok := contract["intent"].(map[string]interface{}); ok {
		intent["approval"] = map[string]interface{}{
			"approver":             "@someone",
			"state":                "approved",
			"approved_section_sha": hash.IntentSection(contract),
		}
	}

	if !hash.IsIntentApprovalValid(contract) {
		t.Error("expected valid with correct approved_section_sha")
	}

	// Tamper with intent → approval should become invalid
	if intent, ok := contract["intent"].(map[string]interface{}); ok {
		intent["problem"] = "tampered problem"
	}

	if hash.IsIntentApprovalValid(contract) {
		t.Error("expected invalid after tampering with intent")
	}
}

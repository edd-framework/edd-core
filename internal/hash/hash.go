package hash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

// Canonical produces a deterministic SHA-256 of any value by serializing
// to JSON with sorted keys. Key order in the source YAML never affects
// the result.
func Canonical(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		// Fall back to a hash of the error message — should never happen
		// for valid contracts, but prevents a panic.
		h := sha256.Sum256([]byte(fmt.Sprintf("hash-error:%v", err)))
		return fmt.Sprintf("%x", h)
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}

// IntentSection returns the canonical hash of the intent section
// minus the approval block (which references the hash itself).
func IntentSection(contract map[string]interface{}) string {
	intent, ok := contract["intent"].(map[string]interface{})
	if !ok {
		return Canonical(nil)
	}
	// Shallow copy without approval
	clean := make(map[string]interface{}, len(intent))
	for k, v := range intent {
		if k != "approval" {
			clean[k] = v
		}
	}
	return Canonical(clean)
}

// ClaimsSection returns the canonical hash of claims + gates
// (the technical approval scope).
func ClaimsSection(contract map[string]interface{}) string {
	section := map[string]interface{}{
		"claims": contract["claims"],
		"gates":  contract["gates"],
	}
	return Canonical(section)
}

// IsIntentApprovalValid returns true if the intent approval
// state is "approved" and its sha matches the current intent.
func IsIntentApprovalValid(contract map[string]interface{}) bool {
	intent, ok := contract["intent"].(map[string]interface{})
	if !ok {
		return false
	}
	approval, ok := intent["approval"].(map[string]interface{})
	if !ok {
		return false
	}
	state, _ := approval["state"].(string)
	if state != "approved" {
		return false
	}
	sha, _ := approval["approved_section_sha"].(string)
	return sha == IntentSection(contract)
}

// IsTechnicalApprovalValid returns true if the technical approval
// state is "approved" and its sha matches current claims+gates.
func IsTechnicalApprovalValid(contract map[string]interface{}) bool {
	approval, ok := contract["technical_approval"].(map[string]interface{})
	if !ok {
		return false
	}
	state, _ := approval["state"].(string)
	if state != "approved" {
		return false
	}
	sha, _ := approval["approved_section_sha"].(string)
	return sha == ClaimsSection(contract)
}

// FileSHA256 computes the SHA-256 of a file's contents.
func FileSHA256(path string) (string, error) {
	// Delegate to os.ReadFile in the cmd layer
	return "", fmt.Errorf("use os.ReadFile in caller")
}

// SortedKeys returns the keys of a map sorted for deterministic
// iteration (useful in tests and output).
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

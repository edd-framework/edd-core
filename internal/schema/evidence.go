package schema

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type EvidenceBundle struct {
	Schema       string              `json:"schema" yaml:"schema"`
	ContractID   string              `json:"contract_id" yaml:"contract_id"`
	RunID        string              `json:"run_id" yaml:"run_id"`
	Phase        string              `json:"phase" yaml:"phase"`
	CommitSHA    string              `json:"commit_sha" yaml:"commit_sha"`
	ContractSHA  string              `json:"contract_sha" yaml:"contract_sha"`
	PolicySHA    string              `json:"policy_sha" yaml:"policy_sha"`
	GeneratedAt  string              `json:"generated_at" yaml:"generated_at"`
	Runner       RunnerInfo          `json:"runner" yaml:"runner"`
	Runtime      RuntimeInfo         `json:"runtime" yaml:"runtime"`
	Evaluators   []EvidenceEvaluator `json:"evaluators" yaml:"evaluators"`
	Gates        []EvidenceGate      `json:"gates" yaml:"gates"`
	Lockfile     *LockfileInfo       `json:"lockfile,omitempty" yaml:"lockfile,omitempty"`
	Reproducible bool                `json:"reproducible" yaml:"reproducible"`
	Decision     *Decision           `json:"decision" yaml:"decision"`
	Status       string              `json:"status" yaml:"status"`
}

type RunnerInfo struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"`
}

type RuntimeInfo struct {
	Platform string `json:"platform" yaml:"platform"`
	Go       string `json:"go" yaml:"go"`
}

type EvidenceEvaluator struct {
	ID         string `json:"id" yaml:"id"`
	ClaimID    string `json:"claim_id" yaml:"claim_id"`
	Selected   bool   `json:"selected" yaml:"selected"`
	Executed   bool   `json:"executed" yaml:"executed"`
	Result     string `json:"result,omitempty" yaml:"result,omitempty"`
	ExitCode   int    `json:"exit_code,omitempty" yaml:"exit_code,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty" yaml:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
}

type EvidenceGate struct {
	ID        string `json:"id" yaml:"id"`
	Phase     string `json:"phase,omitempty" yaml:"phase,omitempty"`
	Mandatory bool   `json:"mandatory" yaml:"mandatory"`
	Approved  bool   `json:"approved" yaml:"approved"`
	Approver  string `json:"approver,omitempty" yaml:"approver,omitempty"`
}

type LockfileInfo struct {
	Checked bool   `json:"checked" yaml:"checked"`
	Path    string `json:"path" yaml:"path"`
	SHA256  string `json:"sha256" yaml:"sha256"`
}

type Decision struct {
	Merge  string `json:"merge" yaml:"merge"`
	Close  string `json:"close" yaml:"close"`
	Status string `json:"status" yaml:"status"`
}

func ValidateEvidence(doc map[string]interface{}) []ValidationError {
	var errs []ValidationError

	schema, _ := doc["schema"].(string)
	if schema != SchemaEvidence {
		errs = append(errs, ValidationError{
			Field: "schema", Message: fmt.Sprintf("must be '%s', got %q", SchemaEvidence, schema),
		})
	}

	for _, field := range []string{"contract_id", "run_id", "phase", "commit_sha", "contract_sha", "policy_sha"} {
		_ = requireString(doc, field, &errs, "")
	}

	phase, _ := doc["phase"].(string)
	if phase != "" && !slices.Contains(evidencePhases, phase) {
		errs = append(errs, ValidationError{
			Field: "phase", Message: fmt.Sprintf("invalid value %q (expected one of %s)", phase, strings.Join(evidencePhases, ", ")),
		})
	}

	evsRaw := doc["evaluators"]
	if evsRaw == nil {
		errs = append(errs, ValidationError{Field: "evaluators", Message: "missing required field"})
	} else if evs, ok := evsRaw.([]interface{}); ok {
		for i, e := range evs {
			where := fmt.Sprintf("evaluators[%d]", i)
			ev, ok := e.(map[string]interface{})
			if !ok {
				errs = append(errs, ValidationError{Field: where, Message: "must be a mapping"})
				continue
			}
			_ = requireString(ev, "id", &errs, where)
			_ = requireString(ev, "claim_id", &errs, where)
			for _, bfield := range []string{"selected", "executed"} {
				if v, ok := ev[bfield]; !ok {
					errs = append(errs, ValidationError{Field: where + "." + bfield, Message: "must be a boolean"})
				} else if _, isBool := v.(bool); !isBool {
					errs = append(errs, ValidationError{Field: where + "." + bfield, Message: "must be a boolean"})
				}
			}
			result, _ := ev["result"].(string)
			if result != "" && !slices.Contains(evaluatorResults, result) {
				errs = append(errs, ValidationError{
					Field: where + ".result", Message: fmt.Sprintf("invalid value %q", result),
				})
			}
		}
	}

	decisionRaw := doc["decision"]
	if decisionRaw == nil {
		errs = append(errs, ValidationError{Field: "decision", Message: "missing required field"})
	} else if decision, ok := decisionRaw.(map[string]interface{}); ok {
		for _, field := range []string{"merge", "close"} {
			val, _ := decision[field].(string)
			if val != "allow" && val != "deny" {
				errs = append(errs, ValidationError{
					Field: "decision." + field, Message: fmt.Sprintf("invalid value %q (expected 'allow' or 'deny')", val),
				})
			}
		}
	}

	return errs
}

func LoadEvidenceFromBytes(data []byte) (*EvidenceBundle, []ValidationError, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("parse error: %w", err)
	}

	errs := ValidateEvidence(raw)
	if len(errs) > 0 {
		return nil, errs, nil
	}

	var eb EvidenceBundle
	if err := yaml.Unmarshal(data, &eb); err != nil {
		return nil, nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &eb, nil, nil
}

package schema

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaContract = "edd/contract/v2"
const SchemaEvidence = "edd/evidence/v2"
const SchemaPolicy = "edd/policy/v2"

var profiles = []string{"express", "standard", "governed"}
var claimTypes = []string{"functional", "regression", "product", "safety", "performance", "security"}
var claimPhases = []string{"premerge", "live"}
var evidencePhases = []string{"premerge", "live", "both"}
var evaluatorTypes = []string{"fail_to_pass", "pass_to_pass", "measurement", "structural", "human"}
var approvalStates = []string{"pending", "approved", "rejected"}
var evaluatorResults = []string{"passed", "failed", "skipped", "xfail"}
var lifecycleStates = []string{
	"draft", "ready", "verified_premerge", "merged", "deployed",
	"validated_live", "closed", "failed", "inconclusive", "waived",
	"falsified", "superseded",
}
var gateTypes = []string{"release", "architecture", "security", "destructive_action"}

type Contract struct {
	Schema            string            `yaml:"schema"`
	ContractID        string            `yaml:"contract_id"`
	Profile           string            `yaml:"profile"`
	Status            string            `yaml:"status"`
	Intent            Intent            `yaml:"intent"`
	Hypothesis        Hypothesis        `yaml:"hypothesis"`
	Claims            []Claim           `yaml:"claims"`
	Gates             []Gate            `yaml:"gates"`
	TechnicalApproval *Approval         `yaml:"technical_approval,omitempty"`
	Decision          *DecisionOverride `yaml:"decision,omitempty"`
}

type Intent struct {
	Actor       string    `yaml:"actor"`
	Problem     string    `yaml:"problem"`
	Outcome     Outcome   `yaml:"outcome"`
	Constraints []string  `yaml:"constraints,omitempty"`
	Approval    *Approval `yaml:"approval,omitempty"`
}

type Outcome struct {
	Observable string `yaml:"observable"`
	Baseline   string `yaml:"baseline"`
	Threshold  string `yaml:"threshold"`
	Instrument string `yaml:"instrument"`
	Sample     string `yaml:"sample"`
	Window     string `yaml:"window"`
}

type Approval struct {
	Approver           string `yaml:"approver"`
	State              string `yaml:"state"`
	ApprovedSectionSHA string `yaml:"approved_section_sha"`
}

type Hypothesis struct {
	Statement     string `yaml:"statement"`
	BasedOn       string `yaml:"based_on"`
	FalsifiedWhen string `yaml:"falsified_when"`
}

type Claim struct {
	ID         string      `yaml:"id"`
	Type       string      `yaml:"type"`
	Phase      string      `yaml:"phase"`
	Blocking   bool        `yaml:"blocking"`
	Statement  string      `yaml:"statement"`
	Evaluators []Evaluator `yaml:"evaluators"`
}

type Evaluator struct {
	ID     string `yaml:"id"`
	Type   string `yaml:"type"`
	Run    string `yaml:"run,omitempty"`
	Expect string `yaml:"expect,omitempty"`
}

type Gate struct {
	ID        string `yaml:"id"`
	Type      string `yaml:"type"`
	Mandatory bool   `yaml:"mandatory"`
	Reason    string `yaml:"reason,omitempty"`
}

type DecisionOverride struct {
	MergeRequires     []string `yaml:"merge_requires,omitempty"`
	LiveCloseRequires []string `yaml:"live_close_requires,omitempty"`
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

func ValidateContract(doc map[string]interface{}) []ValidationError {
	var errs []ValidationError

	schema, _ := doc["schema"].(string)
	if schema != SchemaContract {
		errs = append(errs, ValidationError{
			Field: "schema", Message: fmt.Sprintf("must be '%s', got %q", SchemaContract, schema),
		})
	}

	cid := requireString(doc, "contract_id", &errs, "")
	_ = cid

	profile := requireString(doc, "profile", &errs, "")
	if profile != "" && !slices.Contains(profiles, profile) {
		errs = append(errs, ValidationError{
			Field: "profile", Message: fmt.Sprintf("invalid value %q (expected one of %s)", profile, strings.Join(profiles, ", ")),
		})
	}

	status, _ := doc["status"].(string)
	if status != "" && !slices.Contains(lifecycleStates, status) {
		errs = append(errs, ValidationError{
			Field: "status", Message: fmt.Sprintf("invalid value %q", status),
		})
	}

	validateIntent(doc, &errs)
	validateHypothesis(doc, &errs)
	validateClaims(doc, &errs)
	validateGates(doc, &errs)
	validateApproval(doc, "technical_approval", &errs)

	return errs
}

func requireString(doc map[string]interface{}, key string, errs *[]ValidationError, where string) string {
	fullKey := key
	if where != "" {
		fullKey = where + "." + key
	}
	v, ok := doc[key]
	if !ok || v == nil || v == "" {
		*errs = append(*errs, ValidationError{Field: fullKey, Message: "missing required field"})
		return ""
	}
	s, ok := v.(string)
	if !ok || s == "" {
		*errs = append(*errs, ValidationError{Field: fullKey, Message: "missing required field"})
		return ""
	}
	return s
}

func validateIntent(doc map[string]interface{}, errs *[]ValidationError) {
	intentRaw, ok := doc["intent"]
	if !ok {
		*errs = append(*errs, ValidationError{Field: "intent", Message: "missing required field"})
		return
	}
	intent, ok := intentRaw.(map[string]interface{})
	if !ok {
		*errs = append(*errs, ValidationError{Field: "intent", Message: "must be a mapping"})
		return
	}

	_ = requireString(intent, "actor", errs, "intent")
	_ = requireString(intent, "problem", errs, "intent")

	outcomeRaw := intent["outcome"]
	if outcomeRaw == nil {
		*errs = append(*errs, ValidationError{Field: "intent.outcome", Message: "missing required field"})
	} else if outcome, ok := outcomeRaw.(map[string]interface{}); ok {
		for _, field := range []string{"observable", "baseline", "threshold", "instrument", "sample", "window"} {
			_ = requireString(outcome, field, errs, "intent.outcome")
		}
	} else {
		*errs = append(*errs, ValidationError{Field: "intent.outcome", Message: "must be a mapping"})
	}

	if approvalRaw := intent["approval"]; approvalRaw != nil {
		if approval, ok := approvalRaw.(map[string]interface{}); ok {
			validateApprovalBlock(approval, errs, "intent.approval")
		}
	}
}

func validateHypothesis(doc map[string]interface{}, errs *[]ValidationError) {
	hypRaw, ok := doc["hypothesis"]
	if !ok {
		*errs = append(*errs, ValidationError{Field: "hypothesis", Message: "missing required field"})
		return
	}
	hyp, ok := hypRaw.(map[string]interface{})
	if !ok {
		*errs = append(*errs, ValidationError{Field: "hypothesis", Message: "must be a mapping"})
		return
	}

	_ = requireString(hyp, "statement", errs, "hypothesis")
	_ = requireString(hyp, "falsified_when", errs, "hypothesis")
}

func validateClaims(doc map[string]interface{}, errs *[]ValidationError) {
	claimsRaw := doc["claims"]
	if claimsRaw == nil {
		*errs = append(*errs, ValidationError{Field: "claims", Message: "at least one claim is required"})
		return
	}
	claims, ok := claimsRaw.([]interface{})
	if !ok {
		*errs = append(*errs, ValidationError{Field: "claims", Message: "must be a list"})
		return
	}
	if len(claims) == 0 {
		*errs = append(*errs, ValidationError{Field: "claims", Message: "at least one claim is required"})
		return
	}

	seenIDs := make(map[string]bool)
	for i, c := range claims {
		where := fmt.Sprintf("claims[%d]", i)
		claim, ok := c.(map[string]interface{})
		if !ok {
			*errs = append(*errs, ValidationError{Field: where, Message: "must be a mapping"})
			continue
		}

		cid := requireString(claim, "id", errs, where)
		if cid != "" {
			if seenIDs[cid] {
				*errs = append(*errs, ValidationError{Field: where + ".id", Message: fmt.Sprintf("duplicate claim id %q", cid)})
			}
			seenIDs[cid] = true
		}

		ctype, _ := claim["type"].(string)
		if ctype != "" && !slices.Contains(claimTypes, ctype) {
			*errs = append(*errs, ValidationError{
				Field: where + ".type", Message: fmt.Sprintf("invalid value %q (expected one of %s)", ctype, strings.Join(claimTypes, ", ")),
			})
		}

		phase, _ := claim["phase"].(string)
		if phase != "" && !slices.Contains(claimPhases, phase) {
			*errs = append(*errs, ValidationError{
				Field: where + ".phase", Message: fmt.Sprintf("invalid value %q (expected one of %s)", phase, strings.Join(claimPhases, ", ")),
			})
		}

		if blocking, ok := claim["blocking"]; ok && blocking != nil {
			if _, isBool := blocking.(bool); !isBool {
				*errs = append(*errs, ValidationError{Field: where + ".blocking", Message: "must be a boolean"})
			}
		}

		_ = requireString(claim, "statement", errs, where)

		validateEvaluators(claim, where, errs)
	}
}

func validateEvaluators(claim map[string]interface{}, claimWhere string, errs *[]ValidationError) {
	evsRaw := claim["evaluators"]
	if evsRaw == nil {
		*errs = append(*errs, ValidationError{Field: claimWhere + ".evaluators", Message: "at least one evaluator is required"})
		return
	}
	evs, ok := evsRaw.([]interface{})
	if !ok || len(evs) == 0 {
		*errs = append(*errs, ValidationError{Field: claimWhere + ".evaluators", Message: "at least one evaluator is required"})
		return
	}

	for j, e := range evs {
		where := fmt.Sprintf("%s.evaluators[%d]", claimWhere, j)
		ev, ok := e.(map[string]interface{})
		if !ok {
			*errs = append(*errs, ValidationError{Field: where, Message: "must be a mapping"})
			continue
		}
		_ = requireString(ev, "id", errs, where)
		etype, _ := ev["type"].(string)
		if etype != "" && !slices.Contains(evaluatorTypes, etype) {
			*errs = append(*errs, ValidationError{
				Field: where + ".type", Message: fmt.Sprintf("invalid value %q (expected one of %s)", etype, strings.Join(evaluatorTypes, ", ")),
			})
		}
	}
}

func validateGates(doc map[string]interface{}, errs *[]ValidationError) {
	gatesRaw := doc["gates"]
	if gatesRaw == nil {
		return
	}
	gates, ok := gatesRaw.([]interface{})
	if !ok {
		*errs = append(*errs, ValidationError{Field: "gates", Message: "must be a list"})
		return
	}

	seenIDs := make(map[string]bool)
	for i, g := range gates {
		where := fmt.Sprintf("gates[%d]", i)
		gate, ok := g.(map[string]interface{})
		if !ok {
			*errs = append(*errs, ValidationError{Field: where, Message: "must be a mapping"})
			continue
		}

		gid := requireString(gate, "id", errs, where)
		if gid != "" {
			if seenIDs[gid] {
				*errs = append(*errs, ValidationError{Field: where + ".id", Message: fmt.Sprintf("duplicate gate id %q", gid)})
			}
			seenIDs[gid] = true
		}

		gtype, _ := gate["type"].(string)
		if gtype != "" && !slices.Contains(gateTypes, gtype) {
			*errs = append(*errs, ValidationError{
				Field: where + ".type", Message: fmt.Sprintf("invalid value %q (expected one of %s)", gtype, strings.Join(gateTypes, ", ")),
			})
		}

		if mandatory, ok := gate["mandatory"]; ok && mandatory != nil {
			if _, isBool := mandatory.(bool); !isBool {
				*errs = append(*errs, ValidationError{Field: where + ".mandatory", Message: "must be a boolean"})
			}
		}
	}
}

func validateApproval(doc map[string]interface{}, key string, errs *[]ValidationError) {
	raw := doc[key]
	if raw == nil {
		return
	}
	approval, ok := raw.(map[string]interface{})
	if !ok {
		*errs = append(*errs, ValidationError{Field: key, Message: "must be a mapping"})
		return
	}
	validateApprovalBlock(approval, errs, key)
}

func validateApprovalBlock(approval map[string]interface{}, errs *[]ValidationError, where string) {
	state, _ := approval["state"].(string)
	if state != "" && !slices.Contains(approvalStates, state) {
		*errs = append(*errs, ValidationError{
			Field: where + ".state", Message: fmt.Sprintf("invalid value %q (expected one of %s)", state, strings.Join(approvalStates, ", ")),
		})
	}
	if state == "approved" {
		_ = requireString(approval, "approver", errs, where)
	}
}

func LoadContractFromBytes(data []byte) (*Contract, []ValidationError, error) {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("yaml parse error: %w", err)
	}

	errs := ValidateContract(raw)
	if len(errs) > 0 {
		return nil, errs, nil
	}

	var c Contract
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, nil, fmt.Errorf("yaml unmarshal error: %w", err)
	}

	return &c, nil, nil
}

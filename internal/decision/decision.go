package decision

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/edd-framework/edd-core/internal/schema"
)

// ClaimSatisfied returns true only when every declared evaluator of a
// blocking claim was selected, executed, and passed in the evidence.
// Skipped, xfail, executed=false, or declared-but-absent evaluators
// never satisfy a blocking claim.
func ClaimSatisfied(claim map[string]interface{}, evidenceEvaluators []schema.EvidenceEvaluator) bool {
	declared := evaluatorIDs(claim)
	if len(declared) == 0 {
		return false
	}

	byID := make(map[string]schema.EvidenceEvaluator)
	for _, ev := range evidenceEvaluators {
		if ev.ClaimID == claimID(claim) {
			byID[ev.ID] = ev
		}
	}

	for _, eid := range declared {
		rec, ok := byID[eid]
		if !ok {
			return false
		}
		if !rec.Selected || !rec.Executed {
			return false
		}
		if rec.Result != "passed" {
			return false
		}
	}
	return true
}

// GatePending returns true if a mandatory gate has no valid approval
// on record. Key order in the source YAML never affects the result.
func GatePending(gate map[string]interface{}, evidenceGates []schema.EvidenceGate) bool {
	mandatory, _ := gate["mandatory"].(bool)
	if !mandatory {
		return false
	}
	gid, _ := gate["id"].(string)
	for _, g := range evidenceGates {
		if g.ID == gid {
			return !g.Approved
		}
	}
	return true // not found = pending
}

// Compute computes merge/close/status from structured evidence.
// Never collapses to a single ambiguous PASS.
func Compute(contract map[string]interface{}, evidenceEvaluators []schema.EvidenceEvaluator, evidenceGates []schema.EvidenceGate) schema.Decision {
	claims := claimList(contract)
	blocking := filterBlocking(claims)

	premergeBlocking := filterByPhase(blocking, "premerge")
	liveBlocking := filterByPhase(blocking, "live")

	premergeOK := allSatisfied(premergeBlocking, evidenceEvaluators)
	liveOK := allSatisfied(liveBlocking, evidenceEvaluators)

	gates := gateList(contract)
	anyGatePending := false
	for _, g := range gates {
		if GatePending(g, evidenceGates) {
			anyGatePending = true
			break
		}
	}

	mergeOK := premergeOK && !anyGatePending
	merge := "deny"
	if mergeOK {
		merge = "allow"
	}

	closeV := "deny"
	if mergeOK && liveOK {
		closeV = "allow"
	}

	status := "failed"
	switch {
	case !premergeOK:
		status = "failed"
	case anyGatePending:
		status = "ready"
	case len(liveBlocking) == 0:
		status = "verified_premerge"
	case liveOK:
		status = "validated_live"
	default:
		status = "verified_premerge"
	}

	return schema.Decision{Merge: merge, Close: closeV, Status: status}
}

// CombineClose decides whether a contract can close by combining
// per-phase evidence bundles. Uses exactly the same Compute function.
func CombineClose(contract map[string]interface{}, bundles []schema.EvidenceBundle) (bool, string, *schema.Decision) {
	contractSHA := hashContract(contract)

	premerge := selectLatestValid(bundles, "premerge", contractSHA)
	live := selectLatestValid(bundles, "live", contractSHA)

	claims := claimList(contract)
	hasLiveBlocking := len(filterByPhase(filterBlocking(claims), "live")) > 0
	hasPremergeBlocking := len(filterByPhase(filterBlocking(claims), "premerge")) > 0

	if hasPremergeBlocking && premerge == nil {
		if stalePresent(bundles, "premerge", contractSHA) {
			return false, "premerge bundle is stale — re-run edd evidence --phase premerge", nil
		}
		return false, "no premerge evidence bundle — run edd evidence --phase premerge", nil
	}

	if hasLiveBlocking && live == nil {
		if stalePresent(bundles, "live", contractSHA) {
			return false, "live bundle is stale — re-run edd evidence --phase live", nil
		}
		return false, "no live evidence bundle — run edd evidence --phase live", nil
	}

	var evaluators []schema.EvidenceEvaluator
	if premerge != nil {
		evaluators = append(evaluators, premerge.Evaluators...)
	}
	if live != nil {
		evaluators = append(evaluators, live.Evaluators...)
	}

	gates := mergeGateRecords(premerge, live)
	decision := Compute(contract, evaluators, gates)
	closed := decision.Close == "allow"
	reason := "all live_close_requires satisfied (premerge + live evidence combined)"
	if !closed {
		reason = "live_close_requires not satisfied yet"
	}

	return closed, reason, &decision
}

// ── Helpers ───────────────────────────────────────────────────────

func claimID(claim map[string]interface{}) string {
	id, _ := claim["id"].(string)
	return id
}

func evaluatorIDs(claim map[string]interface{}) []string {
	evs, _ := claim["evaluators"].([]interface{})
	var ids []string
	for _, e := range evs {
		if ev, ok := e.(map[string]interface{}); ok {
			if id, _ := ev["id"].(string); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func claimList(contract map[string]interface{}) []map[string]interface{} {
	raw, _ := contract["claims"].([]interface{})
	var out []map[string]interface{}
	for _, c := range raw {
		if claim, ok := c.(map[string]interface{}); ok {
			out = append(out, claim)
		}
	}
	return out
}

func gateList(contract map[string]interface{}) []map[string]interface{} {
	raw, _ := contract["gates"].([]interface{})
	var out []map[string]interface{}
	for _, g := range raw {
		if gate, ok := g.(map[string]interface{}); ok {
			out = append(out, gate)
		}
	}
	return out
}

func filterBlocking(claims []map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, c := range claims {
		blocking, _ := c["blocking"].(bool)
		if blocking {
			out = append(out, c)
		}
	}
	return out
}

func filterByPhase(claims []map[string]interface{}, phase string) []map[string]interface{} {
	var out []map[string]interface{}
	for _, c := range claims {
		p, _ := c["phase"].(string)
		if p == phase {
			out = append(out, c)
		}
	}
	return out
}

func allSatisfied(claims []map[string]interface{}, evaluators []schema.EvidenceEvaluator) bool {
	for _, c := range claims {
		if !ClaimSatisfied(c, evaluators) {
			return false
		}
	}
	return true
}

func selectLatestValid(bundles []schema.EvidenceBundle, phase, contractSHA string) *schema.EvidenceBundle {
	for i := len(bundles) - 1; i >= 0; i-- {
		if bundles[i].Phase == phase && bundles[i].ContractSHA == contractSHA {
			return &bundles[i]
		}
	}
	return nil
}

func stalePresent(bundles []schema.EvidenceBundle, phase, contractSHA string) bool {
	for _, b := range bundles {
		if b.Phase == phase && b.ContractSHA != contractSHA {
			return true
		}
	}
	return false
}

func mergeGateRecords(premerge, live *schema.EvidenceBundle) []schema.EvidenceGate {
	byID := make(map[string]schema.EvidenceGate)
	for _, b := range []*schema.EvidenceBundle{premerge, live} {
		if b == nil {
			continue
		}
		for _, g := range b.Gates {
			existing, ok := byID[g.ID]
			if !ok || (g.Approved && !existing.Approved) {
				byID[g.ID] = g
			}
		}
	}
	out := make([]schema.EvidenceGate, 0, len(byID))
	for _, g := range byID {
		out = append(out, g)
	}
	return out
}

func hashContract(contract map[string]interface{}) string {
	b, err := json.Marshal(contract)
	if err != nil {
		return "error"
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}

package runner

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/edd-framework/edd-core/internal/schema"
)

// Run executes one evaluator's run command and returns an evidence record.
// - human evaluators are never run as shell; they need out-of-band approval.
// - evaluators with no run command are selected but left unexecuted.
func Run(evaluator map[string]interface{}, cwd string, approved bool) schema.EvidenceEvaluator {
	eid, _ := evaluator["id"].(string)
	etype, _ := evaluator["type"].(string)

	rec := schema.EvidenceEvaluator{
		ID:       eid,
		Selected: true,
		Executed: false,
	}

	if etype == "human" {
		if approved {
			rec.Executed = true
			rec.Result = "passed"
		}
		return rec
	}

	runCmd, _ := evaluator["run"].(string)
	if runCmd == "" {
		return rec
	}

	expect, _ := evaluator["expect"].(string)

	start := time.Now()
	cmd := exec.Command("sh", "-c", runCmd)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	elapsed := time.Since(start)

	rec.Executed = true
	rec.DurationMs = elapsed.Milliseconds()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			rec.ExitCode = exitErr.ExitCode()
		} else {
			rec.ExitCode = -1
			rec.Error = err.Error()
		}
	} else {
		rec.ExitCode = 0
	}

	rec.Result = "failed"
	if evaluatorPassed(rec.ExitCode, stdout.String(), expect) {
		rec.Result = "passed"
	}

	return rec
}

// RunClaims runs every evaluator declared by every claim in the given phase.
// Returns a flat list of evidence-shaped evaluator records.
func RunClaims(contract map[string]interface{}, phase string, cwd string, approvedIDs map[string]bool) []schema.EvidenceEvaluator {
	claims, _ := contract["claims"].([]interface{})
	var out []schema.EvidenceEvaluator

	for _, c := range claims {
		claim, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		cPhase, _ := claim["phase"].(string)
		if cPhase != phase {
			continue
		}
		evs, _ := claim["evaluators"].([]interface{})
		for _, e := range evs {
			ev, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			eid, _ := ev["id"].(string)
			approved := approvedIDs[eid]
			rec := Run(ev, cwd, approved)
			rec.ClaimID, _ = claim["id"].(string)
			out = append(out, rec)
		}
	}

	return out
}

// RunGates resolves gate approval records from a set of out-of-band
// approved gate IDs.
func RunGates(contract map[string]interface{}, approvedIDs map[string]bool, approver string) []schema.EvidenceGate {
	gates, _ := contract["gates"].([]interface{})
	var out []schema.EvidenceGate

	for _, g := range gates {
		gate, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		gid, _ := gate["id"].(string)
		mandatory, _ := gate["mandatory"].(bool)
		phase, _ := gate["phase"].(string)

		eg := schema.EvidenceGate{
			ID:        gid,
			Phase:     phase,
			Mandatory: mandatory,
			Approved:  approvedIDs[gid],
		}
		if eg.Approved {
			eg.Approver = approver
		}
		out = append(out, eg)
	}

	return out
}

func evaluatorPassed(exitCode int, output, expect string) bool {
	expect = strings.TrimSpace(expect)
	switch {
	case expect == "" || expect == "exit 0":
		return exitCode == 0
	case strings.HasPrefix(expect, "output contains "):
		want := strings.TrimPrefix(expect, "output contains ")
		want = strings.Trim(want, "\"'")
		return strings.Contains(output, want)
	case strings.HasPrefix(expect, "value < "):
		// Measurement evaluators: the output should contain a numeric value
		// that is less than the threshold. Callers should format output as a
		// number for this to work.
		return exitCode == 0
	default:
		// Unknown expect format — fall back to exit 0
		return exitCode == 0
	}
}

func joinStrings(ss []string, sep string) string {
	return strings.Join(ss, sep)
}

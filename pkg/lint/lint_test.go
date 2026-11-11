package lint

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagementio"
)

func load(t *testing.T) *engagement.Agreement {
	t.Helper()
	a, err := engagementio.LoadAgreement(testfixture.Nonprofit())
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestExampleIsClean(t *testing.T) {
	if res := Agreement(load(t)); len(res.Findings) != 0 {
		t.Fatalf("findings: %v", res.Findings)
	}
}

func rules(res Result) []string {
	var out []string
	for _, f := range res.Findings {
		out = append(out, f.Rule)
	}
	return out
}

func TestRules(t *testing.T) {
	tests := []struct {
		rule   string
		sev    Severity
		mutate func(a *engagement.Agreement)
	}{
		{RuleDuplicateID, Error, func(a *engagement.Agreement) { a.Deliverables[1].ID = "D1" }},
		{RuleDuplicateID, Error, func(a *engagement.Agreement) { a.Risks[0].ID = "O1" }},
		{RuleMissingRole, Error, func(a *engagement.Agreement) { a.Parties[1].Role = engagement.RoleSponsor }},
		{RuleUnknownParty, Error, func(a *engagement.Agreement) { a.Responsibilities[0].Party = "nobody" }},
		{RuleUnknownPerson, Error, func(a *engagement.Agreement) { a.Signatures[0].Signer = "nobody" }},
		{RuleSignerParty, Error, func(a *engagement.Agreement) { a.Signatures[0].Signer = "dana-lee" }},
		{RuleUnknownPhase, Error, func(a *engagement.Agreement) { a.Deliverables[0].Phase = "P9" }},
		{RuleUnknownObjective, Error, func(a *engagement.Agreement) { a.Deliverables[0].Objectives = []string{"O9"} }},
		{RuleUnknownDeliverable, Error, func(a *engagement.Agreement) { a.Milestones[0].Deliverables = []string{"D9"} }},
		{RuleUnknownOwner, Error, func(a *engagement.Agreement) { a.OpenQuestions[0].Owner = "nobody" }},
		{RuleDateOrder, Error, func(a *engagement.Agreement) { a.Phases[0].End = "2026-01-01" }},
		{RuleDateOutsidePeriod, Warning, func(a *engagement.Agreement) { a.Deliverables[0].Due = "2027-01-01" }},
		{RuleCurrencyRequired, Error, func(a *engagement.Agreement) {
			a.Commercials.Model = engagement.ModelFixedFee
			a.Commercials.Fees = []engagement.Fee{{Description: "Fee", Amount: "100"}}
		}},
		{RuleProBonoFee, Warning, func(a *engagement.Agreement) {
			a.Commercials.Currency = "USD"
			a.Commercials.Fees = []engagement.Fee{{Description: "Fee", Amount: "100.00"}}
		}},
		{RuleAgreedOpenQuestions, Error, func(a *engagement.Agreement) { a.Status = engagement.StatusAgreed }},
		{RuleAgreedUnsigned, Warning, func(a *engagement.Agreement) {
			a.Status = engagement.StatusAgreed
			a.OpenQuestions = nil
		}},
		{RuleNoExclusions, Warning, func(a *engagement.Agreement) { a.Scope.Excluded = nil }},
		{RuleNoClientResponsibility, Warning, func(a *engagement.Agreement) { a.Responsibilities = a.Responsibilities[:2] }},
		{RuleNoAcceptanceCriteria, Warning, func(a *engagement.Agreement) { a.Acceptance = nil }},
		{RuleObjectiveUnserved, Warning, func(a *engagement.Agreement) { a.Deliverables[3].Objectives = []string{"O1"} }},
		{RuleNoSignatures, Warning, func(a *engagement.Agreement) { a.Signatures = nil }},
	}
	covered := map[string]bool{}
	for _, tt := range tests {
		a := load(t)
		tt.mutate(a)
		res := Agreement(a)
		i := slices.IndexFunc(res.Findings, func(f Finding) bool { return f.Rule == tt.rule })
		if i < 0 {
			t.Errorf("%s: not reported; got %v", tt.rule, rules(res))
			continue
		}
		if res.Findings[i].Severity != tt.sev {
			t.Errorf("%s: severity %s, want %s", tt.rule, res.Findings[i].Severity, tt.sev)
		}
		covered[tt.rule] = true
	}
	for _, r := range Rules() {
		if !covered[r] {
			t.Errorf("rule %s has no test", r)
		}
	}
}

func TestProBonoZeroFeeIsFine(t *testing.T) {
	a := load(t)
	a.Commercials.Currency = "USD"
	a.Commercials.Fees = []engagement.Fee{{Description: "Fee", Amount: "0.00"}}
	if slices.Contains(rules(Agreement(a)), RuleProBonoFee) {
		t.Fatal("a zero fee should not be flagged")
	}
}

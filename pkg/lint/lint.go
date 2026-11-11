// Package lint checks an agreement for problems a JSON Schema cannot see:
// unique IDs, references that resolve, dates in order, and gaps that tend
// to cause disputes, such as no exclusions or no client responsibilities.
package lint

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
)

// Severity is how serious a finding is.
type Severity string

// Severities.
const (
	// Error findings make an agreement unfit to send for signature.
	Error Severity = "error"
	// Warning findings are worth fixing but do not block.
	Warning Severity = "warning"
)

// Rule IDs.
const (
	RuleDuplicateID            = "duplicate-id"
	RuleMissingRole            = "missing-role"
	RuleUnknownParty           = "unknown-party"
	RuleUnknownPerson          = "unknown-person"
	RuleSignerParty            = "signer-party"
	RuleUnknownPhase           = "unknown-phase"
	RuleUnknownObjective       = "unknown-objective"
	RuleUnknownDeliverable     = "unknown-deliverable"
	RuleUnknownOwner           = "unknown-owner"
	RuleDateOrder              = "date-order"
	RuleDateOutsidePeriod      = "date-outside-period"
	RuleCurrencyRequired       = "currency-required"
	RuleProBonoFee             = "pro-bono-fee"
	RuleAgreedOpenQuestions    = "agreed-open-questions"
	RuleAgreedUnsigned         = "agreed-unsigned"
	RuleNoExclusions           = "no-exclusions"
	RuleNoClientResponsibility = "no-client-responsibility"
	RuleNoAcceptanceCriteria   = "no-acceptance-criteria"
	RuleObjectiveUnserved      = "objective-unserved"
	RuleNoSignatures           = "no-signatures"
)

// Finding is one problem.
type Finding struct {
	Severity Severity `json:"severity"`
	Rule     string   `json:"rule"`
	// Path locates the element, such as "deliverables/D3/phase".
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (f Finding) String() string {
	return fmt.Sprintf("%s %s: %s (%s)", f.Severity, f.Path, f.Message, f.Rule)
}

// Result is the outcome of linting one agreement.
type Result struct {
	Findings []Finding `json:"findings"`
}

// Errors returns the number of error findings.
func (r Result) Errors() int { return r.count(Error) }

// Warnings returns the number of warning findings.
func (r Result) Warnings() int { return r.count(Warning) }

func (r Result) count(s Severity) int {
	n := 0
	for _, f := range r.Findings {
		if f.Severity == s {
			n++
		}
	}
	return n
}

type linter struct {
	a   *engagement.Agreement
	out []Finding
}

func (l *linter) add(s Severity, rule, path, format string, args ...any) {
	l.out = append(l.out, Finding{Severity: s, Rule: rule, Path: path, Message: fmt.Sprintf(format, args...)})
}

// Agreement lints a. Findings come in a fixed order: the order of the
// checks, then document order, so output is deterministic.
func Agreement(a *engagement.Agreement) Result {
	l := &linter{a: a}
	l.ids()
	l.roles()
	l.references()
	l.dates()
	l.commercials()
	l.lifecycle()
	l.completeness()
	return Result{Findings: l.out}
}

// ids checks that every identified element has a unique ID. Parties,
// people, and all other elements share one namespace, so a reference is
// never ambiguous.
func (l *linter) ids() {
	a := l.a
	seen := map[string]string{}
	check := func(section, id string) {
		path := section + "/" + id
		if prev, ok := seen[id]; ok {
			l.add(Error, RuleDuplicateID, path, "ID %q is already used at %s", id, prev)
			return
		}
		seen[id] = path
	}
	for _, x := range a.Parties {
		check("parties", x.ID)
	}
	for _, x := range a.People {
		check("people", x.ID)
	}
	for _, x := range a.Objectives {
		check("objectives", x.ID)
	}
	for _, x := range a.Scope.Included {
		check("scope/included", x.ID)
	}
	for _, x := range a.Scope.Excluded {
		check("scope/excluded", x.ID)
	}
	for _, x := range a.Phases {
		check("phases", x.ID)
	}
	for _, x := range a.Deliverables {
		check("deliverables", x.ID)
	}
	for _, x := range a.Milestones {
		check("milestones", x.ID)
	}
	for _, x := range a.Responsibilities {
		check("responsibilities", x.ID)
	}
	for _, x := range a.Assumptions {
		check("assumptions", x.ID)
	}
	for _, x := range a.Risks {
		check("risks", x.ID)
	}
	for _, x := range a.SuccessCriteria {
		check("successCriteria", x.ID)
	}
	for _, x := range a.Terms {
		check("terms", x.ID)
	}
	for _, x := range a.OpenQuestions {
		check("openQuestions", x.ID)
	}
}

func (l *linter) roles() {
	if len(l.a.PartiesByRole(engagement.RoleConsultant)) == 0 {
		l.add(Error, RuleMissingRole, "parties", "no party has the role %q", engagement.RoleConsultant)
	}
	if len(l.a.PartiesByRole(engagement.RoleClient)) == 0 {
		l.add(Error, RuleMissingRole, "parties", "no party has the role %q", engagement.RoleClient)
	}
}

func (l *linter) references() {
	a := l.a
	party := func(path, id string) {
		if _, ok := a.Party(id); !ok {
			l.add(Error, RuleUnknownParty, path, "unknown party %q", id)
		}
	}
	phase := func(path, id string) {
		if id == "" {
			return
		}
		if _, ok := a.Phase(id); !ok {
			l.add(Error, RuleUnknownPhase, path, "unknown phase %q", id)
		}
	}
	objectives := map[string]bool{}
	for _, o := range a.Objectives {
		objectives[o.ID] = true
	}
	deliverables := map[string]bool{}
	for _, d := range a.Deliverables {
		deliverables[d.ID] = true
	}

	for _, p := range a.People {
		party("people/"+p.ID+"/party", p.Party)
	}
	for _, d := range a.Deliverables {
		base := "deliverables/" + d.ID
		if d.Owner != "" {
			party(base+"/owner", d.Owner)
		}
		phase(base+"/phase", d.Phase)
		for _, o := range d.Objectives {
			if !objectives[o] {
				l.add(Error, RuleUnknownObjective, base+"/objectives", "unknown objective %q", o)
			}
		}
	}
	for _, m := range a.Milestones {
		base := "milestones/" + m.ID
		phase(base+"/phase", m.Phase)
		for _, d := range m.Deliverables {
			if !deliverables[d] {
				l.add(Error, RuleUnknownDeliverable, base+"/deliverables", "unknown deliverable %q", d)
			}
		}
	}
	for _, r := range a.Responsibilities {
		party("responsibilities/"+r.ID+"/party", r.Party)
	}
	for _, r := range a.Risks {
		if r.Owner != "" {
			party("risks/"+r.ID+"/owner", r.Owner)
		}
	}
	for _, q := range a.OpenQuestions {
		if q.Owner == "" {
			continue
		}
		_, isParty := a.Party(q.Owner)
		_, isPerson := a.Person(q.Owner)
		if !isParty && !isPerson {
			l.add(Error, RuleUnknownOwner, "openQuestions/"+q.ID+"/owner", "owner %q is neither a party nor a person", q.Owner)
		}
	}
	for i, s := range a.Signatures {
		base := fmt.Sprintf("signatures/%d", i)
		party(base+"/party", s.Party)
		p, ok := a.Person(s.Signer)
		if !ok {
			l.add(Error, RuleUnknownPerson, base+"/signer", "unknown person %q", s.Signer)
			continue
		}
		if p.Party != s.Party {
			l.add(Error, RuleSignerParty, base+"/signer", "%s belongs to party %q, not %q", p.Name, p.Party, s.Party)
		}
	}
}

func parseDate(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.DateOnly, s)
	return t, err == nil
}

func (l *linter) dates() {
	a := l.a
	order := func(path, start, end string) {
		s, okS := parseDate(start)
		e, okE := parseDate(end)
		if okS && okE && e.Before(s) {
			l.add(Error, RuleDateOrder, path, "end %s is before start %s", end, start)
		}
	}
	if a.Period != nil {
		order("period", a.Period.Start, a.Period.End)
	}
	if a.ValidUntil != "" {
		order("validUntil", a.Date, a.ValidUntil)
	}
	for _, p := range a.Phases {
		order("phases/"+p.ID, p.Start, p.End)
	}

	if a.Period == nil {
		return
	}
	start, okS := parseDate(a.Period.Start)
	end, okE := parseDate(a.Period.End)
	if !okS || !okE {
		return
	}
	inside := func(path, date string) {
		d, ok := parseDate(date)
		if ok && (d.Before(start) || d.After(end)) {
			l.add(Warning, RuleDateOutsidePeriod, path, "%s is outside the engagement period %s to %s", date, a.Period.Start, a.Period.End)
		}
	}
	for _, p := range a.Phases {
		inside("phases/"+p.ID+"/start", p.Start)
		inside("phases/"+p.ID+"/end", p.End)
	}
	for _, d := range a.Deliverables {
		inside("deliverables/"+d.ID+"/due", d.Due)
	}
	for _, m := range a.Milestones {
		inside("milestones/"+m.ID+"/date", m.Date)
	}
}

func (l *linter) commercials() {
	c := l.a.Commercials
	if len(c.Fees) > 0 && c.Currency == "" {
		l.add(Error, RuleCurrencyRequired, "commercials/currency", "fees are listed but no currency is set")
	}
	if c.Model == engagement.ModelProBono {
		for i, f := range c.Fees {
			if strings.Trim(f.Amount, "0.") != "" {
				l.add(Warning, RuleProBonoFee, fmt.Sprintf("commercials/fees/%d", i), "pro bono engagement lists a fee of %s", f.Amount)
			}
		}
	}
}

func (l *linter) lifecycle() {
	a := l.a
	if a.Status != engagement.StatusAgreed {
		return
	}
	if len(a.OpenQuestions) > 0 {
		l.add(Error, RuleAgreedOpenQuestions, "openQuestions", "status is agreed but %d open questions remain", len(a.OpenQuestions))
	}
	for i, s := range a.Signatures {
		if s.SignedDate == "" {
			l.add(Warning, RuleAgreedUnsigned, fmt.Sprintf("signatures/%d", i), "status is agreed but %s has not signed", s.Signer)
		}
	}
}

func (l *linter) completeness() {
	a := l.a
	if len(a.Scope.Excluded) == 0 {
		l.add(Warning, RuleNoExclusions, "scope/excluded", "nothing is listed as out of scope; unstated exclusions are a common source of disputes")
	}
	clientResp := false
	for _, r := range a.Responsibilities {
		if p, ok := a.Party(r.Party); ok && p.Role == engagement.RoleClient {
			clientResp = true
			break
		}
	}
	if !clientResp {
		l.add(Warning, RuleNoClientResponsibility, "responsibilities", "no responsibilities are assigned to the client")
	}
	if a.Acceptance == nil {
		for _, d := range a.Deliverables {
			if len(d.AcceptanceCriteria) == 0 {
				l.add(Warning, RuleNoAcceptanceCriteria, "deliverables/"+d.ID, "no acceptance criteria and no acceptance process")
			}
		}
	}
	linked := map[string]bool{}
	anyLinks := false
	for _, d := range a.Deliverables {
		for _, o := range d.Objectives {
			linked[o] = true
			anyLinks = true
		}
	}
	if anyLinks {
		for _, o := range a.Objectives {
			if !linked[o.ID] {
				l.add(Warning, RuleObjectiveUnserved, "objectives/"+o.ID, "no deliverable serves this objective")
			}
		}
	}
	if len(a.Signatures) == 0 {
		l.add(Warning, RuleNoSignatures, "signatures", "no signature blocks; the rendered document will have no place to sign")
	}
}

// Rules returns every rule ID, sorted.
func Rules() []string {
	r := []string{
		RuleDuplicateID, RuleMissingRole, RuleUnknownParty, RuleUnknownPerson, RuleSignerParty,
		RuleUnknownPhase, RuleUnknownObjective, RuleUnknownDeliverable, RuleUnknownOwner,
		RuleDateOrder, RuleDateOutsidePeriod, RuleCurrencyRequired, RuleProBonoFee,
		RuleAgreedOpenQuestions, RuleAgreedUnsigned, RuleNoExclusions, RuleNoClientResponsibility,
		RuleNoAcceptanceCriteria, RuleObjectiveUnserved, RuleNoSignatures,
	}
	slices.Sort(r)
	return r
}

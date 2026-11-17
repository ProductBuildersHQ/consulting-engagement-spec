// Package render turns an agreement into a document. Markdown is the
// primary output: Pandoc converts it to DOCX or PDF, and GitHub displays
// it as is. Output is deterministic: the same agreement and options always
// produce the same bytes.
package render

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/crosswalk"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagement"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/engagementio"
	"github.com/ProductBuildersHQ/consulting-engagement-spec/pkg/outline"
)

// Flavor is the Markdown dialect to write.
type Flavor string

// Flavors.
const (
	// FlavorPandoc writes a YAML metadata block for the title and raw page
	// breaks for DOCX and PDF output. It is the input to Pandoc.
	FlavorPandoc Flavor = "pandoc"
	// FlavorGFM writes GitHub Flavored Markdown with no raw blocks, for
	// reading in a repository.
	FlavorGFM Flavor = "gfm"
)

// Options control rendering.
type Options struct {
	Flavor Flavor
	// OmitDigest leaves the content digest out of the document.
	OmitDigest bool
	// Outline lays the agreement out; the full layout when nil.
	Outline *engagement.Outline
	// Crosswalk, when set, appends an appendix showing where each section
	// of this outline is answered in the rendered document.
	Crosswalk *engagement.Outline
}

// Markdown renders a as a Markdown document.
func Markdown(a *engagement.Agreement, opts Options) ([]byte, error) {
	if opts.Flavor == "" {
		opts.Flavor = FlavorPandoc
	}
	if opts.Flavor != FlavorPandoc && opts.Flavor != FlavorGFM {
		return nil, fmt.Errorf("render: unknown flavor %q", opts.Flavor)
	}
	o := opts.Outline
	if o == nil {
		o = outline.Full()
	}
	layout, err := resolve(o)
	if err != nil {
		return nil, err
	}
	digest := ""
	if !opts.OmitDigest {
		d, err := engagementio.ContentDigest(a)
		if err != nil {
			return nil, err
		}
		digest = d
	}
	w := &writer{a: a, opts: opts, digest: digest, outline: o, lvl: 3}
	w.title()
	if o.ShowsInformation() {
		w.info()
	}
	for _, sec := range layout {
		w.writeSection(sec)
	}
	if opts.Crosswalk != nil {
		cw, err := crosswalk.Build(a, o, opts.Crosswalk)
		if err != nil {
			return nil, err
		}
		w.pageBreak()
		w.p("## Appendix: Correspondence with %s\n\n", cw.Target)
		w.p("%s", cw.Markdown())
	}
	return []byte(strings.TrimRight(w.b.String(), "\n") + "\n"), nil
}

// layoutSection is an outline section with its references parsed.
type layoutSection struct {
	engagement.OutlineSection
	refs []outline.Ref
}

func resolve(o *engagement.Outline) ([]layoutSection, error) {
	out := make([]layoutSection, 0, len(o.Sections))
	for _, s := range o.Sections {
		ls := layoutSection{OutlineSection: s}
		for _, name := range s.From {
			r, err := outline.ParseRef(name)
			if err != nil {
				return nil, fmt.Errorf("render: outline %s, section %q: %w", o.ID, s.Heading, err)
			}
			ls.refs = append(ls.refs, r)
		}
		out = append(out, ls)
	}
	return out, nil
}

type writer struct {
	a       *engagement.Agreement
	opts    Options
	digest  string
	outline *engagement.Outline
	b       strings.Builder
	section int
	// lvl is the heading level for headings inside a block.
	lvl int
}

func (w *writer) p(format string, args ...any) {
	fmt.Fprintf(&w.b, format, args...)
}

func (w *writer) para(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	w.p("%s\n\n", s)
}

func (w *writer) h2(title string) {
	if !w.outline.IsNumbered() {
		w.p("## %s\n\n", title)
		return
	}
	w.section++
	w.p("## %d. %s\n\n", w.section, title)
}

// h3 writes a heading inside a block.
func (w *writer) h3(title string) {
	w.p("%s %s\n\n", strings.Repeat("#", w.lvl), title)
}

// writeSection writes one outline section: nothing when none of its blocks has
// content, otherwise the heading, the outline's note, and each block with
// content. The first block listed prints directly under the heading; each
// later one under its own subheading.
func (w *writer) writeSection(sec layoutSection) {
	var shown []outline.Ref
	for _, r := range sec.refs {
		if outline.HasContent(w.a, r) {
			shown = append(shown, r)
		}
	}
	if len(shown) == 0 {
		return
	}
	for _, r := range shown {
		if r.Block.Name == "signatures" {
			w.pageBreak()
			break
		}
	}
	w.h2(sec.Heading)
	w.para(sec.Note)
	for _, r := range shown {
		if r.Block.Name == sec.refs[0].Block.Name && r.Value == sec.refs[0].Value {
			w.lvl = 3
		} else {
			w.lvl = 3
			w.h3(r.Block.Title)
			w.lvl = 4
		}
		w.block(r)
	}
	w.lvl = 3
}

func (w *writer) block(r outline.Ref) {
	a := w.a
	switch r.Block.Name {
	case "summary":
		w.para(a.Summary)
	case "background":
		w.para(a.Background)
	case "objectives":
		w.objectives()
	case "hypotheses":
		w.hypotheses()
	case "scope":
		w.scope()
	case "scope.included":
		w.included()
	case "scope.excluded":
		w.excluded()
	case "approach":
		w.para(a.Approach)
	case "phases":
		w.phases()
	case "deliverables":
		w.deliverables()
	case "deliverables.list":
		w.deliverableList()
	case "milestones":
		w.milestones(outline.Milestones(a, r.Value))
	case "responsibilities":
		w.responsibilities(r.Value)
	case "people":
		w.team()
	case "governance":
		w.governance()
	case "governance.contacts":
		w.contacts(outline.Contacts(a, r.Value))
	case "assumptions":
		w.assumptions()
	case "risks":
		w.risks()
	case "successCriteria":
		w.successCriteria()
	case "acceptance":
		w.acceptance()
	case "closeout":
		w.closeout()
	case "commercials":
		w.commercials()
	case "changeControl":
		w.para(a.ChangeControl)
	case "terms":
		w.terms()
	case "openQuestions":
		w.openQuestions()
	case "signatures":
		w.signatures(outline.Signatures(a, r.Value))
	}
}

func (w *writer) pageBreak() {
	if w.opts.Flavor != FlavorPandoc {
		return
	}
	w.raw(`\newpage`)
	w.rawDOCX(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
}

func (w *writer) title() {
	a := w.a
	if w.opts.Flavor == FlavorPandoc {
		w.p("---\n")
		w.p("title: %s\n", yamlString(a.Title))
		if a.Subtitle != "" {
			w.p("subtitle: %s\n", yamlString(a.Subtitle))
		}
		w.p("date: %s\n", yamlString(formatDate(a.Date)))
		w.p("lang: en-US\n")
		w.p("geometry: margin=1in\n")
		w.p("fontsize: 11pt\n")
		w.p("colorlinks: true\n")
		// The footer names the document and version on every page of the
		// PDF, and the layout when it is not the full agreement. The ID
		// pattern allows only characters that are safe in TeX.
		footer := fmt.Sprintf("%s %s, version %s", a.Kind.Label(), a.ID, a.Version)
		if w.outline.ID != outline.NameFull {
			footer += ", " + texEscape(w.outline.Title)
		}
		w.p("header-includes:\n")
		w.p("  - |\n")
		w.p("    ```{=latex}\n")
		w.p("    \\usepackage{fancyhdr}\n")
		w.p("    \\pagestyle{fancy}\n")
		w.p("    \\fancyhf{}\n")
		w.p("    \\renewcommand{\\headrulewidth}{0pt}\n")
		w.p("    \\fancyfoot[L]{\\footnotesize %s}\n", footer)
		w.p("    \\fancyfoot[R]{\\footnotesize Page \\thepage}\n")
		w.p("    ```\n")
		w.p("---\n\n")
		return
	}
	w.p("# %s\n\n", a.Title)
	if a.Subtitle != "" {
		w.p("*%s*\n\n", a.Subtitle)
	}
}

func (w *writer) info() {
	a := w.a
	rows := [][2]string{
		{"Document", fmt.Sprintf("%s, version %s (%s)", a.Kind.Label(), a.Version, a.Status.Label())},
	}
	if w.outline.ID != outline.NameFull {
		rows = append(rows, [2]string{"Layout", w.outline.Title})
	}
	rows = append(rows, [2]string{"Reference", "`" + a.ID + "`"})
	if names := w.partyNames(engagement.RoleConsultant); names != "" {
		rows = append(rows, [2]string{"Prepared by", names})
	}
	if names := w.partyNames(engagement.RoleClient); names != "" {
		rows = append(rows, [2]string{"Prepared for", names})
	}
	for _, role := range []engagement.PartyRole{engagement.RoleSponsor, engagement.RolePartner} {
		if names := w.partyNames(role); names != "" {
			label := "Sponsor"
			if role == engagement.RolePartner {
				label = "Partner"
			}
			rows = append(rows, [2]string{label, names})
		}
	}
	rows = append(rows, [2]string{"Date", formatDate(a.Date)})
	if a.ValidUntil != "" {
		rows = append(rows, [2]string{"Valid until", formatDate(a.ValidUntil)})
	}
	if a.Period != nil {
		rows = append(rows, [2]string{"Engagement period", w.dateRange(a.Period.Start, a.Period.End)})
	}
	if a.Supersedes != nil {
		rows = append(rows, [2]string{"Supersedes", fmt.Sprintf("`%s` version %s", a.Supersedes.ID, a.Supersedes.Version)})
	}
	if w.digest != "" {
		rows = append(rows, [2]string{"Content digest", "`" + shortDigest(w.digest) + "`"})
	}
	w.p("|   |   |\n|---|---|\n")
	for _, r := range rows {
		w.p("| **%s** | %s |\n", r[0], cell(r[1]))
	}
	w.p("\n")
}

func (w *writer) partyNames(role engagement.PartyRole) string {
	var names []string
	for _, p := range w.a.PartiesByRole(role) {
		names = append(names, p.Name)
	}
	return strings.Join(names, "; ")
}

func (w *writer) objectives() {
	for _, o := range w.a.Objectives {
		w.p("- %s\n", labeled(o.ID, o.Title, o.Description))
	}
	w.p("\n")
}

func (w *writer) hypotheses() {
	a := w.a
	w.para("The engagement tests these working hypotheses. The evidence the work produces confirms, revises, or rejects each one.")
	for _, h := range a.Hypotheses {
		line := labeled(h.ID, h.Statement, h.Test)
		if len(h.Deliverables) > 0 {
			line += " *Evidence:* " + strings.Join(h.Deliverables, ", ") + "."
		}
		if m, ok := a.Milestone(h.Milestone); ok {
			judged := m.Title
			if m.Date != "" {
				judged += ", " + w.date(m.Date)
			}
			line += " *Judged at:* " + judged + "."
		}
		w.p("- %s\n", line)
	}
	w.p("\n")
}

func (w *writer) scope() {
	s := w.a.Scope
	w.para(s.Summary)
	w.h3("In scope")
	w.included()
	if len(s.Excluded) > 0 {
		w.h3("Out of scope")
		w.excluded()
	}
}

func (w *writer) included() {
	for _, it := range w.a.Scope.Included {
		w.p("- %s\n", labeled(it.ID, it.Title, it.Description))
	}
	w.p("\n")
}

func (w *writer) excluded() {
	w.para("The following are not part of this engagement unless the parties agree to add them through the change process.")
	for _, it := range w.a.Scope.Excluded {
		w.p("- %s\n", labeled(it.ID, it.Title, it.Description))
	}
	w.p("\n")
}

func (w *writer) phases() {
	w.p("| Phase | Dates | Goal |\n|------------|----------------------------|------------------------------------|\n")
	for _, ph := range w.a.Phases {
		w.p("| %s | %s | %s |\n", cell(ph.Title), cell(w.dateRange(ph.Start, ph.End)), cell(ph.Goal))
	}
	w.p("\n")
}

func (w *writer) deliverables() {
	a := w.a
	for _, d := range a.Deliverables {
		var meta []string
		if d.Phase != "" {
			if ph, ok := a.Phase(d.Phase); ok {
				meta = append(meta, ph.Title)
			}
		}
		if d.Due != "" {
			meta = append(meta, "due "+w.date(d.Due))
		}
		if owner := a.DeliverableOwner(d); owner != "" {
			if c, ok := a.Consultant(); !ok || owner != c.ID {
				if p, ok := a.Party(owner); ok {
					meta = append(meta, "by "+p.Name)
				}
			}
		}
		lines := []string{fmt.Sprintf("**%s. %s**", d.ID, strings.TrimSpace(d.Title))}
		if len(meta) > 0 {
			lines[0] += " · " + strings.Join(meta, " · ")
		}
		if desc := oneLine(d.Description); desc != "" {
			lines = append(lines, desc)
		}
		if d.Format != "" {
			lines = append(lines, "*Format:* "+oneLine(d.Format))
		}
		if len(d.AcceptanceCriteria) > 0 {
			lines = append(lines, "*Accepted when:*")
		}
		w.p("%s\n\n", strings.Join(lines, "\\\n"))
		for _, c := range d.AcceptanceCriteria {
			w.p("- %s\n", oneLine(c))
		}
		if len(d.AcceptanceCriteria) > 0 {
			w.p("\n")
		}
	}
}

// deliverableList writes the deliverables as a short table: ID, title, and
// due date.
func (w *writer) deliverableList() {
	w.p("| ID | Deliverable | Due |\n|----|----------------------------------------|--------------------|\n")
	for _, d := range w.a.Deliverables {
		w.p("| %s | %s | %s |\n", d.ID, cell(d.Title), cell(w.date(d.Due)))
	}
	w.p("\n")
}

func (w *writer) milestones(ms []engagement.Milestone) {
	kinds := false
	for _, m := range ms {
		if m.Kind != "" {
			kinds = true
			break
		}
	}
	if kinds {
		w.p("| ID | Milestone | Kind | Date | Decision or review |\n|----|----------------|------------|--------------------|----------------------------------|\n")
	} else {
		w.p("| ID | Milestone | Date | Decision or review |\n|----|----------------|--------------------|----------------------------------|\n")
	}
	for _, m := range ms {
		title := m.Title
		if m.Description != "" {
			title += ": " + m.Description
		}
		if kinds {
			w.p("| %s | %s | %s | %s | %s |\n", m.ID, cell(title), cell(m.Kind.Label()), cell(w.date(m.Date)), cell(m.Decision))
			continue
		}
		w.p("| %s | %s | %s | %s |\n", m.ID, cell(title), cell(w.date(m.Date)), cell(m.Decision))
	}
	w.p("\n")
}

// responsibilities writes each party's commitments under the party's name.
// With a role, only parties with that role are listed, and a single party
// needs no heading of its own.
func (w *writer) responsibilities(role string) {
	a := w.a
	if role == "" {
		w.para("Each party commits to the following so that the work can succeed.")
	}
	var parties []engagement.Party
	for _, party := range a.Parties {
		if role == "" || string(party.Role) == role {
			parties = append(parties, party)
		}
	}
	for _, party := range parties {
		var items []engagement.Responsibility
		for _, r := range a.Responsibilities {
			if r.Party == party.ID {
				items = append(items, r)
			}
		}
		if len(items) == 0 {
			continue
		}
		if role == "" || len(parties) > 1 {
			w.h3(party.Name)
		}
		for _, r := range items {
			line := oneLine(r.Description)
			if r.Due != "" {
				line = strings.TrimRight(line, ".") + " (by " + w.date(r.Due) + ")."
			}
			w.p("- %s\n", line)
		}
		w.p("\n")
	}
}

func (w *writer) team() {
	a := w.a
	w.p("| Name | Organization | Role |\n|------------------|------------------|----------------------------------|\n")
	for _, party := range a.Parties {
		for _, p := range a.People {
			if p.Party != party.ID {
				continue
			}
			role := p.Role
			if p.Title != "" {
				if role != "" {
					role = p.Title + "; " + role
				} else {
					role = p.Title
				}
			}
			w.p("| %s | %s | %s |\n", cell(p.Name), cell(party.Name), cell(role))
		}
	}
	w.p("\n")
}

func (w *writer) governance() {
	g := w.a.Governance
	w.para(g.Summary)
	if len(g.Meetings) == 0 {
		return
	}
	w.p("| Meeting | Participants | Frequency | Length |\n|------------------|----------------------------|------------------|----------|\n")
	for _, m := range g.Meetings {
		w.p("| %s | %s | %s | %s |\n", cell(m.Title), cell(m.Participants), cell(m.Frequency), cell(m.Duration))
	}
	w.p("\n")
}

func (w *writer) contacts(cs []engagement.Contact) {
	a := w.a
	w.p("| Organization | Contact | Role |\n|------------------|------------------|----------------------------------|\n")
	for _, c := range cs {
		party, person := c.Party, c.Person
		if p, ok := a.Party(c.Party); ok {
			party = p.Name
		}
		if p, ok := a.Person(c.Person); ok {
			person = p.Name
		}
		w.p("| %s | %s | %s |\n", cell(party), cell(person), cell(c.Role))
	}
	w.p("\n")
}

func (w *writer) assumptions() {
	w.para("This plan assumes the following. If an assumption proves wrong, the parties will review its effect through the change process.")
	for _, x := range w.a.Assumptions {
		w.p("- %s\n", oneLine(x.Description))
	}
	w.p("\n")
}

func (w *writer) risks() {
	a := w.a
	w.p("| Risk | Mitigation |\n|--------------------------------|----------------------------------------|\n")
	for _, r := range a.Risks {
		m := r.Mitigation
		if r.Owner != "" {
			if p, ok := a.Party(r.Owner); ok {
				m = strings.TrimSpace(m) + " (" + p.Name + ")"
			}
		}
		w.p("| %s | %s |\n", cell(r.Description), cell(m))
	}
	w.p("\n")
}

func (w *writer) successCriteria() {
	for _, c := range w.a.SuccessCriteria {
		w.p("- %s\n", oneLine(c.Description))
	}
	w.p("\n")
}

func (w *writer) acceptance() {
	acc := w.a.Acceptance
	w.para(acc.Process)
	if acc.ReviewDays > 0 {
		w.para(fmt.Sprintf("The client reviews each deliverable within %d business days of receiving it.", acc.ReviewDays))
	}
}

func (w *writer) closeout() {
	a := w.a
	c := a.Closeout
	w.para(c.Summary)
	partyName := func(id string) string {
		if p, ok := a.Party(id); ok {
			return p.Name
		}
		return id
	}
	by := func(due string) string {
		if due == "" {
			return ""
		}
		return ", by " + w.date(due)
	}
	if len(c.Evaluations) > 0 {
		w.para("**Evaluations**")
		for _, e := range c.Evaluations {
			line := fmt.Sprintf("**%s.** %s evaluates %s", e.ID, partyName(e.Party), strings.TrimRight(oneLine(e.Subject), "."))
			if e.Method != "" {
				line += " (" + oneLine(e.Method) + ")"
			}
			w.p("- %s%s.\n", line, by(e.Due))
		}
		w.p("\n")
	}
	if len(c.Reports) > 0 {
		w.para("**Reports**")
		for _, r := range c.Reports {
			w.p("- **%s.** %s To %s%s.\n", r.ID, strings.TrimRight(oneLine(r.Description), ".")+".", partyName(r.To), by(r.Due))
		}
		w.p("\n")
	}
}

func (w *writer) commercials() {
	c := w.a.Commercials
	w.para("**Fee model:** " + c.Model.Label() + ".")
	w.para(c.Summary)
	if len(c.Fees) > 0 {
		w.p("| Item | Due | Amount (%s) |\n|----------------------------------|----------------|------------:|\n", c.Currency)
		var total int64
		totalOK := true
		for _, f := range c.Fees {
			w.p("| %s | %s | %s |\n", cell(f.Description), cell(f.Due), formatAmount(f.Amount))
			cents, ok := parseCents(f.Amount)
			if !ok {
				totalOK = false
			}
			total += cents
		}
		if totalOK && len(c.Fees) > 1 {
			w.p("| **Total** | | **%s** |\n", formatCents(total))
		}
		w.p("\n")
	}
	if strings.TrimSpace(c.Expenses) != "" {
		w.para("**Expenses.** " + strings.TrimSpace(c.Expenses))
	}
	if strings.TrimSpace(c.PaymentTerms) != "" {
		w.para("**Payment terms.** " + strings.TrimSpace(c.PaymentTerms))
	}
}

func (w *writer) terms() {
	for _, t := range w.a.Terms {
		w.h3(t.Title)
		w.para(t.Text)
	}
}

func (w *writer) openQuestions() {
	a := w.a
	w.para("To settle before this agreement is signed.")
	w.p("| ID | Question | Owner | By |\n|----|----------------------------------------|----------------|------------|\n")
	for _, q := range a.OpenQuestions {
		owner := q.Owner
		if p, ok := a.Party(q.Owner); ok {
			owner = p.Name
		} else if p, ok := a.Person(q.Owner); ok {
			owner = p.Name
		}
		w.p("| %s | %s | %s | %s |\n", q.ID, cell(q.Question), cell(owner), cell(w.date(q.Due)))
	}
	w.p("\n")
}

const signatureLine = `\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_`

const dateLine = `\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_`

func (w *writer) signatures(sigs []engagement.Signature) {
	a := w.a
	intro := fmt.Sprintf("By signing, the parties agree to this %s, version %s.", strings.ToLower(a.Kind.Label()), a.Version)
	if a.Kind == engagement.KindStatementOfWork {
		intro = fmt.Sprintf("By signing, the parties agree to this Statement of Work, version %s.", a.Version)
	}
	w.para(intro)
	if w.digest != "" {
		w.para("Content digest of the agreed text:\\\n`" + w.digest + "`")
	}
	for _, s := range sigs {
		partyName := s.Party
		if p, ok := a.Party(s.Party); ok {
			partyName = p.Name
		}
		person, _ := a.Person(s.Signer) // lint reports an unknown signer; render what is known
		name := person.Name
		if name == "" {
			name = s.Signer
		}
		var lines []string
		if s.SignedDate != "" {
			method := ""
			if s.Method != "" {
				method = " (" + strings.ReplaceAll(string(s.Method), "_", " ") + ")"
			}
			lines = append(lines, "Signed on "+formatDate(s.SignedDate)+method)
		} else {
			lines = append(lines, "Signature: "+signatureLine)
		}
		lines = append(lines, "Name: "+name)
		if title := cmp.Or(s.Capacity, person.Title, person.Role); title != "" {
			lines = append(lines, "Title: "+title)
		}
		if s.SignedDate == "" {
			lines = append(lines, "Date: "+dateLine)
		}
		// In PDF output each block stays on one page, with room above the
		// signature line to sign.
		w.raw(`\vspace{1em}\noindent\begin{minipage}{\linewidth}`)
		w.p("**For %s**\n\n", partyName)
		if s.SignedDate == "" {
			w.raw(`\vspace{2.5em}`)
			w.rawDOCX(`<w:p/><w:p/>`)
		}
		w.p("%s\n\n", strings.Join(lines, "\\\n"))
		w.raw(`\end{minipage}`)
	}
}

// date formats a date at the outline's precision.
func (w *writer) date(s string) string {
	if w.outline.DatePrecision == engagement.PrecisionMonth {
		return formatMonth(s)
	}
	return formatDate(s)
}

// dateRange formats a range at the outline's precision.
func (w *writer) dateRange(start, end string) string {
	if w.outline.DatePrecision == engagement.PrecisionMonth {
		return formatMonthRange(start, end)
	}
	return formatRange(start, end)
}

// raw writes a raw LaTeX block, which only PDF output uses.
func (w *writer) raw(tex string) {
	if w.opts.Flavor != FlavorPandoc {
		return
	}
	w.p("```{=latex}\n%s\n```\n\n", tex)
}

// rawDOCX writes a raw OpenXML block, which only DOCX output uses.
func (w *writer) rawDOCX(xml string) {
	if w.opts.Flavor != FlavorPandoc {
		return
	}
	w.p("```{=openxml}\n%s\n```\n\n", xml)
}

// labeled formats "**ID. Title.** Description".
func labeled(id, title, desc string) string {
	s := fmt.Sprintf("**%s. %s.**", id, strings.TrimRight(strings.TrimSpace(title), "."))
	if d := oneLine(desc); d != "" {
		s += " " + d
	}
	return s
}

// oneLine collapses a Markdown fragment to one line so it fits in a list
// item or table cell.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// cell makes s safe for a pipe-table cell.
func cell(s string) string {
	return strings.ReplaceAll(oneLine(s), "|", `\|`)
}

func formatDate(s string) string {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return s
	}
	return t.Format("January 2, 2006")
}

func formatRange(start, end string) string {
	switch {
	case start != "" && end != "":
		s, errS := time.Parse(time.DateOnly, start)
		e, errE := time.Parse(time.DateOnly, end)
		if errS == nil && errE == nil && s.Year() == e.Year() {
			return s.Format("January 2") + " – " + e.Format("January 2, 2006")
		}
		return formatDate(start) + " – " + formatDate(end)
	case start != "":
		return "from " + formatDate(start)
	case end != "":
		return "until " + formatDate(end)
	default:
		return ""
	}
}

func formatMonth(s string) string {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return s
	}
	return t.Format("January 2006")
}

func formatMonthRange(start, end string) string {
	s, errS := time.Parse(time.DateOnly, start)
	e, errE := time.Parse(time.DateOnly, end)
	switch {
	case errS == nil && errE == nil && s.Year() == e.Year() && s.Month() == e.Month():
		return s.Format("January 2006")
	case errS == nil && errE == nil && s.Year() == e.Year():
		return s.Format("January") + " – " + e.Format("January 2006")
	case errS == nil && errE == nil:
		return s.Format("January 2006") + " – " + e.Format("January 2006")
	default:
		return formatRange(start, end)
	}
}

// texEscape escapes the characters TeX treats specially.
func texEscape(s string) string {
	r := strings.NewReplacer(`\`, `\textbackslash{}`, "&", `\&`, "%", `\%`, "$", `\$`, "#", `\#`, "_", `\_`, "{", `\{`, "}", `\}`, "~", `\textasciitilde{}`, "^", `\textasciicircum{}`)
	return r.Replace(s)
}

func shortDigest(d string) string {
	const n = len(engagementio.DigestPrefix) + 16
	if len(d) <= n {
		return d
	}
	return d[:n]
}

// yamlString quotes s as a YAML double-quoted scalar.
func yamlString(s string) string {
	return strconv.Quote(s)
}

// parseCents parses a decimal amount ("12500" or "12500.50") into cents.
func parseCents(s string) (int64, bool) {
	whole, frac, hasFrac := strings.Cut(s, ".")
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, false
	}
	var f int64
	if hasFrac {
		if len(frac) != 2 {
			return 0, false
		}
		f, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, false
		}
	}
	return w*100 + f, true
}

func formatCents(c int64) string {
	return groupThousands(strconv.FormatInt(c/100, 10)) + fmt.Sprintf(".%02d", c%100)
}

func formatAmount(s string) string {
	if c, ok := parseCents(s); ok {
		return formatCents(c)
	}
	return cell(s)
}

func groupThousands(digits string) string {
	if len(digits) <= 3 {
		return digits
	}
	var b strings.Builder
	pre := len(digits) % 3
	if pre > 0 {
		b.WriteString(digits[:pre])
	}
	for i := pre; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}

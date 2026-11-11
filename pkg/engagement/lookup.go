package engagement

// Party returns the party with the given ID.
func (a *Agreement) Party(id string) (Party, bool) {
	for _, p := range a.Parties {
		if p.ID == id {
			return p, true
		}
	}
	return Party{}, false
}

// Person returns the person with the given ID.
func (a *Agreement) Person(id string) (Person, bool) {
	for _, p := range a.People {
		if p.ID == id {
			return p, true
		}
	}
	return Person{}, false
}

// Phase returns the phase with the given ID.
func (a *Agreement) Phase(id string) (Phase, bool) {
	for _, p := range a.Phases {
		if p.ID == id {
			return p, true
		}
	}
	return Phase{}, false
}

// PartiesByRole returns the parties with the given role, in document order.
func (a *Agreement) PartiesByRole(role PartyRole) []Party {
	var out []Party
	for _, p := range a.Parties {
		if p.Role == role {
			out = append(out, p)
		}
	}
	return out
}

// Consultant returns the first consultant party.
func (a *Agreement) Consultant() (Party, bool) {
	ps := a.PartiesByRole(RoleConsultant)
	if len(ps) == 0 {
		return Party{}, false
	}
	return ps[0], true
}

// Client returns the first client party.
func (a *Agreement) Client() (Party, bool) {
	ps := a.PartiesByRole(RoleClient)
	if len(ps) == 0 {
		return Party{}, false
	}
	return ps[0], true
}

// DeliverableOwner returns the ID of the party that owns d: its Owner, or
// the consultant when Owner is empty.
func (a *Agreement) DeliverableOwner(d Deliverable) string {
	if d.Owner != "" {
		return d.Owner
	}
	if c, ok := a.Consultant(); ok {
		return c.ID
	}
	return ""
}

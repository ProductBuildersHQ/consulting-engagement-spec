package engagement

import "testing"

func TestEnums(t *testing.T) {
	for _, k := range AllKinds() {
		if !k.Valid() || k.Label() == string(k) {
			t.Errorf("kind %q: invalid or unlabeled", k)
		}
	}
	for _, s := range AllStatuses() {
		if !s.Valid() || s.Label() == string(s) {
			t.Errorf("status %q: invalid or unlabeled", s)
		}
	}
	for _, m := range AllCommercialModels() {
		if !m.Valid() || m.Label() == string(m) {
			t.Errorf("model %q: invalid or unlabeled", m)
		}
	}
	for _, r := range AllPartyRoles() {
		if !r.Valid() {
			t.Errorf("role %q invalid", r)
		}
	}
	for _, m := range AllSignatureMethods() {
		if !m.Valid() {
			t.Errorf("method %q invalid", m)
		}
	}
	if Kind("memo").Valid() || Status("done").Valid() {
		t.Error("unknown values reported valid")
	}
}

func TestLookups(t *testing.T) {
	a := &Agreement{
		Parties: []Party{{ID: "c", Role: RoleConsultant, Name: "C"}, {ID: "k", Role: RoleClient, Name: "K"}},
		People:  []Person{{ID: "p", Name: "P", Party: "k"}},
	}
	if c, ok := a.Consultant(); !ok || c.ID != "c" {
		t.Fatal("Consultant")
	}
	if k, ok := a.Client(); !ok || k.ID != "k" {
		t.Fatal("Client")
	}
	if _, ok := a.Person("p"); !ok {
		t.Fatal("Person")
	}
	if got := a.DeliverableOwner(Deliverable{}); got != "c" {
		t.Fatalf("DeliverableOwner = %q", got)
	}
	if got := a.DeliverableOwner(Deliverable{Owner: "k"}); got != "k" {
		t.Fatalf("DeliverableOwner = %q", got)
	}
}

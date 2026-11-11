package engagement

import "slices"

// Kind is the kind of agreement document.
type Kind string

// Agreement kinds.
const (
	// KindProposal is offered by the consultant for the client to accept.
	KindProposal Kind = "proposal"
	// KindStatementOfWork is the agreed scope, deliverables, and
	// responsibilities for a piece of work.
	KindStatementOfWork Kind = "statement_of_work"
)

// AllKinds returns every Kind in display order.
func AllKinds() []Kind { return []Kind{KindProposal, KindStatementOfWork} }

// Valid reports whether k is a known Kind.
func (k Kind) Valid() bool { return slices.Contains(AllKinds(), k) }

// Label returns the display name of k.
func (k Kind) Label() string {
	switch k {
	case KindProposal:
		return "Proposal"
	case KindStatementOfWork:
		return "Statement of Work"
	default:
		return string(k)
	}
}

// Status is where an agreement version is in its lifecycle.
type Status string

// Agreement statuses.
const (
	StatusDraft      Status = "draft"
	StatusInReview   Status = "in_review"
	StatusAgreed     Status = "agreed"
	StatusSuperseded Status = "superseded"
	StatusWithdrawn  Status = "withdrawn"
)

// AllStatuses returns every Status in lifecycle order.
func AllStatuses() []Status {
	return []Status{StatusDraft, StatusInReview, StatusAgreed, StatusSuperseded, StatusWithdrawn}
}

// Valid reports whether s is a known Status.
func (s Status) Valid() bool { return slices.Contains(AllStatuses(), s) }

// Label returns the display name of s.
func (s Status) Label() string {
	switch s {
	case StatusDraft:
		return "Draft"
	case StatusInReview:
		return "In review"
	case StatusAgreed:
		return "Agreed"
	case StatusSuperseded:
		return "Superseded"
	case StatusWithdrawn:
		return "Withdrawn"
	default:
		return string(s)
	}
}

// PartyRole is a party's role in the agreement.
type PartyRole string

// Party roles.
const (
	RoleConsultant PartyRole = "consultant"
	RoleClient     PartyRole = "client"
	// RoleSponsor funds or sponsors the work without being the client.
	RoleSponsor PartyRole = "sponsor"
	// RolePartner contributes to the work alongside the consultant.
	RolePartner PartyRole = "partner"
)

// AllPartyRoles returns every PartyRole in display order.
func AllPartyRoles() []PartyRole {
	return []PartyRole{RoleConsultant, RoleClient, RoleSponsor, RolePartner}
}

// Valid reports whether r is a known PartyRole.
func (r PartyRole) Valid() bool { return slices.Contains(AllPartyRoles(), r) }

// CommercialModel is how the work is priced.
type CommercialModel string

// Commercial models.
const (
	ModelProBono          CommercialModel = "pro_bono"
	ModelFixedFee         CommercialModel = "fixed_fee"
	ModelTimeAndMaterials CommercialModel = "time_and_materials"
	ModelRetainer         CommercialModel = "retainer"
)

// AllCommercialModels returns every CommercialModel in display order.
func AllCommercialModels() []CommercialModel {
	return []CommercialModel{ModelProBono, ModelFixedFee, ModelTimeAndMaterials, ModelRetainer}
}

// Valid reports whether m is a known CommercialModel.
func (m CommercialModel) Valid() bool { return slices.Contains(AllCommercialModels(), m) }

// Label returns the display name of m.
func (m CommercialModel) Label() string {
	switch m {
	case ModelProBono:
		return "Pro bono"
	case ModelFixedFee:
		return "Fixed fee"
	case ModelTimeAndMaterials:
		return "Time and materials"
	case ModelRetainer:
		return "Retainer"
	default:
		return string(m)
	}
}

// SignatureMethod is how a signature was applied.
type SignatureMethod string

// Signature methods.
const (
	// MethodWetInk is a handwritten signature on paper or a scan of one.
	MethodWetInk SignatureMethod = "wet_ink"
	// MethodElectronic is an electronic signature, such as a typed or
	// drawn signature through an e-signature service.
	MethodElectronic SignatureMethod = "electronic"
	// MethodDigital is a cryptographic signature on the PDF, such as PAdES.
	MethodDigital SignatureMethod = "digital"
)

// AllSignatureMethods returns every SignatureMethod in display order.
func AllSignatureMethods() []SignatureMethod {
	return []SignatureMethod{MethodWetInk, MethodElectronic, MethodDigital}
}

// Valid reports whether m is a known SignatureMethod.
func (m SignatureMethod) Valid() bool { return slices.Contains(AllSignatureMethods(), m) }

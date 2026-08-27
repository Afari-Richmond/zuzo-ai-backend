// Package data holds the server-side partner-scoped dashboard data used to build
// ZIA's context. It intentionally mirrors client/src/data/partner-dashboard-mock.ts —
// the two mock stores need to move together until this is backed by a real database.
package data

type Profile struct {
	ID            string `json:"id"`
	CompanyName   string `json:"companyName"`
	ContactName   string `json:"contactName"`
	Role          string `json:"role"`
	Track         string `json:"track"`
	Tier          string `json:"tier"`
	MarginPercent int    `json:"marginPercent"`
	PartnerSince  string `json:"partnerSince"`
}

type Kpi struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
}

type ActivityItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

type Engagement struct {
	Name           string `json:"name"`
	Track          string `json:"track"`
	Status         string `json:"status"`
	Description    string `json:"description"`
	AgentsAssigned int    `json:"agentsAssigned"`
	SLA            string `json:"sla"`
	CSAT           string `json:"csat"`
}

type TicketCounts struct {
	Open       int `json:"open"`
	InProgress int `json:"in_progress"`
	Resolved   int `json:"resolved"`
	Escalated  int `json:"escalated"`
}

type Ticket struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Agent       string `json:"agent"`
	UpdatedAt   string `json:"updatedAt"`
	Description string `json:"description"`
}

type ContractDocument struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	SignedDate string `json:"signedDate"`
}

type Invoice struct {
	Period     string `json:"period"`
	Amount     string `json:"amount"`
	Status     string `json:"status"`
	IssuedDate string `json:"issuedDate"`
}

type SuccessManager struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type TeamMember struct {
	Name               string `json:"name"`
	Role               string `json:"role"`
	Shift              string `json:"shift"`
	TicketsResolvedMtd int    `json:"ticketsResolvedMtd"`
	CSAT               string `json:"csat"`
}

type Bundle struct {
	Profile           Profile            `json:"profile"`
	Kpis              []Kpi              `json:"kpis"`
	Activity          []ActivityItem     `json:"activity"`
	Engagements       []Engagement       `json:"engagements"`
	TicketCounts      TicketCounts       `json:"ticketCounts"`
	Tickets           []Ticket           `json:"tickets"`
	ContractDocuments []ContractDocument `json:"contractDocuments"`
	Invoices          []Invoice          `json:"invoices"`
	SuccessManager    SuccessManager     `json:"successManager"`
	TeamRoster        []TeamMember       `json:"teamRoster"`
}

var bundles = map[string]Bundle{
	// The one real onboarded partner (see partnerBundles.partner_vcall in
	// partner-dashboard-mock.ts) — still pre-launch, so KPIs/tickets/
	// engagements are empty rather than fabricated.
	"partner_vcall": {
		Profile: Profile{
			ID:            "partner_vcall",
			CompanyName:   "VCall Solutions",
			ContactName:   "Samira Dean",
			Role:          "Founder",
			Track:         "Legal Services",
			Tier:          "Starter",
			MarginPercent: 0,
			PartnerSince:  "2026-08-21",
		},
		Kpis: []Kpi{
			{Label: "Active Agents", Value: "0", Delta: "Pre-launch"},
			{Label: "Tickets Resolved (MTD)", Value: "0", Delta: "Pre-launch"},
			{Label: "CSAT Score", Value: "N/A", Delta: "Pre-launch"},
			{Label: "Margin Earned (MTD)", Value: "$0", Delta: "Pre-launch"},
		},
		Activity: []ActivityItem{
			{Title: "Portal access granted", Description: "Welcome to ZuZo — your shared document library and messaging are ready to use.", Timestamp: "2026-08-21T09:00:00Z"},
		},
		Engagements:       []Engagement{},
		TicketCounts:      TicketCounts{Open: 0, InProgress: 0, Resolved: 0, Escalated: 0},
		Tickets:           []Ticket{},
		ContractDocuments: []ContractDocument{},
		Invoices:          []Invoice{},
		SuccessManager:    SuccessManager{Name: "Gilbert Debrah", Email: "gilbert.debrah@zuzogp.com", Phone: "+233 30 123 4567"},
		TeamRoster:        []TeamMember{},
	},
}

// GetBundle looks up a partner's scoped dashboard data. The caller (the AI
// service) never accepts a raw data blob from the client — only a partnerId —
// so a request can never inject another partner's data into its own context.
func GetBundle(partnerID string) (Bundle, bool) {
	b, ok := bundles[partnerID]
	return b, ok
}

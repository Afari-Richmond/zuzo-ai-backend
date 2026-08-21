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
	"partner_001": {
		Profile: Profile{
			ID:            "partner_001",
			CompanyName:   "Northwind IT Solutions",
			ContactName:   "Richmond Addo Afari",
			Role:          "Operations Director",
			Track:         "MSP",
			Tier:          "Growth",
			MarginPercent: 45,
			PartnerSince:  "2025-11-03",
		},
		Kpis: []Kpi{
			{Label: "Active Agents", Value: "18", Delta: "+3 this month"},
			{Label: "Tickets Resolved (MTD)", Value: "4,286", Delta: "+12% vs last month"},
			{Label: "CSAT Score", Value: "4.8 / 5", Delta: "+0.2 vs last month"},
			{Label: "Margin Earned (MTD)", Value: "$38,940", Delta: "+8% vs last month"},
		},
		Activity: []ActivityItem{
			{Title: "Weekly optimization call completed", Description: "Reviewed Q3 SLA performance with your Success Manager, Priya Nair.", Timestamp: "2026-08-13T15:30:00Z"},
			{Title: "Contract amendment signed", Description: "Tier upgrade to Growth (45% margin) took effect.", Timestamp: "2026-08-10T09:00:00Z"},
			{Title: "3 new agents onboarded to Tier 1 Helpdesk pod", Description: "Ramp-up complete, all agents certified on ConnectWise PSA.", Timestamp: "2026-08-07T12:00:00Z"},
			{Title: "August invoice generated", Description: "Statement #INV-2026-0847 is ready for review.", Timestamp: "2026-08-01T08:00:00Z"},
			{Title: "Escalation resolved: TCK-88213", Description: "Critical outage ticket closed within 42 minutes of the 1-hour SLA.", Timestamp: "2026-07-29T21:10:00Z"},
		},
		Engagements: []Engagement{
			{Name: "Tier 1 Helpdesk Support", Track: "MSP", Status: "active", Description: "Round-the-clock password resets, ticket triage, and endpoint troubleshooting.", AgentsAssigned: 12, SLA: "98.6% within SLA", CSAT: "4.9 / 5"},
			{Name: "After-Hours Escalation Pod", Track: "MSP", Status: "active", Description: "Dedicated overnight pod handling critical escalations.", AgentsAssigned: 4, SLA: "100% within SLA", CSAT: "4.7 / 5"},
			{Name: "WISMO / Order Support Pilot", Track: "3PL / E-commerce", Status: "onboarding", Description: "Resolving \"Where is my order?\" tickets inside Gorgias.", AgentsAssigned: 2, SLA: "Pending baseline", CSAT: "N/A"},
		},
		TicketCounts: TicketCounts{Open: 27, InProgress: 41, Resolved: 4286, Escalated: 3},
		Tickets: []Ticket{
			{ID: "TCK-88231", Subject: "VPN client failing to authenticate for remote users", Status: "in_progress", Priority: "high", Agent: "Kwame Asante", UpdatedAt: "2026-08-14T11:12:00Z", Description: "Multiple remote employees are reporting authentication failures via the corporate VPN client, tied to a recent certificate rotation."},
			{ID: "TCK-88229", Subject: "Password reset requests backlog", Status: "open", Priority: "medium", Agent: "Ama Boateng", UpdatedAt: "2026-08-14T10:40:00Z", Description: "A queue of 12 self-service password reset requests failed to auto-provision overnight due to an Azure AD sync delay."},
			{ID: "TCK-88225", Subject: "Printer driver rollout failing on 3 endpoints", Status: "open", Priority: "low", Agent: "Kwame Asante", UpdatedAt: "2026-08-14T09:05:00Z", Description: "The new driver package pushed via NinjaOne is failing silently on 3 Windows 11 endpoints."},
			{ID: "TCK-88219", Subject: "Server outage - client HQ file share unreachable", Status: "escalated", Priority: "critical", Agent: "Priya Nair", UpdatedAt: "2026-08-13T22:51:00Z", Description: "Primary file server at client HQ is unreachable, escalated per the 1-hour critical SLA."},
			{ID: "TCK-88213", Subject: "Critical outage - core switch down", Status: "resolved", Priority: "critical", Agent: "Ama Boateng", UpdatedAt: "2026-07-29T21:10:00Z", Description: "Core network switch failed; failover activated, resolved within 42 minutes of the 1-hour SLA."},
			{ID: "TCK-88208", Subject: "Shared mailbox permissions not syncing", Status: "resolved", Priority: "low", Agent: "Kofi Mensah", UpdatedAt: "2026-07-27T14:22:00Z", Description: "Forced a manual Exchange Online permissions sync to resolve delegate access."},
			{ID: "TCK-88202", Subject: "Onboarding laptop imaging queue stuck", Status: "in_progress", Priority: "medium", Agent: "Kofi Mensah", UpdatedAt: "2026-07-25T16:05:00Z", Description: "Three new-hire laptops are stuck mid-image; investigating a driver conflict."},
			{ID: "TCK-88196", Subject: "Guest Wi-Fi captive portal not loading", Status: "open", Priority: "low", Agent: "Efua Owusu", UpdatedAt: "2026-07-24T19:48:00Z", Description: "Visitors on the guest Wi-Fi are not redirected to the captive portal on some devices."},
			{ID: "TCK-88190", Subject: "Recurring meeting invites duplicating on sync", Status: "resolved", Priority: "medium", Agent: "Efua Owusu", UpdatedAt: "2026-07-22T12:15:00Z", Description: "Cleared the local calendar cache and re-synced affected mailboxes."},
		},
		ContractDocuments: []ContractDocument{
			{Name: "Master Service Agreement", Type: "MSA", SignedDate: "2025-11-03"},
			{Name: "Statement of Work - Tier 1 Helpdesk", Type: "SOW", SignedDate: "2025-11-13"},
			{Name: "Tier Upgrade Amendment - Growth (45%)", Type: "Amendment", SignedDate: "2026-08-10"},
			{Name: "Mutual NDA", Type: "NDA", SignedDate: "2025-10-28"},
		},
		Invoices: []Invoice{
			{Period: "August 2026", Amount: "$38,940.00", Status: "pending", IssuedDate: "2026-08-01"},
			{Period: "July 2026", Amount: "$36,110.00", Status: "paid", IssuedDate: "2026-07-01"},
			{Period: "June 2026", Amount: "$34,225.00", Status: "paid", IssuedDate: "2026-06-01"},
			{Period: "May 2026", Amount: "$31,860.00", Status: "paid", IssuedDate: "2026-05-01"},
		},
		SuccessManager: SuccessManager{Name: "Priya Nair", Email: "priya.nair@zuzogp.com", Phone: "+233 30 123 4567"},
		TeamRoster: []TeamMember{
			{Name: "Kwame Asante", Role: "Tier 1 Support Lead", Shift: "Day Shift (8am-4pm GMT)", TicketsResolvedMtd: 412, CSAT: "4.9 / 5"},
			{Name: "Ama Boateng", Role: "Senior Support Agent", Shift: "Night Shift (10pm-6am GMT)", TicketsResolvedMtd: 389, CSAT: "4.8 / 5"},
			{Name: "Kofi Mensah", Role: "Support Agent", Shift: "Day Shift (8am-4pm GMT)", TicketsResolvedMtd: 276, CSAT: "4.7 / 5"},
			{Name: "Efua Owusu", Role: "Support Agent", Shift: "Swing Shift (4pm-12am GMT)", TicketsResolvedMtd: 298, CSAT: "4.8 / 5"},
		},
	},
	"partner_002": {
		Profile: Profile{
			ID:            "partner_002",
			CompanyName:   "Solace Fulfillment Group",
			ContactName:   "Jennifer Osei",
			Role:          "VP of Operations",
			Track:         "3PL / E-commerce",
			Tier:          "Enterprise",
			MarginPercent: 55,
			PartnerSince:  "2025-06-16",
		},
		Kpis: []Kpi{
			{Label: "Active Agents", Value: "32", Delta: "+5 this month"},
			{Label: "Tickets Resolved (MTD)", Value: "6,102", Delta: "+9% vs last month"},
			{Label: "CSAT Score", Value: "4.7 / 5", Delta: "+0.1 vs last month"},
			{Label: "Margin Earned (MTD)", Value: "$71,280", Delta: "+11% vs last month"},
		},
		Activity: []ActivityItem{
			{Title: "Peak season staffing plan approved", Description: "Reviewed Q4 ramp plan with your Success Manager, Daniel Owusu — adding 8 agents ahead of Black Friday.", Timestamp: "2026-08-15T14:10:00Z"},
			{Title: "Enterprise tier renewal signed", Description: "12-month renewal at 55% margin took effect.", Timestamp: "2026-08-09T10:00:00Z"},
			{Title: "Returns & Refunds Desk pod onboarded", Description: "6 agents certified on Loop Returns and NetSuite.", Timestamp: "2026-08-05T13:00:00Z"},
			{Title: "August invoice generated", Description: "Statement #INV-2026-1102 is ready for review.", Timestamp: "2026-08-01T08:00:00Z"},
			{Title: "Escalation resolved: TCK-91045", Description: "Shopify webhook outage causing order status delays fixed within 38 minutes.", Timestamp: "2026-07-30T19:44:00Z"},
		},
		Engagements: []Engagement{
			{Name: "WISMO / Order Support", Track: "3PL / E-commerce", Status: "active", Description: "Resolving \"Where is my order?\" tickets across Shopify storefronts, synced with ShipStation.", AgentsAssigned: 14, SLA: "99.1% within SLA", CSAT: "4.7 / 5"},
			{Name: "Returns & Refunds Desk", Track: "3PL / E-commerce", Status: "onboarding", Description: "Dedicated pod handling return authorizations and refund processing ahead of peak season.", AgentsAssigned: 6, SLA: "Pending baseline", CSAT: "N/A"},
		},
		TicketCounts: TicketCounts{Open: 44, InProgress: 58, Resolved: 6102, Escalated: 2},
		Tickets: []Ticket{
			{ID: "TCK-91052", Subject: "Warehouse mis-pick causing wrong-item shipments", Status: "in_progress", Priority: "high", Agent: "Nana Yeboah", UpdatedAt: "2026-08-14T13:40:00Z", Description: "A batch of East Coast fulfillment center orders shipped with an incorrect SKU substitution; issuing proactive replacements."},
			{ID: "TCK-91049", Subject: "Refund requests stuck in Loop Returns queue", Status: "open", Priority: "medium", Agent: "Abena Frimpong", UpdatedAt: "2026-08-14T11:15:00Z", Description: "A sync delay between Loop Returns and NetSuite is holding up 9 approved refunds; processing manually."},
			{ID: "TCK-91045", Subject: "Shopify webhook outage delaying order status updates", Status: "resolved", Priority: "critical", Agent: "Daniel Owusu", UpdatedAt: "2026-07-30T19:44:00Z", Description: "A Shopify platform-wide webhook outage delayed order status syncs ~40 minutes; backlog cleared within 38 minutes of detection."},
			{ID: "TCK-91038", Subject: "Chargeback dispute needs shipment proof", Status: "escalated", Priority: "high", Agent: "Nana Yeboah", UpdatedAt: "2026-07-29T16:05:00Z", Description: "A customer filed a chargeback despite confirmed signature on file; escalated with the ShipStation proof-of-delivery packet."},
			{ID: "TCK-91031", Subject: "Guest checkout tracking link returns 404", Status: "open", Priority: "low", Agent: "Abena Frimpong", UpdatedAt: "2026-07-28T09:30:00Z", Description: "Guest checkout tracking links 404 intermittently, traced to a caching issue after a theme update."},
			{ID: "TCK-91020", Subject: "Peak season staffing request approved", Status: "resolved", Priority: "medium", Agent: "Daniel Owusu", UpdatedAt: "2026-07-24T12:00:00Z", Description: "Approved request to add 8 agents to the WISMO pod ahead of Black Friday / Cyber Monday volume."},
		},
		ContractDocuments: []ContractDocument{
			{Name: "Master Service Agreement", Type: "MSA", SignedDate: "2025-06-16"},
			{Name: "Statement of Work - WISMO / Order Support", Type: "SOW", SignedDate: "2025-06-23"},
			{Name: "Enterprise Tier Renewal (55%)", Type: "Amendment", SignedDate: "2026-08-09"},
			{Name: "Mutual NDA", Type: "NDA", SignedDate: "2025-06-10"},
		},
		Invoices: []Invoice{
			{Period: "August 2026", Amount: "$71,280.00", Status: "pending", IssuedDate: "2026-08-01"},
			{Period: "July 2026", Amount: "$64,150.00", Status: "paid", IssuedDate: "2026-07-01"},
			{Period: "June 2026", Amount: "$61,900.00", Status: "paid", IssuedDate: "2026-06-01"},
			{Period: "May 2026", Amount: "$58,420.00", Status: "paid", IssuedDate: "2026-05-01"},
		},
		SuccessManager: SuccessManager{Name: "Daniel Owusu", Email: "daniel.owusu@zuzogp.com", Phone: "+233 30 987 6543"},
		TeamRoster: []TeamMember{
			{Name: "Nana Yeboah", Role: "WISMO Specialist Lead", Shift: "Day Shift (8am-4pm GMT)", TicketsResolvedMtd: 588, CSAT: "4.8 / 5"},
			{Name: "Abena Frimpong", Role: "Returns Desk Lead", Shift: "Swing Shift (4pm-12am GMT)", TicketsResolvedMtd: 214, CSAT: "4.6 / 5"},
			{Name: "Kojo Antwi", Role: "Order Support Agent", Shift: "Day Shift (8am-4pm GMT)", TicketsResolvedMtd: 401, CSAT: "4.7 / 5"},
			{Name: "Yaa Asantewaa", Role: "Order Support Agent", Shift: "Night Shift (10pm-6am GMT)", TicketsResolvedMtd: 376, CSAT: "4.7 / 5"},
		},
	},
}

// GetBundle looks up a partner's scoped dashboard data. The caller (the AI
// service) never accepts a raw data blob from the client — only a partnerId —
// so a request can never inject another partner's data into its own context.
func GetBundle(partnerID string) (Bundle, bool) {
	b, ok := bundles[partnerID]
	return b, ok
}

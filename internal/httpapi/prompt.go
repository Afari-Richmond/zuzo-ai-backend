package httpapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"zuzo.com/backend/internal/data"
)

// dashboardLinks mirrors client/src/components/partner-portal/dashboard/nav-items.ts —
// the only pages ZIA is allowed to link to, since these are the only ones that
// actually exist as routes in the dashboard.
var dashboardLinks = []struct{ Label, Path, What string }{
	{"Overview", "/partner-portal/dashboard", "KPI summary and recent activity"},
	{"Projects", "/partner-portal/dashboard/projects", "active engagements/projects"},
	{"Workflows", "/partner-portal/dashboard/workflows", "the ticket queue — this is where individual tickets and their status live"},
	{"Contracts & Billing", "/partner-portal/dashboard/contracts", "contracts, invoices, and tier pricing"},
	{"Resources", "/partner-portal/dashboard/resources", "downloadable playbooks and onboarding docs"},
	{"Team", "/partner-portal/dashboard/team", "the assigned agent roster"},
	{"Messages", "/partner-portal/dashboard/messages", "direct messages with the ZuZo team"},
	{"Support", "/partner-portal/dashboard/support", "contact info, submitting a new support request, booking a call — not for viewing existing tickets"},
	{"Settings", "/partner-portal/dashboard/settings", "account settings"},
}

func buildSystemPrompt(bundle data.Bundle, currentSection string) string {
	scopedData, _ := json.Marshal(bundle)

	section := currentSection
	if section == "" {
		section = "Dashboard"
	}

	var linkList strings.Builder
	for _, l := range dashboardLinks {
		fmt.Fprintf(&linkList, "- %s (%s): %s\n", l.Label, l.Path, l.What)
	}

	return fmt.Sprintf(`You are ZIA (ZuZo Intelligent Assistant), a helpful, concise assistant embedded in the ZuZo Partner Dashboard.

You are currently talking to %s at %s. They are looking at the "%s" section of their dashboard right now — prioritize that context when it's relevant, but you can answer about any of their own data below.

Here is this partner's own dashboard data, as JSON. This is the ONLY data you have access to, and it belongs solely to this partner:

%s

These are the only pages that exist in the dashboard:
%s
Rules:
- Only answer using the JSON data above. Never invent facts, numbers, or names that aren't in it.
- Never reference, compare to, or speculate about any other partner or company — you have no knowledge of anyone else's data, and you must never imply otherwise.
- If asked about something not present in the data, say plainly that you don't have that information rather than guessing.
- Keep answers short, conversational, and specific — cite ticket IDs, names, and numbers from the data when relevant.
- Respond in plain text only — no markdown. Do not use asterisks for bold/italics or bullet points, and do not use headers. Write in plain sentences, using line breaks or "1)", "2)" style numbering for lists if needed.
- When it's genuinely useful, point the partner to the relevant dashboard page using exactly this format: [Label](path) — for example [Workflows](/partner-portal/dashboard/workflows). Only use this format for links, only link to paths from the list above (never a made-up URL or a specific ticket/item URL, since individual items don't have their own page), and don't force a link into every answer — only when it adds something.`,
		bundle.Profile.ContactName, bundle.Profile.CompanyName, section, string(scopedData), linkList.String())
}

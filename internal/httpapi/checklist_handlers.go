package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
)

// ChecklistItem is one static, admin-facing task within a stage's
// checklist. Keys must stay identical (and stable — never renumbered once
// used) to the frontend's mirror list in use-onboarding-checklist.tsx,
// since completion rows in partner_onboarding_checklist are keyed by
// item_key.
type ChecklistItem struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type ChecklistSubStep struct {
	Title string          `json:"title"`
	Items []ChecklistItem `json:"items"`
}

// checklistCatalog maps each of the 4 onboarding stages to its sub-steps,
// derived from ZuZo's full BPO client-onboarding SOP (7 detailed
// components, condensed here into the 4 high-level stages the dropdown/
// badge already use). This is the single source of truth on the Go side;
// the frontend keeps an identical copy for rendering.
var checklistCatalog = map[string][]ChecklistSubStep{
	"Assessment & Setup": {
		{
			Title: "Preparation",
			Items: []ChecklistItem{
				{Key: "prep_preliminary_activities", Label: "Preliminary Activities", Description: "Internal readiness assessment, team assigned, preliminary timeline drafted."},
				{Key: "prep_nda", Label: "Sign NDA", Description: "Non-disclosure agreement reviewed and signed by both parties."},
				{Key: "prep_documents", Label: "Exchange Documents", Description: "Client documentation received; ZuZo methodology shared."},
				{Key: "prep_sow_contract", Label: "SOW & Contract Agreement", Description: "Statement of Work and service scope defined."},
				{Key: "prep_contract_review", Label: "Contract Review", Description: "Internal legal, finance, and ops review completed."},
				{Key: "prep_sign_contract", Label: "Sign Contract", Description: "Contract signed by authorized signatories."},
				{Key: "prep_po", Label: "Purchase Order", Description: "Purchase order received and validated against contract terms."},
				{Key: "prep_payment", Label: "One-Time Payment Fee", Description: "Implementation fee invoiced and payment received."},
			},
		},
		{
			Title: "Kick-off",
			Items: []ChecklistItem{
				{Key: "kickoff_meeting", Label: "Kick-off Meeting", Description: "Kick-off meeting held with client stakeholders."},
				{Key: "kickoff_signoff", Label: "Sign-Off of Activities", Description: "Implementation plan approved and action items documented."},
			},
		},
		{
			Title: "Site Assessment",
			Items: []ChecklistItem{
				{Key: "site_assessment", Label: "Site Inspection & Assessment", Description: "Facility, connectivity, and security requirements assessed."},
			},
		},
	},
	"Training & Integration": {
		{
			Title: "Solution Design",
			Items: []ChecklistItem{
				{Key: "design_solution", Label: "Solution Design", Description: "Technical architecture and integration points defined."},
				{Key: "design_process", Label: "Process Development", Description: "Current and future-state processes mapped and documented."},
				{Key: "design_planning", Label: "Implementation Planning", Description: "Timeline refined, resources allocated, dependencies identified."},
			},
		},
		{
			Title: "Site & System Setup",
			Items: []ChecklistItem{
				{Key: "setup_infrastructure", Label: "Infrastructure Setup", Description: "Network, telephony, and connectivity with client systems established."},
				{Key: "setup_omnichannel", Label: "ZuZo Omnichannel Setup", Description: "Communication channels, routing, and IVR configured."},
				{Key: "setup_workstations", Label: "Agent Workstations Setup", Description: "Hardware, software, and user accounts provisioned."},
			},
		},
		{
			Title: "Staff Training",
			Items: []ChecklistItem{
				{Key: "staff_recruitment", Label: "Resource Search & Acquisition", Description: "Job descriptions posted, candidates screened and hired."},
				{Key: "staff_training_curriculum", Label: "Staff Training", Description: "Training curriculum delivered on processes and procedures."},
				{Key: "staff_platform_training", Label: "ZuZo Platform Training", Description: "Agents trained on the ZuZo Omnichannel platform."},
				{Key: "staff_product_training", Label: "Client Product Training", Description: "Agents trained on client-specific products and terminology."},
			},
		},
	},
	"Pilot Launch": {
		{
			Title: "Setup Validation",
			Items: []ChecklistItem{
				{Key: "validate_testing", Label: "Testing & Validation", Description: "Infrastructure and system integration testing completed."},
				{Key: "validate_security", Label: "Security & Compliance", Description: "Access controls, data protection, and compliance measures verified."},
				{Key: "validate_dr", Label: "Disaster Recovery Plan", Description: "Backup, redundancy, and business continuity measures in place."},
				{Key: "validate_support", Label: "Ongoing Support Setup", Description: "Support procedures, escalation paths, and monitoring established."},
			},
		},
		{
			Title: "Staff Readiness",
			Items: []ChecklistItem{
				{Key: "readiness_qa", Label: "Quality Assurance Standards", Description: "QA standards and monitoring processes introduced to staff."},
				{Key: "readiness_certification", Label: "Certification", Description: "Staff assessed and certified on products and systems."},
				{Key: "readiness_nesting", Label: "Nesting & Live Onboarding", Description: "New agents mentored through supervised live interactions."},
				{Key: "readiness_continuous", Label: "Continuous Development Plan", Description: "Ongoing coaching and refresher training scheduled."},
			},
		},
		{
			Title: "Platform & Product Testing",
			Items: []ChecklistItem{
				{Key: "testing_plan", Label: "Test Plan Development", Description: "Test scenarios, cases, and acceptance criteria defined."},
				{Key: "testing_system", Label: "System Testing", Description: "ZuZo Omnichannel functionality and integrations verified."},
				{Key: "testing_process", Label: "Process Testing", Description: "End-to-end process flows and escalation procedures validated."},
				{Key: "testing_uat", Label: "User Acceptance Testing", Description: "Client-led testing completed and feedback addressed."},
				{Key: "testing_performance", Label: "Performance Testing", Description: "Load, stress, and response-time testing completed."},
				{Key: "testing_mock_ops", Label: "Mock Operations", Description: "Full-scale simulation of live operations conducted."},
				{Key: "testing_issues", Label: "Issue Management", Description: "Identified issues logged, prioritized, and resolved."},
				{Key: "testing_readiness", Label: "Final Readiness Assessment", Description: "Go/no-go recommendation made for go-live."},
			},
		},
	},
	"Full Deployment": {
		{
			Title: "Go Live",
			Items: []ChecklistItem{
				{Key: "golive_prep", Label: "Go-Live Preparation", Description: "Final systems check, staff scheduling, and command center ready."},
				{Key: "golive_contact_centre", Label: "Contact Centre Go-Live", Description: "Live operations activated with floor support in place."},
				{Key: "golive_platform", Label: "Platform & Product Go-Live", Description: "All production systems activated and monitored."},
				{Key: "golive_hypercare", Label: "Hypercare Support", Description: "Enhanced monitoring and rapid issue resolution during initial period."},
				{Key: "golive_bau", Label: "Transition to Business as Usual", Description: "Hypercare wound down; standard governance and reporting in place."},
			},
		},
	},
}

// checklistItemKeys flattens every item key for a stage across its sub-steps.
func checklistItemKeys(stage string) []string {
	var keys []string
	for _, sub := range checklistCatalog[stage] {
		for _, item := range sub.Items {
			keys = append(keys, item.Key)
		}
	}
	return keys
}

func isValidChecklistItem(stage, itemKey string) bool {
	return slices.Contains(checklistItemKeys(stage), itemKey)
}

// nextOnboardingStage returns the stage after the given one, or ("", true)
// if stage is already the last one.
func nextOnboardingStage(stage string) (next string, isLast bool) {
	idx := slices.Index(onboardingStages, stage)
	if idx == -1 || idx == len(onboardingStages)-1 {
		return "", true
	}
	return onboardingStages[idx+1], false
}

// currentSubStepTitle returns the title of the first sub-step within a
// stage that still has an incomplete item — i.e. what the partner should
// see as "what's being worked on right now" within their current stage.
// Falls back to the last sub-step's title if everything is already
// complete (a transient state in practice, since completing a stage
// immediately advances onboarding_stage to the next one).
func currentSubStepTitle(stage string, completedKeys map[string]bool) string {
	subSteps := checklistCatalog[stage]
	for _, sub := range subSteps {
		for _, item := range sub.Items {
			if !completedKeys[item.Key] {
				return sub.Title
			}
		}
	}
	if len(subSteps) > 0 {
		return subSteps[len(subSteps)-1].Title
	}
	return ""
}

type checklistItemResponse struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type checklistSubStepResponse struct {
	Title string                  `json:"title"`
	Items []checklistItemResponse `json:"items"`
}

type stageSummary struct {
	Stage          string `json:"stage"`
	TotalItems     int    `json:"totalItems"`
	CompletedItems int    `json:"completedItems"`
	IsComplete     bool   `json:"isComplete"`
}

// AdminGetOnboardingChecklistSummaryHandler returns completion counts for
// all 4 stages in one call — used to render the stage list (locked/current/
// complete) without fetching every stage's full item catalog.
// GET /api/admin/partners/{id}/onboarding-checklist/summary
func (s *Server) AdminGetOnboardingChecklistSummaryHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}
	isAdmin, err := s.Supabase.IsAdmin(r.Context(), user.ID)
	if err != nil {
		log.Printf("IsAdmin check failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to verify admin status"})
		return
	}
	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
		return
	}

	partnerID := r.PathValue("id")
	summaries := make([]stageSummary, 0, len(onboardingStages))
	var currentSubStage string
	foundCurrent := false
	for _, stage := range onboardingStages {
		totalKeys := checklistItemKeys(stage)
		completions, err := s.Supabase.ListChecklistCompletions(r.Context(), partnerID, stage)
		if err != nil {
			log.Printf("ListChecklistCompletions failed for stage %q: %v", stage, err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load checklist summary"})
			return
		}
		completedKeys := make(map[string]bool, len(completions))
		completedCount := 0
		for _, c := range completions {
			if c.Completed {
				completedCount++
				completedKeys[c.ItemKey] = true
			}
		}
		isComplete := completedCount == len(totalKeys)
		summaries = append(summaries, stageSummary{
			Stage:          stage,
			TotalItems:     len(totalKeys),
			CompletedItems: completedCount,
			IsComplete:     isComplete,
		})
		// The first not-yet-complete stage is the one currently being worked
		// on — capture its in-progress sub-step for the admin's collapsed
		// stage pill (mirrors currentSubStepTitle's use on the partner side).
		if !isComplete && !foundCurrent {
			currentSubStage = currentSubStepTitle(stage, completedKeys)
			foundCurrent = true
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"stages": summaries, "currentSubStage": currentSubStage})
}

// AdminGetOnboardingChecklistHandler returns the item catalog for one stage
// merged with this partner's completion state.
// GET /api/admin/partners/{id}/onboarding-checklist?stage=...
func (s *Server) AdminGetOnboardingChecklistHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}
	isAdmin, err := s.Supabase.IsAdmin(r.Context(), user.ID)
	if err != nil {
		log.Printf("IsAdmin check failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to verify admin status"})
		return
	}
	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
		return
	}

	stage := r.URL.Query().Get("stage")
	subSteps, ok := checklistCatalog[stage]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown stage"})
		return
	}

	partnerID := r.PathValue("id")
	completions, err := s.Supabase.ListChecklistCompletions(r.Context(), partnerID, stage)
	if err != nil {
		log.Printf("ListChecklistCompletions failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load checklist"})
		return
	}
	completedKeys := make(map[string]bool, len(completions))
	for _, c := range completions {
		if c.Completed {
			completedKeys[c.ItemKey] = true
		}
	}

	respSubSteps := make([]checklistSubStepResponse, 0, len(subSteps))
	for _, sub := range subSteps {
		items := make([]checklistItemResponse, 0, len(sub.Items))
		for _, item := range sub.Items {
			items = append(items, checklistItemResponse{
				Key:         item.Key,
				Label:       item.Label,
				Description: item.Description,
				Completed:   completedKeys[item.Key],
			})
		}
		respSubSteps = append(respSubSteps, checklistSubStepResponse{Title: sub.Title, Items: items})
	}

	writeJSON(w, http.StatusOK, map[string]any{"stage": stage, "subSteps": respSubSteps})
}

type toggleChecklistItemRequest struct {
	Stage     string `json:"stage"`
	Completed bool   `json:"completed"`
}

// AdminToggleChecklistItemHandler marks one checklist item done/not-done.
// PATCH /api/admin/partners/{id}/onboarding-checklist/{itemKey}
func (s *Server) AdminToggleChecklistItemHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}
	isAdmin, err := s.Supabase.IsAdmin(r.Context(), user.ID)
	if err != nil {
		log.Printf("IsAdmin check failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to verify admin status"})
		return
	}
	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
		return
	}

	var req toggleChecklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	itemKey := r.PathValue("itemKey")
	if !isValidChecklistItem(req.Stage, itemKey) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown stage or checklist item"})
		return
	}

	partnerID := r.PathValue("id")
	if err := s.Supabase.SetChecklistItemCompletion(r.Context(), partnerID, req.Stage, itemKey, req.Completed); err != nil {
		log.Printf("SetChecklistItemCompletion failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to update checklist item"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"key": itemKey, "completed": req.Completed})
}

type completeStageRequest struct {
	Stage string `json:"stage"`
}

// AdminCompleteOnboardingStageHandler verifies every checklist item for a
// stage is complete, then advances the partner to the next stage (or, if
// this was the last stage, flips their status to 'active').
// POST /api/admin/partners/{id}/onboarding-checklist/complete-stage
func (s *Server) AdminCompleteOnboardingStageHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}
	isAdmin, err := s.Supabase.IsAdmin(r.Context(), user.ID)
	if err != nil {
		log.Printf("IsAdmin check failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to verify admin status"})
		return
	}
	if !isAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
		return
	}

	var req completeStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if _, ok := checklistCatalog[req.Stage]; !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown stage"})
		return
	}

	partnerID := r.PathValue("id")
	completions, err := s.Supabase.ListChecklistCompletions(r.Context(), partnerID, req.Stage)
	if err != nil {
		log.Printf("ListChecklistCompletions failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load checklist"})
		return
	}
	completedKeys := make(map[string]bool, len(completions))
	for _, c := range completions {
		if c.Completed {
			completedKeys[c.ItemKey] = true
		}
	}

	var missing []string
	for _, key := range checklistItemKeys(req.Stage) {
		if !completedKeys[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   "checklist is not fully complete",
			"missing": missing,
		})
		return
	}

	next, isLast := nextOnboardingStage(req.Stage)
	target := req.Stage
	if !isLast {
		target = next
	}
	if err := s.Supabase.SetOnboardingStage(r.Context(), partnerID, target); err != nil {
		log.Printf("SetOnboardingStage failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to advance onboarding stage"})
		return
	}
	if isLast {
		if err := s.Supabase.SetPartnerStatus(r.Context(), partnerID, "active"); err != nil {
			log.Printf("SetPartnerStatus failed: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "stage advanced but failed to activate partner"})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"completedStage": req.Stage,
		"nextStage":      next,
		"isFinalStage":   isLast,
	})
}

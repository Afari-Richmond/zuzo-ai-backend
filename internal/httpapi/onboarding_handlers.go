package httpapi

import (
	"log"
	"net/http"
)

// onboardingStages must stay byte-for-byte identical to the check
// constraint on partners.onboarding_stage (client/supabase/schema.sql,
// section 22), the frontend's ONBOARDING_STAGES (use-admin-partners.tsx),
// and checklistCatalog's keys (checklist_handlers.go) — there's no shared
// codegen across Go/SQL/TypeScript in this stack, so this three-way list
// has to be kept in sync by hand. Advancing through these stages now only
// happens via AdminCompleteOnboardingStageHandler (checklist_handlers.go),
// gated on every checklist item for a stage being complete — there is no
// direct "set any stage" admin endpoint.
var onboardingStages = []string{
	"Assessment & Setup",
	"Training & Integration",
	"Pilot Launch",
	"Full Deployment",
}

// PartnerGetOnboardingStageHandler returns the logged-in partner's current
// onboarding stage and whether they've already seen it change.
// GET /api/partner/onboarding-stage
func (s *Server) PartnerGetOnboardingStageHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}

	partner, err := s.Supabase.FindPartnerByAuthUserID(r.Context(), user.ID)
	if err != nil {
		log.Printf("FindPartnerByAuthUserID failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load partner"})
		return
	}
	if partner == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no partner record for this account"})
		return
	}

	unseen := partner.Stage != nil && partner.SeenAt == nil

	var subStage string
	if partner.Stage != nil {
		completions, err := s.Supabase.ListChecklistCompletions(r.Context(), partner.ID, *partner.Stage)
		if err != nil {
			log.Printf("ListChecklistCompletions failed: %v", err)
			// Non-fatal — the stage badge still renders without a sub-stage.
		} else {
			completedKeys := make(map[string]bool, len(completions))
			for _, c := range completions {
				if c.Completed {
					completedKeys[c.ItemKey] = true
				}
			}
			subStage = currentSubStepTitle(*partner.Stage, completedKeys)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stage":    partner.Stage,
		"subStage": subStage,
		"unseen":   unseen,
	})
}

// partnerStageProgress is one stage's entry in the full progress view — the
// partner-facing analog of AdminGetOnboardingChecklistSummaryHandler +
// AdminGetOnboardingChecklistHandler combined into a single read-only call,
// since a partner views this rarely (unlike admin, which polls/toggles).
type partnerStageProgress struct {
	Stage          string                     `json:"stage"`
	Status         string                     `json:"status"` // "complete" | "current" | "locked"
	TotalItems     int                        `json:"totalItems"`
	CompletedItems int                        `json:"completedItems"`
	SubSteps       []checklistSubStepResponse `json:"subSteps"`
}

// PartnerGetOnboardingProgressHandler returns every onboarding stage with
// its lock/current/complete status and full sub-step/item breakdown, so the
// partner can see exactly where they are and what's left — unlike
// PartnerGetOnboardingStageHandler, which only exposes the current stage's
// label. Read-only: no PATCH counterpart, unlike the admin checklist
// endpoints this mirrors.
// GET /api/partner/onboarding-progress
func (s *Server) PartnerGetOnboardingProgressHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}

	partner, err := s.Supabase.FindPartnerByAuthUserID(r.Context(), user.ID)
	if err != nil {
		log.Printf("FindPartnerByAuthUserID failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load partner"})
		return
	}
	if partner == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no partner record for this account"})
		return
	}

	currentIndex := -1
	if partner.Stage != nil {
		currentIndex = 0
		for i, stage := range onboardingStages {
			if stage == *partner.Stage {
				currentIndex = i
				break
			}
		}
	}

	stages := make([]partnerStageProgress, 0, len(onboardingStages))
	for i, stage := range onboardingStages {
		subSteps := checklistCatalog[stage]
		totalKeys := checklistItemKeys(stage)

		// Always read the real completion state for every stage — a stage
		// is only supposed to advance once every item in it is complete
		// (AdminCompleteOnboardingStageHandler re-validates this before
		// advancing), but admin can still toggle any item in any stage
		// directly (including a passed stage, or one not reached yet —
		// the admin UI no longer locks stage navigation), and the
		// partner's view needs to reflect that live rather than assuming
		// a past stage is frozen at 100% or a future one is untouched.
		var status string
		switch {
		case currentIndex == -1:
			status = "locked"
		case i < currentIndex:
			status = "complete"
		case i == currentIndex:
			status = "current"
		default:
			status = "locked"
		}

		completions, err := s.Supabase.ListChecklistCompletions(r.Context(), partner.ID, stage)
		if err != nil {
			log.Printf("ListChecklistCompletions failed for stage %q: %v", stage, err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load onboarding progress"})
			return
		}
		completedKeys := make(map[string]bool, len(completions))
		for _, c := range completions {
			if c.Completed {
				completedKeys[c.ItemKey] = true
			}
		}

		respSubSteps := make([]checklistSubStepResponse, 0, len(subSteps))
		completedCount := 0
		for _, sub := range subSteps {
			items := make([]checklistItemResponse, 0, len(sub.Items))
			for _, item := range sub.Items {
				completed := completedKeys[item.Key]
				if completed {
					completedCount++
				}
				items = append(items, checklistItemResponse{
					Key:         item.Key,
					Label:       item.Label,
					Description: item.Description,
					Completed:   completed,
				})
			}
			respSubSteps = append(respSubSteps, checklistSubStepResponse{Title: sub.Title, Items: items})
		}

		stages = append(stages, partnerStageProgress{
			Stage:          stage,
			Status:         status,
			TotalItems:     len(totalKeys),
			CompletedItems: completedCount,
			SubSteps:       respSubSteps,
		})
	}

	var currentStage *string
	if currentIndex != -1 {
		currentStage = &onboardingStages[currentIndex]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"currentStage": currentStage,
		"stages":       stages,
	})
}

// PartnerAckOnboardingStageHandler marks the current stage as seen, so the
// one-time success modal doesn't fire again until the stage next changes.
// POST /api/partner/onboarding-stage/ack
func (s *Server) PartnerAckOnboardingStageHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return
	}

	partner, err := s.Supabase.FindPartnerByAuthUserID(r.Context(), user.ID)
	if err != nil {
		log.Printf("FindPartnerByAuthUserID failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to load partner"})
		return
	}
	if partner == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no partner record for this account"})
		return
	}

	if err := s.Supabase.MarkOnboardingStageSeen(r.Context(), partner.ID); err != nil {
		log.Printf("MarkOnboardingStageSeen failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to acknowledge onboarding stage"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

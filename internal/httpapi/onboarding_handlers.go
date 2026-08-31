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

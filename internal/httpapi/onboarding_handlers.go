package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

// onboardingStages must stay byte-for-byte identical to the check
// constraint on partners.onboarding_stage (client/supabase/schema.sql,
// section 22) and the frontend's ONBOARDING_STAGES (use-admin-partners.tsx)
// — there's no shared codegen across Go/SQL/TypeScript in this stack, so
// this three-way list has to be kept in sync by hand.
var onboardingStages = []string{
	"Assessment & Setup",
	"Training & Integration",
	"Pilot Launch",
	"Full Deployment",
}

func isValidOnboardingStage(stage string) bool {
	return slices.Contains(onboardingStages, stage)
}

type setOnboardingStageRequest struct {
	Stage string `json:"stage"`
}

// AdminSetOnboardingStageHandler lets an admin set a partner's onboarding
// stage. PATCH /api/admin/partners/{id}/onboarding-stage
func (s *Server) AdminSetOnboardingStageHandler(w http.ResponseWriter, r *http.Request) {
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

	var req setOnboardingStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !isValidOnboardingStage(req.Stage) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stage must be one of: " + joinStages()})
		return
	}

	partnerID := r.PathValue("id")
	if partnerID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "partner id is required"})
		return
	}

	if err := s.Supabase.SetOnboardingStage(r.Context(), partnerID, req.Stage); err != nil {
		log.Printf("SetOnboardingStage failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to update onboarding stage"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"partnerId": partnerID, "stage": req.Stage})
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
	writeJSON(w, http.StatusOK, map[string]any{
		"stage":  partner.Stage,
		"unseen": unseen,
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

func joinStages() string {
	return strings.Join(onboardingStages, ", ")
}

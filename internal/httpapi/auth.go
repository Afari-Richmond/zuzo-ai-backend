package httpapi

import (
	"context"
	"net/http"
	"strings"

	"zuzo.com/backend/internal/supabase"
)

type ctxKey string

const userCtxKey ctxKey = "authUser"

// Auth verifies the request's bearer token against Supabase Auth and stashes
// the resolved user on the request context. It's applied per-route (not via
// the global middleware Chain in main.go) since it must gate only the
// partner/admin data endpoints, not /api/health or the existing AI chat
// endpoint.
//
// configured must be false whenever SUPABASE_URL/SUPABASE_SERVICE_ROLE_KEY
// aren't set — checked before attempting verification, since VerifyUser
// against an empty base URL would otherwise fail and surface as a
// misleading "invalid or expired session" instead of "not configured yet".
func Auth(supa *supabase.Client, configured bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !configured {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "onboarding-stage feature isn't configured"})
				return
			}

			authHeader := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			user, err := supa.VerifyUser(r.Context(), token)
			if err != nil {
				http.Error(w, "invalid or expired session", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userFromContext(r *http.Request) (*supabase.AuthUser, bool) {
	user, ok := r.Context().Value(userCtxKey).(*supabase.AuthUser)
	return user, ok
}

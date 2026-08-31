// Package supabase talks to a Supabase project's Auth (GoTrue) and PostgREST
// APIs directly over HTTP, using only the Go standard library — no
// third-party SDK or Postgres driver dependency (mirrors the hand-rolled
// HTTP client style in internal/ai/gemini.go).
//
// All PostgREST calls authenticate as the service-role key, which bypasses
// Row Level Security entirely. That means the authorization checks RLS
// would otherwise provide (is_admin(), my_partner_id(), etc. in
// client/supabase/schema.sql) must be reimplemented here in Go — see
// IsAdmin and FindPartnerByAuthUserID below.
package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL        string
	serviceRoleKey string
	httpClient     *http.Client
}

func NewClient(baseURL, serviceRoleKey string) *Client {
	return &Client{
		baseURL:        strings.TrimRight(baseURL, "/"),
		serviceRoleKey: serviceRoleKey,
		httpClient:     &http.Client{Timeout: 15 * time.Second},
	}
}

// AuthUser is the caller identity resolved from a bearer token.
type AuthUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// VerifyUser asks Supabase's own Auth API to validate the caller's bearer
// token (rather than verifying the JWT signature ourselves, which would
// need a JWKS/JWT library this stdlib-only, no-network-egress-for-`go get`
// binary doesn't have). A 200 response means Supabase considers the token a
// valid, current session.
func (c *Client) VerifyUser(ctx context.Context, bearerToken string) (*AuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/v1/user", nil)
	if err != nil {
		return nil, fmt.Errorf("build auth/v1/user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("apikey", c.serviceRoleKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call auth/v1/user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth/v1/user returned status %d", resp.StatusCode)
	}

	var user AuthUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode auth/v1/user response: %w", err)
	}
	if user.ID == "" {
		return nil, fmt.Errorf("auth/v1/user response had no id")
	}
	return &user, nil
}

// restGet performs a PostgREST GET against the given table + query string
// (already url-encoded), decoding the JSON array response into dest.
func (c *Client) restGet(ctx context.Context, table, rawQuery string, dest any) error {
	reqURL := fmt.Sprintf("%s/rest/v1/%s?%s", c.baseURL, table, rawQuery)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("build %s request: %w", table, err)
	}
	c.setServiceRoleHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", table, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s returned status %d: %s", table, resp.StatusCode, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode %s response: %w", table, err)
	}
	return nil
}

// restPatch performs a PostgREST PATCH against the given table + filter
// query string, with the given JSON-encodable body.
func (c *Client) restPatch(ctx context.Context, table, rawQuery string, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode %s patch body: %w", table, err)
	}

	reqURL := fmt.Sprintf("%s/rest/v1/%s?%s", c.baseURL, table, rawQuery)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build %s patch request: %w", table, err)
	}
	c.setServiceRoleHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s patch: %w", table, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s patch returned status %d: %s", table, resp.StatusCode, string(respBody))
	}
	return nil
}

func (c *Client) setServiceRoleHeaders(req *http.Request) {
	req.Header.Set("apikey", c.serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)
}

// restUpsert performs a PostgREST POST with an upsert Prefer header against
// the given table, resolving conflicts on the given comma-separated column
// list (must match a unique constraint on the table).
func (c *Client) restUpsert(ctx context.Context, table, onConflict string, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode %s upsert body: %w", table, err)
	}

	reqURL := fmt.Sprintf("%s/rest/v1/%s?on_conflict=%s", c.baseURL, table, url.QueryEscape(onConflict))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build %s upsert request: %w", table, err)
	}
	c.setServiceRoleHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s upsert: %w", table, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s upsert returned status %d: %s", table, resp.StatusCode, string(respBody))
	}
	return nil
}

// IsAdmin replicates the is_admin() SQL function: does a row exist in
// `admins` whose id equals this Supabase Auth user id?
func (c *Client) IsAdmin(ctx context.Context, userID string) (bool, error) {
	var rows []struct {
		ID string `json:"id"`
	}
	query := fmt.Sprintf("id=eq.%s&select=id", url.QueryEscape(userID))
	if err := c.restGet(ctx, "admins", query, &rows); err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

// PartnerOnboarding is the subset of a partners row the onboarding-stage
// feature needs.
type PartnerOnboarding struct {
	ID     string  `json:"id"`
	Stage  *string `json:"onboarding_stage"`
	SeenAt *string `json:"onboarding_stage_seen_at"`
}

// FindPartnerByAuthUserID replicates my_partner_id(): resolves a Supabase
// Auth user id to their own partners row. Never trusts a client-supplied
// partner id — the caller's identity comes only from their verified token.
func (c *Client) FindPartnerByAuthUserID(ctx context.Context, userID string) (*PartnerOnboarding, error) {
	var rows []PartnerOnboarding
	query := fmt.Sprintf(
		"auth_user_id=eq.%s&select=id,onboarding_stage,onboarding_stage_seen_at",
		url.QueryEscape(userID),
	)
	if err := c.restGet(ctx, "partners", query, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// SetOnboardingStage sets a partner's onboarding stage (admin-only action,
// authorization enforced by the caller before this is invoked).
func (c *Client) SetOnboardingStage(ctx context.Context, partnerID, stage string) error {
	query := fmt.Sprintf("id=eq.%s", url.QueryEscape(partnerID))
	return c.restPatch(ctx, "partners", query, map[string]string{"onboarding_stage": stage})
}

// MarkOnboardingStageSeen stamps onboarding_stage_seen_at with the current
// time, acknowledging the partner has seen the one-time stage-change modal.
func (c *Client) MarkOnboardingStageSeen(ctx context.Context, partnerID string) error {
	query := fmt.Sprintf("id=eq.%s", url.QueryEscape(partnerID))
	return c.restPatch(ctx, "partners", query, map[string]string{
		"onboarding_stage_seen_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// SetPartnerStatus sets a partner's lifecycle status (admin-only action,
// authorization enforced by the caller before this is invoked). Used to
// flip a partner from 'onboarding' to 'active' once every onboarding-stage
// checklist item, across all stages, is complete.
func (c *Client) SetPartnerStatus(ctx context.Context, partnerID, status string) error {
	query := fmt.Sprintf("id=eq.%s", url.QueryEscape(partnerID))
	return c.restPatch(ctx, "partners", query, map[string]string{"status": status})
}

// ChecklistItemCompletion is one row of onboarding-checklist completion
// state for a partner.
type ChecklistItemCompletion struct {
	ItemKey   string `json:"item_key"`
	Completed bool   `json:"completed"`
}

// ListChecklistCompletions returns the completion state of every checklist
// item recorded so far for a partner within one stage. Items never toggled
// on for this partner simply won't appear — callers should treat any item
// key absent from the result as incomplete.
func (c *Client) ListChecklistCompletions(ctx context.Context, partnerID, stage string) ([]ChecklistItemCompletion, error) {
	var rows []ChecklistItemCompletion
	query := fmt.Sprintf(
		"partner_id=eq.%s&stage=eq.%s&select=item_key,completed",
		url.QueryEscape(partnerID), url.QueryEscape(stage),
	)
	if err := c.restGet(ctx, "partner_onboarding_checklist", query, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// SetChecklistItemCompletion upserts one checklist item's completion state
// for a partner (admin-only action, item key validated by the caller
// against the static per-stage item catalog before this is invoked).
func (c *Client) SetChecklistItemCompletion(ctx context.Context, partnerID, stage, itemKey string, completed bool) error {
	var completedAt any
	if completed {
		completedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return c.restUpsert(ctx, "partner_onboarding_checklist", "partner_id,item_key", map[string]any{
		"partner_id":   partnerID,
		"stage":        stage,
		"item_key":     itemKey,
		"completed":    completed,
		"completed_at": completedAt,
	})
}

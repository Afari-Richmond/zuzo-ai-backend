package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"zuzo.com/backend/internal/ai"
	"zuzo.com/backend/internal/data"
)

const (
	maxMessageChars = 2000
	maxHistoryTurns = 10
)

type Server struct {
	AI           *ai.Client
	AIConfigured bool
}

type chatRequest struct {
	PartnerID      string           `json:"partnerId"`
	Message        string           `json:"message"`
	History        []ai.ChatMessage `json:"history"`
	CurrentSection string           `json:"currentSection"`
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ChatHandler(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if req.PartnerID == "" || req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "partnerId and message are required"})
		return
	}
	if len(req.Message) > maxMessageChars {
		req.Message = req.Message[:maxMessageChars]
	}
	if len(req.History) > maxHistoryTurns {
		req.History = req.History[len(req.History)-maxHistoryTurns:]
	}

	// The server looks up this partner's own data by id — it never accepts a data
	// payload from the client, so one partner's session can never smuggle another
	// partner's information into its own context.
	bundle, ok := data.GetBundle(req.PartnerID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown partner"})
		return
	}

	if !s.AIConfigured {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "AI assistant isn't configured"})
		return
	}

	systemPrompt := buildSystemPrompt(bundle, req.CurrentSection)
	messages := append(sanitizeHistory(req.History), ai.ChatMessage{Role: "user", Content: req.Message})

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	writeFrame := func(payload map[string]string) {
		b, _ := json.Marshal(payload)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}

	err := s.AI.StreamReply(r.Context(), systemPrompt, messages, func(delta string) {
		writeFrame(map[string]string{"delta": delta})
	})
	if err != nil {
		log.Printf("AI stream error: %v", err)
		writeFrame(map[string]string{"error": "ZIA hit a snag answering that — please try again."})
	}
}

func sanitizeHistory(history []ai.ChatMessage) []ai.ChatMessage {
	clean := make([]ai.ChatMessage, 0, len(history))
	for _, m := range history {
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		clean = append(clean, m)
	}
	return clean
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

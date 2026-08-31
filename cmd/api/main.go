package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zuzo.com/backend/internal/ai"
	"zuzo.com/backend/internal/config"
	"zuzo.com/backend/internal/httpapi"
	"zuzo.com/backend/internal/supabase"
)

func main() {
	cfg := config.Load()

	if cfg.GeminiAPIKey == "" {
		log.Println("warning: GEMINI_API_KEY is not set — /api/partner/ai/chat will return 503 until it is configured")
	}

	supabaseConfigured := cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != ""
	if !supabaseConfigured {
		log.Println("warning: SUPABASE_URL/SUPABASE_SERVICE_ROLE_KEY not set — onboarding-stage endpoints will return 503 until configured")
	}
	supa := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey)

	server := &httpapi.Server{
		AI:           ai.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel),
		AIConfigured: cfg.GeminiAPIKey != "",

		Supabase: supa,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", server.HealthHandler)
	mux.HandleFunc("POST /api/partner/ai/chat", server.ChatHandler)
	mux.Handle("PATCH /api/admin/partners/{id}/onboarding-stage", httpapi.Auth(supa, supabaseConfigured)(http.HandlerFunc(server.AdminSetOnboardingStageHandler)))
	mux.Handle("GET /api/partner/onboarding-stage", httpapi.Auth(supa, supabaseConfigured)(http.HandlerFunc(server.PartnerGetOnboardingStageHandler)))
	mux.Handle("POST /api/partner/onboarding-stage/ack", httpapi.Auth(supa, supabaseConfigured)(http.HandlerFunc(server.PartnerAckOnboardingStageHandler)))

	handler := httpapi.Chain(mux, httpapi.Recover, httpapi.RequestLogger, httpapi.CORS(cfg.CORSOrigin))

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		log.Printf("ZuZo AI backend listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

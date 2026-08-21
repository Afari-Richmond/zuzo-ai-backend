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
)

func main() {
	cfg := config.Load()

	if cfg.GeminiAPIKey == "" {
		log.Println("warning: GEMINI_API_KEY is not set — /api/partner/ai/chat will return 503 until it is configured")
	}

	server := &httpapi.Server{
		AI:           ai.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel),
		AIConfigured: cfg.GeminiAPIKey != "",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", server.HealthHandler)
	mux.HandleFunc("POST /api/partner/ai/chat", server.ChatHandler)

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

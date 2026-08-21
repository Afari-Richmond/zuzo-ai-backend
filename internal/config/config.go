package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port         string
	GeminiAPIKey string
	GeminiModel  string
	CORSOrigin   string
}

func Load() Config {
	loadDotEnv(".env")

	return Config{
		Port:         getEnv("PORT", "4100"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-3.5-flash-lite"),
		CORSOrigin:   getEnv("CORS_ORIGIN", "*"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv is a minimal KEY=VALUE .env reader (mirrors the Node server's use of
// dotenv). It never overrides variables already set in the real environment, and
// silently no-ops if the file doesn't exist — a missing .env just means you rely
// on real env vars instead.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

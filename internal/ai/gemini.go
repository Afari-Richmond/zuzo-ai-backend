// Package ai talks to the Google Gemini API directly over HTTP/SSE
// (https://ai.google.dev/api/generate-content#method:-models.streamgeneratecontent)
// using only the Go standard library — no third-party SDK dependency.
package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Generous headroom: this model generation spends part of its output budget on
// a hidden "thinking" pass before the visible answer, so max_output_tokens has
// to cover both.
const maxOutputTokens = 4096

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type geminiPart struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  struct {
		MaxOutputTokens int `json:"maxOutputTokens"`
	} `json:"generationConfig"`
}

type geminiStreamChunk struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// geminiRole maps our internal role names to Gemini's ("user" / "model").
func geminiRole(role string) string {
	if role == "assistant" {
		return "model"
	}
	return "user"
}

// StreamReply streams ZIA's reply for one chat turn, invoking onDelta for each
// chunk of text as it arrives. It returns once the model finishes or an error
// occurs (including a partial stream that errors mid-way).
func (c *Client) StreamReply(ctx context.Context, systemPrompt string, history []ChatMessage, onDelta func(string)) error {
	contents := make([]geminiContent, 0, len(history))
	for _, m := range history {
		contents = append(contents, geminiContent{
			Role:  geminiRole(m.Role),
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: systemPrompt}}},
		Contents:          contents,
	}
	reqBody.GenerationConfig.MaxOutputTokens = maxOutputTokens

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse", c.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody geminiStreamChunk
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error != nil {
			return fmt.Errorf("Gemini API error (%d): %s", resp.StatusCode, errBody.Error.Message)
		}
		return fmt.Errorf("Gemini API returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var dataLines []string
	flush := func() error {
		if len(dataLines) == 0 {
			return nil
		}
		payload := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]

		var chunk geminiStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return nil // ignore malformed/unknown frames rather than aborting the whole stream
		}
		if chunk.Error != nil {
			return fmt.Errorf("Gemini stream error: %s", chunk.Error.Message)
		}
		for _, cand := range chunk.Candidates {
			for _, part := range cand.Content.Parts {
				if part.Text != "" && !part.Thought {
					onDelta(part.Text)
				}
			}
		}
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "data:"):
			dataLines = append(dataLines, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		case line == "":
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := flush(); err != nil {
		return err
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}

	return nil
}

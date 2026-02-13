// Package copilot – llm.go implements the LLM client for chat completions.
// Uses the OpenAI-compatible API format, which works with OpenAI, Anthropic
// proxies, GLM (api.z.ai), and any compatible endpoint.
package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// LLMClient handles communication with the LLM provider API.
type LLMClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewLLMClient creates a new LLM client from config.
func NewLLMClient(cfg *Config, logger *slog.Logger) *LLMClient {
	baseURL := cfg.API.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	// Ensure no trailing slash.
	baseURL = strings.TrimRight(baseURL, "/")

	return &LLMClient{
		baseURL: baseURL,
		apiKey:  cfg.API.APIKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger.With("component", "llm"),
	}
}

// chatMessage represents a message in the OpenAI chat format.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the OpenAI-compatible chat completions request.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

// chatResponse is the OpenAI-compatible chat completions response.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// Complete sends a chat completion request and returns the response text.
func (c *LLMClient) Complete(ctx context.Context, systemPrompt string, history []ConversationEntry, userMessage string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("API key not configured. Run 'copilot config set-key' or set GOCLAW_API_KEY")
	}

	// Build messages array.
	messages := make([]chatMessage, 0, len(history)*2+2)

	// System prompt.
	if systemPrompt != "" {
		messages = append(messages, chatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Conversation history.
	for _, entry := range history {
		messages = append(messages, chatMessage{
			Role:    "user",
			Content: entry.UserMessage,
		})
		if entry.AssistantResponse != "" {
			messages = append(messages, chatMessage{
				Role:    "assistant",
				Content: entry.AssistantResponse,
			})
		}
	}

	// Current user message.
	messages = append(messages, chatMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Build request (no temperature — some models only support default).
	reqBody := chatRequest{
		Model:    c.model,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	// Send HTTP request.
	endpoint := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	c.logger.Debug("sending chat completion",
		"model", c.model,
		"messages", len(messages),
		"endpoint", endpoint,
	)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	duration := time.Since(start)

	// Handle HTTP errors.
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("API error",
			"status", resp.StatusCode,
			"body", truncate(string(respBody), 200),
		)
		return "", fmt.Errorf("API returned %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	// Parse response.
	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	// Check for API-level error.
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	c.logger.Info("chat completion done",
		"model", c.model,
		"duration_ms", duration.Milliseconds(),
		"prompt_tokens", chatResp.Usage.PromptTokens,
		"completion_tokens", chatResp.Usage.CompletionTokens,
	)

	return content, nil
}

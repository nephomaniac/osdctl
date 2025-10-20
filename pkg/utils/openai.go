package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAIClient wraps the OpenAI-compatible API client
type OpenAIClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewOpenAIClient creates a new OpenAI-compatible API client
func NewOpenAIClient(baseURL, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ChatMessage represents a message in the chat completion
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request to the chat completion API
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// ChatCompletionResponse represents the response from the chat completion API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatCompletion makes a chat completion request to the OpenAI-compatible API
func (c *OpenAIClient) ChatCompletion(systemPrompt, userPrompt, model string) (string, error) {
	// Prepare request
	reqBody := ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "OpenAI request failed. URL:'%s'\n\nRequest data...\n'%s'\n", url, jsonData)
		// Check for authentication errors
		if resp.StatusCode == http.StatusUnauthorized {
			fmt.Fprintf(os.Stderr, "Authorization Headers...\n'%s'\n", strings.Join(req.Header["Authorization"], ","))
			return "", fmt.Errorf("authentication failed (status 401): Invalid or missing API key.\n\nPlease check your OpenAI key configuration:\n  - Set via config: osdctl config --key openai_key --value YOUR_KEY\n  - Set via env var: export OPENAI_API_KEY=YOUR_KEY\n  - Set via flag: --openai-key YOUR_KEY\n\nAPI response: %s", string(body))
		}
		// Check for forbidden errors
		if resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("authentication failed (status 403): API key does not have permission.\n\nPlease verify your OpenAI key has the correct permissions.\nAPI response: %s", string(body))
		}
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var completion ChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return completion.Choices[0].Message.Content, nil
}

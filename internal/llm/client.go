package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var (
	ErrEmptyResponse   = errors.New("empty response received from LLM")
	ErrInvalidJSON     = errors.New("failed to parse valid JSON from LLM response")
	ErrClientExecution = errors.New("failed to execute LLM completion request")
)

// StreamCallback receives chunks of text in real-time as the AI generates them.
type StreamCallback func(chunk string)

// Client wraps the OpenAI API client to support structured JSON generation
// across Ollama, Gemini, Claude, and OpenAI.
type Client struct {
	apiClient *openai.Client
	model     string
	provider  string
	baseURL   string
}

// NewClient creates an LLM client configured with provider credentials.
func NewClient(provider, baseURL, apiKey, model string) *Client {
	var clientConfig openai.ClientConfig

	if baseURL != "" {
		clientConfig = openai.DefaultConfig(apiKey)
		// Ensure BaseURL doesn't end with a trailing slash for consistency
		clientConfig.BaseURL = strings.TrimRight(baseURL, "/")
	} else {
		clientConfig = openai.DefaultConfig(apiKey)
	}

	apiClient := openai.NewClientWithConfig(clientConfig)

	return &Client{
		apiClient: apiClient,
		model:     model,
		provider:  provider,
		baseURL:   baseURL,
	}
}

// GenerateJSON executes a chat completion without streaming.
func (c *Client) GenerateJSON(ctx context.Context, systemPrompt, userPrompt string) ([]byte, error) {
	return c.GenerateJSONStream(ctx, systemPrompt, userPrompt, nil)
}

// GenerateJSONStream executes chat completion with real-time token streaming,
// automatic retry for transient spikes (503/429), and extracts clean JSON.
func (c *Client) GenerateJSONStream(ctx context.Context, systemPrompt, userPrompt string, onChunk StreamCallback) ([]byte, error) {
	const maxRetries = 4
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req := openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.2,
			Stream:      onChunk != nil,
		}

		if onChunk != nil {
			// Streaming mode
			stream, err := c.apiClient.CreateChatCompletionStream(ctx, req)
			if err != nil {
				errStr := strings.ToLower(err.Error())
				if strings.Contains(errStr, "response_format") || strings.Contains(errStr, "unsupported") || strings.Contains(errStr, "unknown") {
					req.ResponseFormat = nil
					stream, err = c.apiClient.CreateChatCompletionStream(ctx, req)
				}
			}

			if err == nil {
				var fullContent strings.Builder
				streamSucceeded := true

				for {
					response, recvErr := stream.Recv()
					if errors.Is(recvErr, io.EOF) {
						break
					}
					if recvErr != nil {
						lastErr = recvErr
						streamSucceeded = false
						break
					}
					if len(response.Choices) > 0 {
						chunk := response.Choices[0].Delta.Content
						fullContent.WriteString(chunk)
						if onChunk != nil && chunk != "" {
							onChunk(chunk)
						}
					}
				}
				stream.Close()

				if streamSucceeded {
					content := fullContent.String()
					if strings.TrimSpace(content) != "" {
						cleaned := CleanJSONResponse(content)
						if json.Valid(cleaned) {
							return cleaned, nil
						}
					}
				}
			} else {
				lastErr = err
			}
		}

		// Fallback or Non-streaming mode
		req.Stream = false
		resp, err := c.apiClient.CreateChatCompletion(ctx, req)

		if err != nil {
			lastErr = err
			errStr := strings.ToLower(err.Error())

			isTransient := strings.Contains(errStr, "503") ||
				strings.Contains(errStr, "429") ||
				strings.Contains(errStr, "500") ||
				strings.Contains(errStr, "502") ||
				strings.Contains(errStr, "504") ||
				strings.Contains(errStr, "demand") ||
				strings.Contains(errStr, "unavailable") ||
				strings.Contains(errStr, "rate") ||
				strings.Contains(errStr, "quota") ||
				strings.Contains(errStr, "deadline") ||
				strings.Contains(errStr, "timeout")

			if isTransient && attempt < maxRetries {
				backoff := time.Duration(attempt*2) * time.Second
				fmt.Printf("      ⏳ AI service busy (%s). Retrying in %v (attempt %d/%d)...\n", extractShortError(errStr), backoff, attempt, maxRetries)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}

			return nil, fmt.Errorf("%w: %v", ErrClientExecution, err)
		}

		if len(resp.Choices) == 0 {
			lastErr = ErrEmptyResponse
			continue
		}

		content := resp.Choices[0].Message.Content
		if strings.TrimSpace(content) == "" {
			lastErr = ErrEmptyResponse
			continue
		}

		cleaned := CleanJSONResponse(content)
		if !json.Valid(cleaned) {
			return nil, fmt.Errorf("%w: raw text received: %s", ErrInvalidJSON, content)
		}

		return cleaned, nil
	}

	return nil, fmt.Errorf("%w: exhausted retries: %v", ErrClientExecution, lastErr)
}

func extractShortError(errStr string) string {
	if strings.Contains(errStr, "503") || strings.Contains(errStr, "demand") {
		return "503 High Demand"
	}
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "quota") {
		return "429 Rate Limit"
	}
	if strings.Contains(errStr, "deadline") || strings.Contains(errStr, "timeout") {
		return "Timeout / Deadline"
	}
	return "Temporary Service Busy"
}

// CleanJSONResponse extracts and cleans raw JSON from markdown wrappers or commentary.
func CleanJSONResponse(raw string) []byte {
	trimmed := strings.TrimSpace(raw)

	// Strip markdown code fences if present (```json ... ``` or ``` ...)
	if strings.HasPrefix(trimmed, "```") {
		// Find first newline after ```
		newlineIdx := strings.Index(trimmed, "\n")
		if newlineIdx != -1 {
			trimmed = trimmed[newlineIdx+1:]
		} else {
			trimmed = strings.TrimPrefix(trimmed, "```")
		}

		// Find trailing ```
		if closingIdx := strings.LastIndex(trimmed, "```"); closingIdx != -1 {
			trimmed = trimmed[:closingIdx]
		}
		trimmed = strings.TrimSpace(trimmed)
	}

	// In case there is still commentary before the first '{' or after the last '}'
	startIdx := strings.Index(trimmed, "{")
	endIdx := strings.LastIndex(trimmed, "}")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		trimmed = trimmed[startIdx : endIdx+1]
	}

	// Normalize byte slice and remove any non-printable leading runes
	byteData := []byte(trimmed)
	byteData = bytes.TrimSpace(byteData)

	return byteData
}

// Ping checks whether the configured model endpoint is responsive.
func (c *Client) Ping(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Respond with 'ok'",
			},
		},
		MaxTokens: 5,
	}

	_, err := c.apiClient.CreateChatCompletion(ctx, req)
	return err
}

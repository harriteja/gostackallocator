package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// OpenAIAdapter implements the AIClient interface using OpenAI's API
type OpenAIAdapter struct {
	client      *openai.Client
	model       string
	maxTokens   int
	temperature float32
	logger      *zap.Logger
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(apiKey, model string, maxTokens int, temperature float32, logger *zap.Logger) *OpenAIAdapter {
	if logger == nil {
		logger = zap.NewNop()
	}

	client := openai.NewClient(apiKey)

	return &OpenAIAdapter{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		logger:      logger,
	}
}

// SuggestFix generates a code suggestion using OpenAI's API
func (a *OpenAIAdapter) SuggestFix(ctx context.Context, snippet, issueMsg string) (string, error) {
	if a.client == nil {
		return "", fmt.Errorf("OpenAI client not initialized")
	}

	prompt := a.buildPrompt(snippet, issueMsg)

	// Create completion request
	req := openai.ChatCompletionRequest{
		Model:       a.model,
		MaxTokens:   a.maxTokens,
		Temperature: a.temperature,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a Go programming expert specializing in memory optimization and stack allocation. Provide concise, actionable code suggestions.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Make API call
	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		a.logger.Error("OpenAI API call failed", zap.Error(err))
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no suggestions returned from OpenAI")
	}

	suggestion := strings.TrimSpace(resp.Choices[0].Message.Content)

	a.logger.Debug("OpenAI suggestion generated",
		zap.String("issue", issueMsg),
		zap.String("suggestion", suggestion),
	)

	return suggestion, nil
}

// buildPrompt constructs the prompt for OpenAI
func (a *OpenAIAdapter) buildPrompt(snippet, issueMsg string) string {
	return fmt.Sprintf(`Analyze this Go code snippet and provide a specific suggestion to fix the memory allocation issue:

Issue: %s

Code:
%s

Please provide:
1. A brief explanation of the problem
2. A specific code change recommendation
3. Why this change improves memory allocation

Keep the response concise and focused on the specific issue.`, issueMsg, snippet)
}

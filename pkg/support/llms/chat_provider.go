// Package provider provides AI model provider implementations for the agent-go framework.
// This file contains the Service ChatProvider Interface (SPI) definitions for AI model providers.
package llms

import (
	"context"
	"fmt"
	"io"
	"iter"

	"github.com/oopslink/agent-go/pkg/commons/utils"
)

func OfProviderOptions(opts ...ProviderOption) *ProviderOptions {
	options := &ProviderOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// ProviderOption is a function that configures provider behavior.
type ProviderOption func(p *ProviderOptions)

// ProviderOptions holds configuration settings for provider instances.
type ProviderOptions struct {
	BaseUrl                 string // Base URL for the provider's API
	ApiKey                  string // API key for authentication
	SkipVerifySSL           bool   // Whether to skip SSL certificate verification
	Debug                   bool   // Whether to enable Debug logging
	OpenaiCompatibilityMode bool   // Whether to enable openai compatibility mode
}

func (o *ProviderOptions) String() string {
	return fmt.Sprintf("BaseUrl: %s, ApiKey: %s, SkipVerifySSL: %t, Debug: %t, OpenaiCompatibilityMode: %t",
		o.BaseUrl, utils.Sensitive(o.ApiKey, "***", 3, 3), o.SkipVerifySSL, o.Debug, o.OpenaiCompatibilityMode)
}

// WithBaseUrl sets the base URL for the provider's API.
func WithBaseUrl(url string) ProviderOption {
	return func(p *ProviderOptions) {
		p.BaseUrl = url
	}
}

// WithAPIKey sets the API key for provider authentication.
func WithAPIKey(apiKey string) ProviderOption {
	return func(p *ProviderOptions) {
		p.ApiKey = apiKey
	}
}

// SkipVerifySSL configures the provider to skip SSL certificate verification.
// This is useful for development or testing environments with self-signed certificates.
func SkipVerifySSL() ProviderOption {
	return func(p *ProviderOptions) {
		p.SkipVerifySSL = true
	}
}

// EnableDebug enables Debug logging for the provider.
func EnableDebug() ProviderOption {
	return func(p *ProviderOptions) {
		p.Debug = true
	}
}

// OpenAICompatibilityMode enables openai compatibility mode.
func OpenAICompatibilityMode() ProviderOption {
	return func(p *ProviderOptions) {
		p.OpenaiCompatibilityMode = true
	}
}

// ChatProvider is the main interface for AI model providers.
// It defines the contract that all provider implementations must fulfill.
type ChatProvider interface {
	io.Closer // Providers must be closable to free resources

	// NewChat creates a new chat session with the given system prompt and model.
	// The chat session can be used to send messages and receive responses.
	NewChat(systemPrompt string, model *Model) (Chat, error)

	// IsRetryableError determines if an error should trigger a retry.
	// This is used by retry mechanisms to decide whether to retry failed operations.
	IsRetryableError(error) bool
}

// ChatOption is a function that configures chat behavior.
type ChatOption func(opt *ChatOptions)

// ChatOptions holds configuration settings for chat sessions.
type ChatOptions struct {
	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will
	// make the output more random, while lower values like 0.2 will make it more
	// focused and deterministic. We generally recommend altering this or `top_p` but
	// not both.
	Temperature *float64
	// An alternative to sampling with temperature, called nucleus sampling, where the
	// model considers the results of the tokens with top_p probability mass. So 0.1
	// means only the tokens comprising the top 10% probability mass are considered.
	//
	// We generally recommend altering this or `temperature` but not both.
	TopP *float64
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on their
	// existing frequency in the text so far, decreasing the model's likelihood to
	// repeat the same line verbatim.
	FrequencyPenalty *float64
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on
	// whether they appear in the text so far, increasing the model's likelihood to
	// talk about new topics.
	PresencePenalty *float64
	// An upper bound for the number of tokens that can be generated for a completion,
	// including visible output tokens and reasoning tokens
	MaxCompletionTokens *int64
	// ReasoningEffort specifies the level of reasoning effort required
	ReasoningEffort ReasoningEffort

	// Tools defines the tools available for the chat session
	Tools []*ToolDescriptor

	// Streaming enables streaming responses from the model
	Streaming bool
}

// WithTemperature sets the sampling temperature for the chat session.
func WithTemperature(temperature float64) ChatOption {
	return func(p *ChatOptions) {
		p.Temperature = &temperature
	}
}

// WithTopP sets the nucleus sampling parameter for the chat session.
func WithTopP(topP float64) ChatOption {
	return func(p *ChatOptions) {
		p.TopP = &topP
	}
}

// WithFrequencyPenalty sets the frequency penalty for the chat session.
func WithFrequencyPenalty(frequencyPenalty float64) ChatOption {
	return func(p *ChatOptions) {
		p.FrequencyPenalty = &frequencyPenalty
	}
}

// WithPresencePenalty sets the presence penalty for the chat session.
func WithPresencePenalty(presencePenalty float64) ChatOption {
	return func(p *ChatOptions) {
		p.PresencePenalty = &presencePenalty
	}
}

// WithMaxCompletionTokens sets the maximum number of tokens for completion.
func WithMaxCompletionTokens(maxCompletionTokens int64) ChatOption {
	return func(p *ChatOptions) {
		p.MaxCompletionTokens = &maxCompletionTokens
	}
}

// WithStreaming enables or disables streaming responses.
func WithStreaming(streaming bool) ChatOption {
	return func(p *ChatOptions) {
		p.Streaming = streaming
	}
}

// WithReasoningEffort sets the reasoning effort level for the chat session.
func WithReasoningEffort(reasoningEffort ReasoningEffort) ChatOption {
	return func(p *ChatOptions) {
		p.ReasoningEffort = reasoningEffort
	}
}

// WithTools adds tools to the chat session.
func WithTools(tools ...*ToolDescriptor) ChatOption {
	return func(p *ChatOptions) {
		p.Tools = append(p.Tools, tools...)
	}
}

// Chat represents a chat session with an AI model.
// It provides methods for sending messages and receiving responses.
type Chat interface {
	// Send sends messages to the AI model and returns an iterator for responses.
	// The context can be used to cancel the operation.
	// Options can be used to configure the chat behavior.
	Send(ctx context.Context, messages []*Message, options ...ChatOption) (ChatResponseIterator, error)
}

// ChatResponseIterator is an iterator that yields chat responses and errors.
type ChatResponseIterator iter.Seq2[*ChatResponse, error]

// ChatResponse represents a response from the AI model.
type ChatResponse struct {
	Message                    // The response message
	Usage        UsageMetadata // Token usage information
	FinishReason FinishReason  // Why the response generation finished
}

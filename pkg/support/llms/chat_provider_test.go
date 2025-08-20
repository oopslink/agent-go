package llms

import (
	"errors"
	cerrors "github.com/oopslink/agent-go/pkg/commons/errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestProviderOptions(t *testing.T) {
	// Test WithBaseUrl
	opts := &ProviderOptions{}
	WithBaseUrl("https://api.example.com")(opts)
	if opts.BaseUrl != "https://api.example.com" {
		t.Errorf("WithBaseUrl: BaseUrl = %v, want https://api.example.com", opts.BaseUrl)
	}

	// Test WithAPIKey
	opts = &ProviderOptions{}
	WithAPIKey("test-api-key")(opts)
	if opts.ApiKey != "test-api-key" {
		t.Errorf("WithAPIKey: ApiKey = %v, want test-api-key", opts.ApiKey)
	}

	// Test SkipVerifySSL
	opts = &ProviderOptions{}
	SkipVerifySSL()(opts)
	if !opts.SkipVerifySSL {
		t.Error("SkipVerifySSL: SkipVerifySSL should be true")
	}

	// Test EnableDebug
	opts = &ProviderOptions{}
	EnableDebug()(opts)
	if !opts.Debug {
		t.Error("EnableDebug: Debug should be true")
	}
}

func TestChatOptions(t *testing.T) {
	// Test WithTemperature
	opts := &ChatOptions{}
	temperature := 0.7
	WithTemperature(temperature)(opts)
	if opts.Temperature == nil {
		t.Error("WithTemperature: Temperature should not be nil")
	}
	if *opts.Temperature != temperature {
		t.Errorf("WithTemperature: Temperature = %v, want %v", *opts.Temperature, temperature)
	}

	// Test WithTopP
	opts = &ChatOptions{}
	topP := 0.9
	WithTopP(topP)(opts)
	if opts.TopP == nil {
		t.Error("WithTopP: TopP should not be nil")
	}
	if *opts.TopP != topP {
		t.Errorf("WithTopP: TopP = %v, want %v", *opts.TopP, topP)
	}

	// Test WithFrequencyPenalty
	opts = &ChatOptions{}
	freqPenalty := 0.5
	WithFrequencyPenalty(freqPenalty)(opts)
	if opts.FrequencyPenalty == nil {
		t.Error("WithFrequencyPenalty: FrequencyPenalty should not be nil")
	}
	if *opts.FrequencyPenalty != freqPenalty {
		t.Errorf("WithFrequencyPenalty: FrequencyPenalty = %v, want %v", *opts.FrequencyPenalty, freqPenalty)
	}

	// Test WithPresencePenalty
	opts = &ChatOptions{}
	presencePenalty := 0.3
	WithPresencePenalty(presencePenalty)(opts)
	if opts.PresencePenalty == nil {
		t.Error("WithPresencePenalty: PresencePenalty should not be nil")
	}
	if *opts.PresencePenalty != presencePenalty {
		t.Errorf("WithPresencePenalty: PresencePenalty = %v, want %v", *opts.PresencePenalty, presencePenalty)
	}

	// Test WithMaxCompletionTokens
	opts = &ChatOptions{}
	maxTokens := int64(1000)
	WithMaxCompletionTokens(maxTokens)(opts)
	if opts.MaxCompletionTokens == nil {
		t.Error("WithMaxCompletionTokens: MaxCompletionTokens should not be nil")
	}
	if *opts.MaxCompletionTokens != maxTokens {
		t.Errorf("WithMaxCompletionTokens: MaxCompletionTokens = %v, want %v", *opts.MaxCompletionTokens, maxTokens)
	}

	// Test WithStreaming
	opts = &ChatOptions{}
	WithStreaming(true)(opts)
	if !opts.Streaming {
		t.Error("WithStreaming: Streaming should be true")
	}

	// Test WithReasoningEffort
	opts = &ChatOptions{}
	reasoningEffort := ReasoningEffortHigh
	WithReasoningEffort(reasoningEffort)(opts)
	if opts.ReasoningEffort != reasoningEffort {
		t.Errorf("WithReasoningEffort: ReasoningEffort = %v, want %v", opts.ReasoningEffort, reasoningEffort)
	}

	// Test WithTools
	opts = &ChatOptions{}
	tool1 := &ToolDescriptor{Name: "tool1"}
	tool2 := &ToolDescriptor{Name: "tool2"}
	WithTools(tool1, tool2)(opts)
	if len(opts.Tools) != 2 {
		t.Errorf("WithTools: len(Tools) = %d, want 2", len(opts.Tools))
	}
	if opts.Tools[0] != tool1 {
		t.Errorf("WithTools: Tools[0] = %v, want %v", opts.Tools[0], tool1)
	}
	if opts.Tools[1] != tool2 {
		t.Errorf("WithTools: Tools[1] = %v, want %v", opts.Tools[1], tool2)
	}
}

func TestAPIError(t *testing.T) {
	// Test APIError with original error
	originalErr := errors.New("original error")
	apiErr := &cerrors.APIError{
		StatusCode: http.StatusTooManyRequests,
		Message:    "Rate limit exceeded",
		Err:        originalErr,
	}

	expectedMsg := "API Error: Status=429, Message='Rate limit exceeded', OriginalErr=original error"
	if apiErr.Error() != expectedMsg {
		t.Errorf("APIError.Error() = %v, want %v", apiErr.Error(), expectedMsg)
	}

	if !errors.Is(originalErr, apiErr.Err) {
		t.Errorf("APIError.Unwrap() = %v, want %v", apiErr.Err, originalErr)
	}

	// Test APIError without original error
	apiErr = &cerrors.APIError{
		StatusCode: http.StatusBadRequest,
		Message:    "Invalid request",
	}

	expectedMsg = "API Error: Status=400, Message='Invalid request'"
	if apiErr.Error() != expectedMsg {
		t.Errorf("APIError.Error() = %v, want %v", apiErr.Error(), expectedMsg)
	}

	if apiErr.Err != nil {
		t.Errorf("APIError.Unwrap() = %v, want nil", apiErr.Err)
	}
}

func TestIsRetryableError(t *testing.T) {
	// Test nil error
	if cerrors.IsRetryableError(nil) {
		t.Error("IsRetryableError(nil) should be false")
	}

	// Test retryable HTTP status codes
	retryableStatusCodes := []int{
		http.StatusConflict,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}

	for _, statusCode := range retryableStatusCodes {
		apiErr := &cerrors.APIError{StatusCode: statusCode}
		if !cerrors.IsRetryableError(apiErr) {
			t.Errorf("IsRetryableError should be true for status code %d", statusCode)
		}
	}

	// Test non-retryable HTTP status codes
	nonRetryableStatusCodes := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
	}

	for _, statusCode := range nonRetryableStatusCodes {
		apiErr := &cerrors.APIError{StatusCode: statusCode}
		if cerrors.IsRetryableError(apiErr) {
			t.Errorf("IsRetryableError should be false for status code %d", statusCode)
		}
	}

	// Test network timeout error
	timeoutErr := &net.DNSError{
		Err:        "timeout",
		Name:       "example.com",
		IsTimeout:  true,
		IsNotFound: false,
	}
	if !cerrors.IsRetryableError(timeoutErr) {
		t.Error("IsRetryableError should be true for network timeout error")
	}

	// Test non-timeout network error
	nonTimeoutErr := &net.DNSError{
		Err:        "not found",
		Name:       "example.com",
		IsTimeout:  false,
		IsNotFound: true,
	}
	if cerrors.IsRetryableError(nonTimeoutErr) {
		t.Error("IsRetryableError should be false for non-timeout network error")
	}

	// Test other error types
	otherErr := errors.New("some other error")
	if cerrors.IsRetryableError(otherErr) {
		t.Error("IsRetryableError should be false for other error types")
	}
}

func TestUsageMetadata(t *testing.T) {
	usage := UsageMetadata{
		InputTokens:         100,
		OutputTokens:        50,
		CacheCreationTokens: 10,
		CacheReadTokens:     5,
	}

	if usage.InputTokens != 100 {
		t.Errorf("UsageMetadata.InputTokens = %d, want 100", usage.InputTokens)
	}

	if usage.OutputTokens != 50 {
		t.Errorf("UsageMetadata.OutputTokens = %d, want 50", usage.OutputTokens)
	}

	if usage.CacheCreationTokens != 10 {
		t.Errorf("UsageMetadata.CacheCreationTokens = %d, want 10", usage.CacheCreationTokens)
	}

	if usage.CacheReadTokens != 5 {
		t.Errorf("UsageMetadata.CacheReadTokens = %d, want 5", usage.CacheReadTokens)
	}
}

func TestChatResponse(t *testing.T) {
	message := &Message{
		MessageId: "msg-123",
		Creator: MessageCreator{
			Role: MessageRoleAssistant,
		},
		Parts: []Part{
			&TextPart{Text: "Hello, world!"},
		},
		Timestamp: time.Now(),
	}

	usage := UsageMetadata{
		InputTokens:  100,
		OutputTokens: 50,
	}

	response := &ChatResponse{
		Message:      *message,
		Usage:        usage,
		FinishReason: FinishReasonNormalEnd,
	}

	if response.MessageId != "msg-123" {
		t.Errorf("ChatResponse.MessageId = %s, want msg-123", response.MessageId)
	}

	if response.Usage.InputTokens != 100 {
		t.Errorf("ChatResponse.Usage.InputTokens = %d, want 100", response.Usage.InputTokens)
	}

	if response.FinishReason != FinishReasonNormalEnd {
		t.Errorf("ChatResponse.FinishReason = %s, want %s", response.FinishReason, FinishReasonNormalEnd)
	}
}

func TestMultipleChatOptions(t *testing.T) {
	opts := &ChatOptions{}

	// Apply multiple options
	WithTemperature(0.8)(opts)
	WithTopP(0.9)(opts)
	WithStreaming(true)(opts)
	WithReasoningEffort(ReasoningEffortMedium)(opts)

	// Verify all options were applied
	if opts.Temperature == nil || *opts.Temperature != 0.8 {
		t.Error("Temperature should be 0.8")
	}

	if opts.TopP == nil || *opts.TopP != 0.9 {
		t.Error("TopP should be 0.9")
	}

	if !opts.Streaming {
		t.Error("Streaming should be true")
	}

	if opts.ReasoningEffort != ReasoningEffortMedium {
		t.Error("ReasoningEffort should be medium")
	}
}

func TestMultipleProviderOptions(t *testing.T) {
	opts := &ProviderOptions{}

	// Apply multiple options
	WithBaseUrl("https://api.example.com")(opts)
	WithAPIKey("test-key")(opts)
	SkipVerifySSL()(opts)
	EnableDebug()(opts)

	// Verify all options were applied
	if opts.BaseUrl != "https://api.example.com" {
		t.Error("BaseUrl should be https://api.example.com")
	}

	if opts.ApiKey != "test-key" {
		t.Error("ApiKey should be test-key")
	}

	if !opts.SkipVerifySSL {
		t.Error("SkipVerifySSL should be true")
	}

	if !opts.Debug {
		t.Error("Debug should be true")
	}
}

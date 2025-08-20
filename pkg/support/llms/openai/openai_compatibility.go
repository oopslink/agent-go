package openai

import (
	"context"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

const (
	ModelProviderOpenAICompatibility = "openai-compatibility"
)

func init() {
	_ = llms.RegisterChatProvider(ModelProviderOpenAICompatibility, newCompatibilityChatProvider)
}

func newCompatibilityChatProvider(opts ...llms.ProviderOption) (llms.ChatProvider, error) {
	openaiProvider, err := newChatProvider(opts...)
	if err != nil {
		return nil, err
	}
	return &openAICompatibilityChatProvider{
		openaiProvider: openaiProvider,
	}, nil
}

type openAICompatibilityChat struct {
	model      *llms.Model
	openaiChat llms.Chat
}

func (c *openAICompatibilityChat) Send(ctx context.Context, messages []*llms.Message, options ...llms.ChatOption) (llms.ChatResponseIterator, error) {
	options = c.compatibilityOptions(options)
	return c.openaiChat.Send(ctx, messages, options...)
}

func (c *openAICompatibilityChat) compatibilityOptions(options []llms.ChatOption) []llms.ChatOption {
	// For now, just return the options unchanged.
	return options
}

type openAICompatibilityChatProvider struct {
	openaiProvider llms.ChatProvider
}

func (p *openAICompatibilityChatProvider) Close() error {
	return p.openaiProvider.Close()
}

func (p *openAICompatibilityChatProvider) IsRetryableError(err error) bool {
	return p.openaiProvider.IsRetryableError(err)
}

func (p *openAICompatibilityChatProvider) NewChat(systemPrompt string, model *llms.Model) (llms.Chat, error) {
	m := *model

	if m.ModelId.Provider != ModelProviderOpenAI {
		// If the model provider is not openai,
		// leave the features unchanged, remove others.
		m.Features = append(m.Features, llms.ModelFeatureAttachment)
		m.Features = append(m.Features, llms.ModelFeatureCompletion)
	}
	openaiChat, err := p.openaiProvider.NewChat(systemPrompt, &m)
	if err != nil {
		return nil, err
	}
	return &openAICompatibilityChat{
		model:      &m,
		openaiChat: openaiChat,
	}, nil
}

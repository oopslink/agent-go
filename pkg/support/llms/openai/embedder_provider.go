package openai

import (
	"context"

	"github.com/openai/openai-go"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func init() {
	_ = llms.RegisterEmbedderProvider(ModelProviderOpenAI, newEmbedderProvider)
}

// openAIEmbedderProvider implements EmbedderProvider for Gemini
func newEmbedderProvider(model *llms.Model, opts ...llms.ProviderOption) (llms.EmbedderProvider, error) {
	options := llms.OfProviderOptions(opts...)
	if !model.IsSupport(llms.ModelFeatureEmbedding) {
		return nil, errors.Errorf(llms.ErrorCodeCreateEmbedderProviderFailed,
			"model %s is not support embedding", model.ModelId.String())
	}
	client, err := createOpenAIClient(options)
	if err != nil {
		return nil, err
	}
	return &openAIEmbedderProvider{
		client: client,
		model:  model,
	}, nil
}

type openAIEmbedderProvider struct {
	client openai.Client
	model  *llms.Model
}

func (o *openAIEmbedderProvider) GetEmbeddings(ctx context.Context, texts []string) (*llms.EmbeddingResponse, error) {
	params := openai.EmbeddingNewParams{
		Model:          o.model.ApiModelName,
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
	}
	response, err := o.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, err
	}

	var vectors []llms.FloatVector
	for idx := range response.Data {
		embedding := response.Data[idx]
		vectors = append(vectors, embedding.Embedding)
	}

	result := &llms.EmbeddingResponse{
		Model:   o.model,
		Usage:   o.getUsageStats(&response.Usage),
		Vectors: vectors,
	}
	return result, nil
}

func (o *openAIEmbedderProvider) getUsageStats(usage *openai.CreateEmbeddingResponseUsage) llms.UsageMetadata {
	if usage == nil {
		return llms.UsageMetadata{}
	}
	return llms.UsageMetadata{
		InputTokens:         usage.PromptTokens,
		OutputTokens:        usage.TotalTokens,
		CacheReadTokens:     0,
		CacheCreationTokens: 0,
	}
}

// Package gemini provides Google Gemini AI model provider implementations for the agent-go framework.
// This file contains the embedder provider implementation for Google's Gemini models.
package gemini

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"os"

	"google.golang.org/genai"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// init registers the Gemini embedder provider with the global provider registry.
// This function is called automatically when the package is imported.
func init() {
	_ = llms.RegisterEmbedderProvider(ModelProviderGemini, newEmbedderProvider)
}

// newEmbedderProvider creates a new Gemini embedder provider instance.
// It configures the provider with API key, project settings, and other options.
func newEmbedderProvider(model *llms.Model, opts ...llms.ProviderOption) (llms.EmbedderProvider, error) {
	options := llms.OfProviderOptions(opts...)
	// Create client config with Google Cloud project settings
	config := &genai.ClientConfig{
		Project: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Backend: genai.BackendGeminiAPI,
	}

	// Set API key from options or environment variables
	apiKey := options.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey != "" {
		config.APIKey = apiKey
	}

	// Create Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Errorf(llms.ErrorCodeCreateEmbedderProviderFailed,
			"failed to create Gemini client: %s", err.Error())
	}

	if options.Debug {
		journal.Info("provider/embedding", "gemini",
			fmt.Sprintf("[DEBUG] GeminiEmbedderProvider params: ApiKey=****, clientConfig=%v, SkipVerifySSL=%v", config, options.SkipVerifySSL))
	}

	return &geminiEmbedderProvider{
		client: client,
		model:  model,
		debug:  options.Debug,
	}, nil
}

// geminiEmbedderProvider implements the EmbedderProvider interface for Gemini models.
type geminiEmbedderProvider struct {
	client *genai.Client
	model  *llms.Model
	debug  bool
}

// GetEmbeddings generates embeddings for the provided texts using the configured Gemini model.
// It processes each text individually and returns a combined response with all vectors.
func (g *geminiEmbedderProvider) GetEmbeddings(ctx context.Context, texts []string) (*llms.EmbeddingResponse, error) {
	// Check if model is provided
	if g.model == nil {
		return nil, errors.Errorf(llms.ErrorCodeEmbeddingSessionFailed,
			"model is required for Gemini embedder provider")
	}

	// Validate that the model supports embedding
	if !g.model.IsSupport(llms.ModelFeatureEmbedding) {
		return nil, errors.Errorf(llms.ErrorCodeEmbeddingSessionFailed,
			"model %s does not support embedding, features: [%v]", g.model.ModelId.String(), g.model.Features)
	}

	var vectors []llms.FloatVector
	var totalInputTokens, totalOutputTokens int64

	// Process each text individually
	for _, text := range texts {
		// Create content for embedding using Gemini's embedContent API
		contents := []*genai.Content{
			genai.NewContentFromText(text, genai.RoleUser),
		}

		// Call Gemini's embedContent API
		result, err := g.client.Models.EmbedContent(ctx, g.model.ApiModelName, contents, nil)
		if err != nil {
			return nil, errors.Errorf(llms.ErrorCodeEmbeddingSessionFailed,
				"failed to get embedding for text: %s", err.Error())
		}

		if g.debug {
			journal.Info("provider/embedding", "gemini",
				fmt.Sprintf("[DEBUG] Gemini embedding response for text: %s", text[:utils.MinInt(50, len(text))]))
		}

		// Extract embedding vector from the response
		if len(result.Embeddings) > 0 {
			embedding := result.Embeddings[0]
			if embedding.Values != nil {
				// Convert float32 to float64 for consistency with other providers
				vector := make([]float64, len(embedding.Values))
				for i, v := range embedding.Values {
					vector[i] = float64(v)
				}
				vectors = append(vectors, vector)
			} else {
				return nil, errors.Errorf(llms.ErrorCodeEmbeddingSessionFailed,
					"no embedding values returned for text")
			}
		} else {
			return nil, errors.Errorf(llms.ErrorCodeEmbeddingSessionFailed,
				"no embeddings returned for text")
		}

		// Update token counts (approximate since Gemini might not provide detailed usage)
		totalInputTokens += int64(len(text) / 4) // Rough estimate
		totalOutputTokens += int64(len(result.Embeddings[0].Values))
	}

	usage := llms.UsageMetadata{
		InputTokens:  totalInputTokens,
		OutputTokens: totalOutputTokens,
	}
	result := &llms.EmbeddingResponse{
		Model:   g.model,
		Usage:   usage,
		Vectors: vectors,
	}

	journal.AccumulateUsage("embedding", usage.AsMap())

	return result, nil
}

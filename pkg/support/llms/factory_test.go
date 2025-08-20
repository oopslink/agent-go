package llms

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_AddModel(t *testing.T) {
	reg := &registry{}

	// Test adding a new model
	model := &Model{
		ModelId: ModelId{
			Provider: "test-provider",
			ID:       "test-model",
		},
		Name: "Test Model",
	}

	err := reg.AddModel(model)
	assert.NoError(t, err)

	// Test adding the same model again (should fail)
	err = reg.AddModel(model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model already exists")
}

func TestRegistry_GetModel(t *testing.T) {
	reg := &registry{}

	// Test getting a non-existent model
	modelId := ModelId{
		Provider: "test-provider",
		ID:       "non-existent",
	}

	model, found := reg.GetModel(modelId)
	assert.False(t, found)
	assert.NotNil(t, model)
	assert.Equal(t, modelId, model.ModelId)
	assert.Equal(t, modelId.ID, model.Name)
	assert.Contains(t, model.Features, ModelFeatureEmbedding)
	assert.Contains(t, model.Features, ModelFeatureCompletion)

	// Test getting an existing model
	existingModel := &Model{
		ModelId: ModelId{
			Provider: "test-provider",
			ID:       "existing-model",
		},
		Name: "Existing Model",
	}

	err := reg.AddModel(existingModel)
	require.NoError(t, err)

	retrievedModel, found := reg.GetModel(existingModel.ModelId)
	assert.True(t, found)
	assert.Equal(t, existingModel, retrievedModel)
}

func TestRegistry_AddChatProvider(t *testing.T) {
	reg := &registry{}

	// Test adding a chat provider
	constructor := func(opts ...ProviderOption) (ChatProvider, error) {
		return &mockChatProvider{}, nil
	}

	err := reg.AddChatProvider("test-provider", constructor)
	assert.NoError(t, err)

	// Test adding the same provider again (should fail)
	err = reg.AddChatProvider("test-provider", constructor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat provider already registered")
}

func TestRegistry_AddEmbedderProvider(t *testing.T) {
	reg := &registry{}

	// Test adding an embedder provider
	constructor := func(model *Model, opts ...ProviderOption) (EmbedderProvider, error) {
		return &mockEmbedderProvider{}, nil
	}

	err := reg.AddEmbedderProvider("test-provider", constructor)
	assert.NoError(t, err)

	// Test adding the same provider again (should fail)
	err = reg.AddEmbedderProvider("test-provider", constructor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "embedder provider already registered")
}

func TestRegistry_NewChatProvider(t *testing.T) {
	reg := &registry{}

	// Test creating a provider that doesn't exist
	_, err := reg.NewChatProvider("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat provider not registered")

	// Test creating a registered provider
	constructor := func(opts ...ProviderOption) (ChatProvider, error) {
		return &mockChatProvider{}, nil
	}

	err = reg.AddChatProvider("test-provider", constructor)
	require.NoError(t, err)

	provider, err := reg.NewChatProvider("test-provider")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestRegistry_NewEmbedderProvider(t *testing.T) {
	reg := &registry{}

	model := &Model{
		ModelId: ModelId{
			Provider: "test-provider",
			ID:       "test-model",
		},
		Name: "Test Model",
	}

	// Test creating a provider that doesn't exist
	_, err := reg.NewEmbedderProvider("non-existent", model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "embedder provider not registered")

	// Test creating a registered provider
	constructor := func(model *Model, opts ...ProviderOption) (EmbedderProvider, error) {
		return &mockEmbedderProvider{}, nil
	}

	err = reg.AddEmbedderProvider("test-provider", constructor)
	require.NoError(t, err)

	provider, err := reg.NewEmbedderProvider("test-provider", model)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestRegistry_ThreadSafety(t *testing.T) {
	reg := &registry{}

	// Test concurrent model additions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			model := &Model{
				ModelId: ModelId{
					Provider: "test-provider",
					ID:       fmt.Sprintf("model-%d", id),
				},
				Name: fmt.Sprintf("Model %d", id),
			}
			reg.AddModel(model)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all models were added
	assert.Len(t, reg.models, 10)
}

func TestGlobalFunctions(t *testing.T) {
	// Test RegisterModel and GetModel
	model := &Model{
		ModelId: ModelId{
			Provider: "global-test",
			ID:       "test-model",
		},
		Name: "Global Test Model",
	}

	err := RegisterModel(model)
	assert.NoError(t, err)

	retrievedModel, found := GetModel(model.ModelId)
	assert.True(t, found)
	assert.Equal(t, model, retrievedModel)

	// Test RegisterChatProvider and NewChatProvider
	constructor := func(opts ...ProviderOption) (ChatProvider, error) {
		return &mockChatProvider{}, nil
	}

	err = RegisterChatProvider("global-test", constructor)
	assert.NoError(t, err)

	provider, err := NewChatProvider("global-test")
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Test RegisterEmbedderProvider and NewEmbedderProvider
	embedderConstructor := func(model *Model, opts ...ProviderOption) (EmbedderProvider, error) {
		return &mockEmbedderProvider{}, nil
	}

	err = RegisterEmbedderProvider("global-test", embedderConstructor)
	assert.NoError(t, err)

	embedderProvider, err := NewEmbedderProvider("global-test", model)
	assert.NoError(t, err)
	assert.NotNil(t, embedderProvider)
}

// Mock implementations for testing
type mockChatProvider struct{}

func (m *mockChatProvider) Close() error {
	return nil
}

func (m *mockChatProvider) NewChat(systemPrompt string, model *Model) (Chat, error) {
	return &mockChat{}, nil
}

func (m *mockChatProvider) IsRetryableError(err error) bool {
	return false
}

type mockChat struct{}

func (m *mockChat) Send(ctx context.Context, messages []*Message, options ...ChatOption) (ChatResponseIterator, error) {
	return nil, nil
}

type mockEmbedderProvider struct{}

func (m *mockEmbedderProvider) GetEmbeddings(ctx context.Context, texts []string) (*EmbeddingResponse, error) {
	return &EmbeddingResponse{
		Model:   &Model{},
		Vectors: []FloatVector{},
		Usage:   UsageMetadata{},
	}, nil
}

// Package provider provides AI model provider implementations for the agent-go framework.
// This file contains the factory pattern implementation for managing model providers and registrations.
package llms

import (
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"sync"
)

// _registry is the global registry instance for managing models and providers.
var _registry = &registry{}

// GetModel retrieves a model by its ID from the global registry.
// Returns the model and a boolean indicating if it was found.
func GetModel(modelId ModelId) (*Model, bool) {
	return _registry.GetModel(modelId)
}

// RegisterModel adds a new model to the global registry.
// Returns an error if a model with the same ID already exists.
func RegisterModel(model *Model) error {
	return _registry.AddModel(model)
}

// ChatProviderConstructor is a function type that creates new chat provider instances.
// It takes provider options and returns a ChatProvider and potential error.
type ChatProviderConstructor func(opts ...ProviderOption) (ChatProvider, error)

// EmbedderProviderConstructor is a function type that creates new embedder provider instances.
// It takes a model and provider options, returning an EmbedderProvider and potential error.
type EmbedderProviderConstructor func(model *Model, opts ...ProviderOption) (EmbedderProvider, error)

// RegisterChatProvider registers a chat provider constructor for a given provider name.
// This allows the factory to create instances of the registered provider type.
func RegisterChatProvider(providerName ModelProvider, constructor ChatProviderConstructor) error {
	return _registry.AddChatProvider(providerName, constructor)
}

// RegisterEmbedderProvider registers an embedder provider constructor for a given provider name.
// This allows the factory to create instances of the registered provider type.
func RegisterEmbedderProvider(providerName ModelProvider, constructor EmbedderProviderConstructor) error {
	return _registry.AddEmbedderProvider(providerName, constructor)
}

// NewChatProvider creates a new chat provider instance for the specified provider name.
// Returns an error if the provider is not registered.
func NewChatProvider(providerName ModelProvider, opts ...ProviderOption) (ChatProvider, error) {
	return _registry.NewChatProvider(providerName, opts...)
}

// NewEmbedderProvider creates a new embedder provider instance for the specified provider name.
// Returns an error if the provider is not registered.
func NewEmbedderProvider(providerName ModelProvider, model *Model, opts ...ProviderOption) (EmbedderProvider, error) {
	return _registry.NewEmbedderProvider(providerName, model, opts...)
}

// registry manages the registration and instantiation of models and providers.
// It provides thread-safe access to registered models, chat providers, and embedder providers.
type registry struct {
	lock              sync.RWMutex
	models            []*Model
	chatProviders     map[ModelProvider]ChatProviderConstructor
	embedderProviders map[ModelProvider]EmbedderProviderConstructor
}

// AddModel adds a new model to the registry.
// Uses write lock to ensure thread safety.
// Returns an error if a model with the same ID already exists.
func (r *registry) AddModel(m *Model) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	model := r.findModel(m.ModelId)
	if model != nil {
		return errors.Errorf(ErrorCodeModelAlreadyExists, "model already exists: %v", m.ModelId)
	}

	r.models = append(r.models, m)
	return nil
}

// GetModel retrieves a model by its ID from the registry.
// Uses read lock for thread-safe access.
// If the model is not found, returns a default model with basic features.
func (r *registry) GetModel(modelId ModelId) (*Model, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	model := r.findModel(modelId)
	if model == nil {
		return &Model{
			ModelId:            modelId,
			Name:               modelId.ID,
			CostPer1MIn:        0,
			CostPer1MOut:       0,
			CostPer1MInCached:  0,
			CostPer1MOutCached: 0,
			ContextWindowSize:  0,
			Features: []ModelFeature{
				ModelFeatureEmbedding,
				ModelFeatureCompletion,
			},
		}, false
	}

	return model, true
}

// findModel searches for a model by its ID within the registry.
// This is an internal helper method that does not use locks.
func (r *registry) findModel(modelId ModelId) *Model {
	for _, m := range r.models {
		if m.ModelId.Provider == modelId.Provider && m.ModelId.ID == modelId.ID {
			return m
		}
	}
	return nil
}

// AddChatProvider registers a chat provider constructor for the given provider name.
// Uses write lock to ensure thread safety.
// Returns an error if the provider is already registered.
func (r *registry) AddChatProvider(providerName ModelProvider, constructor ChatProviderConstructor) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.chatProviders == nil {
		r.chatProviders = make(map[ModelProvider]ChatProviderConstructor)
	}

	if _, exists := r.chatProviders[providerName]; exists {
		return errors.Errorf(ErrorCodeChatProviderAlreadyExists,
			"chat provider already registered: %s", providerName)
	}

	r.chatProviders[providerName] = constructor
	return nil
}

// AddEmbedderProvider registers an embedder provider constructor for the given provider name.
// Uses write lock to ensure thread safety.
// Returns an error if the provider is already registered.
func (r *registry) AddEmbedderProvider(providerName ModelProvider, constructor EmbedderProviderConstructor) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.embedderProviders == nil {
		r.embedderProviders = make(map[ModelProvider]EmbedderProviderConstructor)
	}

	if _, exists := r.embedderProviders[providerName]; exists {
		return errors.Errorf(ErrorCodeEmbedderProviderAlreadyExists,
			"embedder provider already registered: %s", providerName)
	}

	r.embedderProviders[providerName] = constructor
	return nil
}

// NewChatProvider creates a new chat provider instance using the registered constructor.
// Uses read lock for thread-safe access.
// Returns an error if the provider is not registered.
func (r *registry) NewChatProvider(providerName ModelProvider, opts ...ProviderOption) (ChatProvider, error) {
	r.lock.RLock()
	constructor, exists := r.chatProviders[providerName]
	r.lock.RUnlock()

	if !exists {
		return nil, errors.Errorf(ErrorCodeChatProviderNotFound,
			"chat provider not registered: %s", providerName)
	}

	return constructor(opts...)
}

// NewEmbedderProvider creates a new embedder provider instance using the registered constructor.
// Uses read lock for thread-safe access.
// Returns an error if the provider is not registered.
func (r *registry) NewEmbedderProvider(providerName ModelProvider, model *Model, opts ...ProviderOption) (EmbedderProvider, error) {
	r.lock.RLock()
	constructor, exists := r.embedderProviders[providerName]
	r.lock.RUnlock()

	if !exists {
		return nil, errors.Errorf(ErrorCodeEmbedderProviderNotFound,
			"embedder provider not registered: %s", providerName)
	}

	return constructor(model, opts...)
}

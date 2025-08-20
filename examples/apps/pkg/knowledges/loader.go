package knowledges

import (
	"context"
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oopslink/agent-go-apps/pkg/config"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/support/document"
	"github.com/oopslink/agent-go/pkg/support/embedder"
	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/oopslink/agent-go/pkg/support/vectordb"
	"github.com/oopslink/agent-go/pkg/support/vectordb/milvus"
)

var (
	ErrorCodeCreateVectorDatabaseFailed = errors.ErrorCode{
		Code:           92000,
		Name:           "CreateVectorDatabaseFailed",
		DefaultMessage: "Failed to create vector database",
	}
	ErrorCodeCreateEmbedderFailed = errors.ErrorCode{
		Code:           92001,
		Name:           "CreateVectorDatabaseFailed",
		DefaultMessage: "Failed to create vector database",
	}
	ErrorCodeCreateKnowledgeBaseFailed = errors.ErrorCode{
		Code:           92002,
		Name:           "CreateKnowledgeBaseFailed",
		DefaultMessage: "Failed to create knowledge base",
	}
	ErrorCodeSaveKnowledgeStateFailed = errors.ErrorCode{
		Code:           92002,
		Name:           "SaveKnowledgeStateFailed",
		DefaultMessage: "Failed to save knowledge state",
	}
)

type KnowledgeData struct {
	KnowledgeBases []KnowledgeBaseData `json:"knownedge_bases"`
}

type KnowledgeBaseData struct {
	Name        string   `json:"name"`
	Domain      string   `json:"domain"`
	Description string   `json:"description"`
	Items       []string `json:"items"`
}

// KnowledgeState represents the loading state of knowledge bases
type KnowledgeState struct {
	LoadedKnowledgeBases map[string]KnowledgeBaseState `json:"loaded_knowledge_bases"`
	LastLoadTime         time.Time                     `json:"last_load_time"`
	DataHash             string                        `json:"data_hash"`
}

// KnowledgeBaseState represents the state of a single knowledge base
type KnowledgeBaseState struct {
	Name         string    `json:"name"`
	ItemCount    int       `json:"item_count"`
	LoadTime     time.Time `json:"load_time"`
	CollectionID string    `json:"collection_id"`
}

//go:embed data.json
var _data []byte

// LoadKnowledgeBases loads knowledge bases from the JSON file using vector database storage
func LoadKnowledgeBases(cfg *config.Config) ([]knowledge.KnowledgeBase, error) {
	fmt.Println("# Loading knowledge bases from state...")

	// Parse JSON
	var knowledgeData KnowledgeData
	if err := json.Unmarshal(_data, &knowledgeData); err != nil {
		return nil, errors.Errorf(ErrorCodeCreateKnowledgeBaseFailed,
			"failed to parse knowledge data: %v", err)
	}

	// Calculate data hash to detect changes
	dataHash := calculateDataHash(_data)

	// Get runtime directory and state file path
	runtimeDir := cfg.RuntimeDir
	if runtimeDir == "" {
		runtimeDir = "./runtime"
	}
	stateFilePath := filepath.Join(runtimeDir, "knowledge_state.json")

	// Load existing state
	existingState := loadKnowledgeState(stateFilePath)

	// Check if we need to reload
	needsReload := shouldReloadKnowledgeBases(existingState, dataHash, knowledgeData.KnowledgeBases)

	// Create vector database
	vectorDB, err := createVectorDB(cfg.VectorDB)
	if err != nil {
		return nil, errors.Wrap(ErrorCodeCreateKnowledgeBaseFailed, err)
	}

	// Create embedder
	embedder, err := createEmbedder(cfg.Embedder)
	if err != nil {
		return nil, errors.Wrap(ErrorCodeCreateKnowledgeBaseFailed, err)
	}

	// Create knowledge bases
	var knowledgeBases []knowledge.KnowledgeBase
	factory := knowledge.NewBaseKnowledgeItemFactory()

	// Initialize new state
	newState := &KnowledgeState{
		LoadedKnowledgeBases: make(map[string]KnowledgeBaseState),
		LastLoadTime:         time.Now(),
		DataHash:             dataHash,
	}

	for _, kbData := range knowledgeData.KnowledgeBases {
		// Create metadata
		metadata := knowledge.NewKnowledgeBaseMetadata(
			kbData.Name,
			kbData.Description,
			[]string{kbData.Domain},
			map[string]string{"source": "demo_data"},
		)

		fmt.Printf("# Creating knowledge base: %s (domain: %s)\n", kbData.Name, kbData.Domain)
		fmt.Printf("  - Description: %s\n", kbData.Description)

		// Create vector database storage
		collectionName := fmt.Sprintf("demo_%s", utils.HashString(kbData.Name))
		storage := knowledge.NewVectorDBStorage(collectionName, vectorDB, embedder)

		// Create knowledge base
		kb := knowledge.NewKnowledgeBase(storage, factory, metadata)

		// Check if this knowledge base needs to be loaded
		existingKbState, exists := existingState.LoadedKnowledgeBases[kbData.Name]
		shouldLoadItems := needsReload || !exists || existingKbState.ItemCount != len(kbData.Items)

		if shouldLoadItems {
			fmt.Printf("  - Loading items for knowledge base: %s\n", kbData.Name)
			// Add items to the knowledge base
			for idx, item := range kbData.Items {
				hash := utils.HashString(item)
				doc := &document.Document{
					Id:       document.DocumentId(fmt.Sprintf("%s_%s", kbData.Name, hash)),
					Name:     fmt.Sprintf("%s_%d", kbData.Name, idx),
					Content:  item,
					Metadata: map[string]any{"domain": kbData.Domain},
				}

				fmt.Println("  - Adding item ...")
				fmt.Printf("    -   ID: %s\n", doc.Id)
				fmt.Printf("    - Name: %s\n", doc.Name)
				// Create knowledge item
				knowledgeItem := factory.NewKnowledgeItem(doc)

				// Add to knowledge base
				if err = kb.AddItem(context.Background(), knowledgeItem); err != nil {
					return nil, errors.Wrap(ErrorCodeCreateKnowledgeBaseFailed,
						fmt.Errorf("failed to add item to knowledge base %s: %v", kbData.Name, err))
				}
			}
		} else {
			fmt.Printf("  - Knowledge base '%s' already loaded, skipping item loading\n", kbData.Name)
		}

		// Update state for this knowledge base
		newState.LoadedKnowledgeBases[kbData.Name] = KnowledgeBaseState{
			Name:         kbData.Name,
			ItemCount:    len(kbData.Items),
			LoadTime:     time.Now(),
			CollectionID: collectionName,
		}

		knowledgeBases = append(knowledgeBases, kb)
	}

	// Save state
	if err = saveKnowledgeState(stateFilePath, newState); err != nil {
		fmt.Printf("Warning: failed to save knowledge state: %v\n", err)
	}

	return knowledgeBases, nil
}

// createVectorDB creates a vector database based on configuration
func createVectorDB(cfg config.VectorDBConfig) (vectordb.VectorDB, error) {
	switch cfg.Type {
	case "milvus":
		// Create Milvus client config
		// Use default credentials if not provided
		username := cfg.Username
		password := cfg.Password
		if username == "" {
			username = "root"
		}
		if password == "" {
			password = "Milvus"
		}

		milvusConfig, err := milvus.NewClientConfig(
			cfg.Endpoint,
			username,
			password,
		)
		if err != nil {
			return nil, errors.Errorf(ErrorCodeCreateVectorDatabaseFailed,
				"failed to create milvus client config: %v", err)
		}

		// Create Milvus store
		vectorDB, err := milvus.New(context.Background(), *milvusConfig)
		if err != nil {
			return nil, errors.Errorf(ErrorCodeCreateVectorDatabaseFailed,
				"failed to create milvus store: %v", err)
		}

		return vectorDB, nil
	default:
		return nil, errors.Errorf(ErrorCodeCreateVectorDatabaseFailed,
			"unsupported vector database type: %s", cfg.Type)
	}
}

// createEmbedder creates an embedder based on configuration
func createEmbedder(cfg config.EmbedderConfig) (embedder.Embedder, error) {
	switch cfg.Type {
	case "llm":
		// Create embedder provider
		var providerOpts []llms.ProviderOption
		providerOpts = append(providerOpts, llms.WithAPIKey(cfg.APIKey))
		if cfg.BaseURL != "" {
			providerOpts = append(providerOpts, llms.WithBaseUrl(cfg.BaseURL))
		}

		// Create model for embedder
		modelId := llms.ModelId{
			Provider: llms.ModelProvider(cfg.Provider),
			ID:       cfg.Model,
		}
		model, _ := llms.GetModel(modelId)

		embedderProvider, err := llms.NewEmbedderProvider(
			llms.ModelProvider(cfg.Provider),
			model,
			providerOpts...,
		)
		if err != nil {
			return nil, errors.Errorf(ErrorCodeCreateEmbedderFailed,
				"failed to create embedder provider: %v", err)
		}

		// Create LLM embedder
		return createLlmEmbedder(embedderProvider), nil
	default:
		return nil, errors.Errorf(ErrorCodeCreateEmbedderFailed,
			"unsupported embedder type: %s", cfg.Type)
	}
}

// createLlmEmbedder creates an embedder with the given provider
func createLlmEmbedder(embedderProvider llms.EmbedderProvider) embedder.Embedder {
	// Since we can't access the private field directly, we'll create a wrapper
	// that implements the Embedder interface
	return &embedderWrapper{
		provider: embedderProvider,
	}
}

// embedderWrapper wraps the embedder provider to implement the Embedder interface
type embedderWrapper struct {
	provider llms.EmbedderProvider
}

func (w *embedderWrapper) Embed(ctx context.Context, texts []string) ([]embedder.FloatVector, error) {
	response, err := w.provider.GetEmbeddings(ctx, texts)
	if err != nil {
		return nil, err
	}
	var vectors []embedder.FloatVector
	for idx := range response.Vectors {
		vector := embedder.FloatVector(response.Vectors[idx])
		vectors = append(vectors, vector)
	}
	return vectors, nil
}

// calculateDataHash calculates MD5 hash of the data for change detection
func calculateDataHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// loadKnowledgeState loads the knowledge state from file
func loadKnowledgeState(stateFilePath string) *KnowledgeState {
	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		// File doesn't exist or can't be read, return empty state
		return &KnowledgeState{
			LoadedKnowledgeBases: make(map[string]KnowledgeBaseState),
		}
	}

	var state KnowledgeState
	if err := json.Unmarshal(data, &state); err != nil {
		// Invalid JSON, return empty state
		fmt.Printf("Warning: failed to parse knowledge state file: %v\n", err)
		return &KnowledgeState{
			LoadedKnowledgeBases: make(map[string]KnowledgeBaseState),
		}
	}

	return &state
}

// saveKnowledgeState saves the knowledge state to file
func saveKnowledgeState(stateFilePath string, state *KnowledgeState) error {
	// Ensure directory exists
	dir := filepath.Dir(stateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Errorf(ErrorCodeSaveKnowledgeStateFailed,
			"failed to create runtime directory: %v", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errors.Errorf(ErrorCodeSaveKnowledgeStateFailed,
			"failed to marshal knowledge state: %v", err)
	}

	// Write to file
	if err = os.WriteFile(stateFilePath, data, 0644); err != nil {
		return errors.Errorf(ErrorCodeSaveKnowledgeStateFailed,
			"failed to write knowledge state file: %v", err)
	}

	fmt.Println("# Knowledge state saved")
	return nil
}

// shouldReloadKnowledgeBases determines if knowledge bases should be reloaded
func shouldReloadKnowledgeBases(existingState *KnowledgeState, dataHash string, kbDataList []KnowledgeBaseData) bool {
	// If data hash changed, reload everything
	if existingState.DataHash != dataHash {
		fmt.Println(" < Data has changed, reloading all knowledge bases...")
		return true
	}

	// If no existing state, need to load
	if len(existingState.LoadedKnowledgeBases) == 0 {
		fmt.Println(" < No existing knowledge base state found, loading all...")
		return false // We'll check individually for each KB
	}

	// Check if the number of knowledge bases changed
	if len(existingState.LoadedKnowledgeBases) != len(kbDataList) {
		fmt.Println(" < Number of knowledge bases changed, reloading all...")
		return true
	}

	fmt.Println("   < Data unchanged, will check individual knowledge bases...")
	return false
}

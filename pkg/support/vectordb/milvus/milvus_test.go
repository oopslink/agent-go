package milvus

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/oopslink/agent-go/pkg/support/document"
	"github.com/oopslink/agent-go/pkg/support/embedder"
	"github.com/oopslink/agent-go/pkg/support/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbedder is a mock implementation of the embedder.Embedder interface
type MockEmbedder struct {
	dimension int
}

func (m *MockEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.FloatVector, error) {
	vectors := make([]embedder.FloatVector, len(texts))
	for i := range texts {
		// Generate deterministic vectors based on text length and dimension
		vector := make([]float64, m.dimension)
		for j := 0; j < m.dimension; j++ {
			vector[j] = float64((len(texts[i])+j)%10) / 10.0
		}
		vectors[i] = embedder.FloatVector(vector)
	}
	return vectors, nil
}

// getMilvusConfig returns Milvus configuration from environment variables
// Returns nil if required environment variables are not set
func getMilvusConfig() *client.Config {
	endpoint := os.Getenv("MILVUS_ENDPOINT")
	if endpoint == "" {
		return nil
	}

	// Add authentication if provided
	username := os.Getenv("MILVUS_USER_NAME")
	password := os.Getenv("MILVUS_PASSWORD")
	if username == "" || password == "" {
		return nil
	}

	config, _ := NewClientConfig(endpoint, username, password)
	return config
}

func setupMilvusTest(t *testing.T) (*Store, func()) {
	ctx := context.Background()

	// Get Milvus configuration from environment variables
	config := getMilvusConfig()
	if config == nil {
		t.Skip("Skipping Milvus tests: MILVUS_ENDPOINT not set")
	}

	store, err := New(ctx, *config)
	if err != nil {
		t.Skipf("Skipping Milvus tests: cannot connect to Milvus at %s: %v", config.Address, err)
	}

	// Cleanup function
	cleanup := func() {
		if store != nil && store.client != nil {
			// Clean up any test collections
			collections, _ := store.client.ListCollections(ctx)
			for _, collection := range collections {
				if collection.Name == "test_collection" ||
					collection.Name == "test_documents" ||
					collection.Name == "test_search_collection" ||
					collection.Name == "test_threshold_collection" ||
					collection.Name == "concurrent_test_collection" {
					store.client.DropCollection(ctx, collection.Name)
				}
			}
		}
	}

	return store, cleanup
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	config := getMilvusConfig()
	if config == nil {
		t.Skip("Skipping test: MILVUS_ENDPOINT not set")
	}

	store, err := New(ctx, *config)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to Milvus: %v", err)
	}
	defer func() {
		if store != nil && store.client != nil {
			// Cleanup any collections created during test
		}
	}()

	assert.NotNil(t, store)
	assert.NotNil(t, store.client)
	assert.NotNil(t, store.collections)
}

func TestTimeoutConfiguration(t *testing.T) {
	ctx := context.Background()
	config := getMilvusConfig()
	if config == nil {
		t.Skip("Skipping test: MILVUS_ENDPOINT not set")
	}

	// Test default timeout
	store, err := New(ctx, *config)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to Milvus: %v", err)
	}
	assert.NotNil(t, store)

	// Test custom timeout
	storeWithTimeout, err := NewWithTimeout(ctx, *config, 30*time.Second)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to Milvus with custom timeout: %v", err)
	}
	assert.NotNil(t, storeWithTimeout)
}

func TestCollectionSchema(t *testing.T) {
	schema := DefaultCollectionSchema()

	assert.Equal(t, "documents", schema.CollectionName)
	assert.Equal(t, "text", schema.TextField)
	assert.Equal(t, "metadata", schema.MetaField)
	assert.Equal(t, "id", schema.PrimaryField)
	assert.Equal(t, "vector", schema.VectorField)
	assert.Equal(t, 65535, schema.MaxTextLength)
	assert.Equal(t, int32(1), schema.ShardNum)
	assert.Equal(t, entity.L2, schema.MetricType)
	assert.NotNil(t, schema.Index)
}

func TestMilvusInsertOptions(t *testing.T) {
	options := NewMilvusInsertOptions()

	// Test default values
	assert.NotNil(t, options.GetCollectionSchema())
	assert.False(t, options.GetDropOld())
	assert.False(t, options.GetAsync())
	assert.Empty(t, options.GetCollection())
	assert.Nil(t, options.GetEmbedder())
	assert.Empty(t, options.GetPartitionName())
	assert.False(t, options.GetSkipFlushOnWrite())

	// Test setters
	mockEmbedder := &MockEmbedder{dimension: 128}
	customSchema := &CollectionSchema{
		CollectionName: "custom",
		TextField:      "content",
		MetaField:      "meta",
		PrimaryField:   "pk",
		VectorField:    "vec",
		MaxTextLength:  1000,
		ShardNum:       2,
		MetricType:     entity.IP,
	}

	options.SetCollection("test_collection")
	options.SetEmbedder(mockEmbedder)
	options.SetPartitionName("test_partition")
	options.SetSkipFlushOnWrite(true)
	options.SetCollectionSchema(customSchema)
	options.SetDropOld(true)
	options.SetAsync(true)

	assert.Equal(t, "test_collection", options.GetCollection())
	assert.Equal(t, mockEmbedder, options.GetEmbedder())
	assert.Equal(t, "test_partition", options.GetPartitionName())
	assert.True(t, options.GetSkipFlushOnWrite())
	assert.Equal(t, customSchema, options.GetCollectionSchema())
	assert.True(t, options.GetDropOld())
	assert.True(t, options.GetAsync())
}

func TestMilvusSearchOptions(t *testing.T) {
	options := NewMilvusSearchOptions()

	// Test default values
	assert.Equal(t, entity.ClBounded, options.GetConsistencyLevel())
	assert.NotNil(t, options.GetSearchParameters())
	assert.Empty(t, options.GetCollection())
	assert.Nil(t, options.GetEmbedder())
	assert.Empty(t, options.GetPartitionNames())
	assert.Equal(t, float32(0), options.GetScoreThreshold())
	assert.Empty(t, options.GetFilterExpression())

	// Test setters
	mockEmbedder := &MockEmbedder{dimension: 128}
	searchParams, _ := entity.NewIndexHNSWSearchParam(32)
	partitions := []string{"partition1", "partition2"}

	options.SetCollection("test_collection")
	options.SetEmbedder(mockEmbedder)
	options.SetPartitionNames(partitions)
	options.SetScoreThreshold(0.8)
	options.SetFilterExpression("id > 100")
	options.SetConsistencyLevel(entity.ClStrong)
	options.SetSearchParameters(searchParams)

	assert.Equal(t, "test_collection", options.GetCollection())
	assert.Equal(t, mockEmbedder, options.GetEmbedder())
	assert.Equal(t, partitions, options.GetPartitionNames())
	assert.Equal(t, float32(0.8), options.GetScoreThreshold())
	assert.Equal(t, "id > 100", options.GetFilterExpression())
	assert.Equal(t, entity.ClStrong, options.GetConsistencyLevel())
	assert.Equal(t, searchParams, options.GetSearchParameters())
}

func TestMilvusOptionFunctions(t *testing.T) {
	// Test Insert Options
	insertOpts := NewMilvusInsertOptions()
	mockEmbedder := &MockEmbedder{dimension: 128}
	customSchema := DefaultCollectionSchema()
	customSchema.CollectionName = "custom_collection"

	vectordb.WithInsertCollection("test_collection")(insertOpts)
	vectordb.WithInsertEmbedder(mockEmbedder)(insertOpts)
	WithMilvusInsertPartition("test_partition")(insertOpts)
	WithMilvusSkipFlush(true)(insertOpts)
	WithMilvusCollectionSchema(customSchema)(insertOpts)
	WithMilvusDropOld(true)(insertOpts)
	WithMilvusAsync(true)(insertOpts)

	assert.Equal(t, "test_collection", insertOpts.GetCollection())
	assert.Equal(t, mockEmbedder, insertOpts.GetEmbedder())
	assert.Equal(t, "test_partition", insertOpts.GetPartitionName())
	assert.True(t, insertOpts.GetSkipFlushOnWrite())
	assert.Equal(t, customSchema, insertOpts.GetCollectionSchema())
	assert.True(t, insertOpts.GetDropOld())
	assert.True(t, insertOpts.GetAsync())

	// Test Search Options
	searchOpts := NewMilvusSearchOptions()
	searchParams, _ := entity.NewIndexHNSWSearchParam(32)

	vectordb.WithSearchCollection("test_collection")(searchOpts)
	vectordb.WithSearchEmbedder(mockEmbedder)(searchOpts)
	vectordb.WithScoreThreshold(0.7)(searchOpts)
	vectordb.WithFilters("test filter")(searchOpts)
	WithMilvusSearchPartitions("partition1", "partition2")(searchOpts)
	WithMilvusSearchParams(searchParams)(searchOpts)
	WithMilvusConsistencyLevel(entity.ClStrong)(searchOpts)
	WithMilvusFilterExpression("id > 50")(searchOpts)

	assert.Equal(t, "test_collection", searchOpts.GetCollection())
	assert.Equal(t, mockEmbedder, searchOpts.GetEmbedder())
	assert.Equal(t, float32(0.7), searchOpts.GetScoreThreshold())
	assert.Equal(t, "id > 50", searchOpts.GetFilterExpression()) // Milvus-specific filter takes precedence
	assert.Equal(t, []string{"partition1", "partition2"}, searchOpts.GetPartitionNames())
	assert.Equal(t, searchParams, searchOpts.GetSearchParameters())
	assert.Equal(t, entity.ClStrong, searchOpts.GetConsistencyLevel())
}

func TestAddDocuments(t *testing.T) {
	store, cleanup := setupMilvusTest(t)
	defer cleanup()

	ctx := context.Background()
	mockEmbedder := &MockEmbedder{dimension: 128}

	// Test documents
	docs := []*document.Document{
		{
			Id:      document.DocumentId("doc1"),
			Content: "This is the first test document",
			Metadata: map[string]any{
				"category": "test",
				"priority": 1,
			},
		},
		{
			Id:      document.DocumentId("doc2"),
			Content: "This is the second test document",
			Metadata: map[string]any{
				"category": "test",
				"priority": 2,
			},
		},
	}

	// Add documents with custom collection schema
	customSchema := DefaultCollectionSchema()
	customSchema.CollectionName = "test_documents"

	ids, err := store.AddDocuments(ctx, docs,
		vectordb.WithInsertEmbedder(mockEmbedder),
		WithMilvusCollectionSchema(customSchema),
		WithMilvusDropOld(true),
		WithMilvusAsync(false),
	)

	require.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Equal(t, document.DocumentId("doc1"), ids[0])
	assert.Equal(t, document.DocumentId("doc2"), ids[1])

	// Verify collection was created
	assert.Contains(t, store.collections, "test_documents")
	collectionInfo := store.collections["test_documents"]
	assert.True(t, collectionInfo.collectionExists)
	assert.NotNil(t, collectionInfo.schema)
	assert.Equal(t, customSchema, collectionInfo.collectionSchema)
}

func TestSearch(t *testing.T) {
	store, cleanup := setupMilvusTest(t)
	defer cleanup()

	ctx := context.Background()
	mockEmbedder := &MockEmbedder{dimension: 128}

	// First add some documents
	docs := []*document.Document{
		{
			Id:      document.DocumentId("doc1"),
			Content: "Machine learning and artificial intelligence",
			Metadata: map[string]any{
				"category": "AI",
				"score":    0.95,
			},
		},
		{
			Id:      document.DocumentId("doc2"),
			Content: "Database systems and data management",
			Metadata: map[string]any{
				"category": "Database",
				"score":    0.87,
			},
		},
		{
			Id:      document.DocumentId("doc3"),
			Content: "Web development and frontend frameworks",
			Metadata: map[string]any{
				"category": "Web",
				"score":    0.78,
			},
		},
	}

	customSchema := DefaultCollectionSchema()
	customSchema.CollectionName = "test_search_collection"

	_, err := store.AddDocuments(ctx, docs,
		vectordb.WithInsertEmbedder(mockEmbedder),
		WithMilvusCollectionSchema(customSchema),
		WithMilvusDropOld(true),
		WithMilvusAsync(false),
	)
	require.NoError(t, err)

	// Search for documents
	results, err := store.Search(ctx, "artificial intelligence", 10,
		vectordb.WithSearchCollection("test_search_collection"),
		vectordb.WithSearchEmbedder(mockEmbedder),
		vectordb.WithScoreThreshold(0.0),
		WithMilvusConsistencyLevel(entity.ClStrong),
	)

	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// Verify search results structure
	for _, result := range results {
		assert.NotEmpty(t, result.Content)
		assert.NotNil(t, result.Metadata)
		assert.GreaterOrEqual(t, result.Score, float32(0.0))
		// Note: Milvus L2 distance can be greater than 1, so we don't enforce upper bound
	}
}

func TestSearchWithScoreThreshold(t *testing.T) {
	store, cleanup := setupMilvusTest(t)
	defer cleanup()

	ctx := context.Background()
	mockEmbedder := &MockEmbedder{dimension: 128}

	// Add test documents
	docs := []*document.Document{
		{
			Id:       document.DocumentId("doc1"),
			Content:  "This is a test document",
			Metadata: map[string]any{"id": 1},
		},
	}

	customSchema := DefaultCollectionSchema()
	customSchema.CollectionName = "test_threshold_collection"

	_, err := store.AddDocuments(ctx, docs,
		vectordb.WithInsertEmbedder(mockEmbedder),
		WithMilvusCollectionSchema(customSchema),
		WithMilvusDropOld(true),
	)
	require.NoError(t, err)

	// Search with high score threshold (should return fewer or no results)
	results, err := store.Search(ctx, "completely different content", 10,
		vectordb.WithSearchCollection("test_threshold_collection"),
		vectordb.WithSearchEmbedder(mockEmbedder),
		vectordb.WithScoreThreshold(0.9),
	)

	require.NoError(t, err)
	// With a high threshold and different content, we should get fewer results
	for _, result := range results {
		assert.GreaterOrEqual(t, result.Score, float32(0.9))
	}
}

func TestConcurrentOperations(t *testing.T) {
	store, cleanup := setupMilvusTest(t)
	defer cleanup()

	ctx := context.Background()
	mockEmbedder := &MockEmbedder{dimension: 128}

	// Test concurrent collection creation
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			docs := []*document.Document{
				{
					Id:       document.DocumentId(fmt.Sprintf("doc_%d", id)),
					Content:  fmt.Sprintf("Test document %d", id),
					Metadata: map[string]any{"id": id},
				},
			}

			customSchema := DefaultCollectionSchema()
			customSchema.CollectionName = "concurrent_test_collection"

			_, err := store.AddDocuments(ctx, docs,
				vectordb.WithInsertEmbedder(mockEmbedder),
				WithMilvusCollectionSchema(customSchema),
				WithMilvusDropOld(false), // Don't drop in concurrent test
			)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// Verify collection was created only once
	assert.Contains(t, store.collections, "concurrent_test_collection")
}

func TestErrorHandling(t *testing.T) {
	store, cleanup := setupMilvusTest(t)
	defer cleanup()

	ctx := context.Background()

	docs := []*document.Document{
		{
			Id:       document.DocumentId("doc1"),
			Content:  "Test document",
			Metadata: map[string]any{"test": true},
		},
	}

	// Test missing embedder
	_, err := store.AddDocuments(ctx, docs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedder provided")

	// Test search without embedder
	_, err = store.Search(ctx, "test query", 10,
		vectordb.WithSearchCollection("nonexistent"),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedder provided")

	// Test search on nonexistent collection
	mockEmbedder := &MockEmbedder{dimension: 128}
	_, err = store.Search(ctx, "test query", 10,
		vectordb.WithSearchCollection("nonexistent"),
		vectordb.WithSearchEmbedder(mockEmbedder),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection nonexistent not found")
}

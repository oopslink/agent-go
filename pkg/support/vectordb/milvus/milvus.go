package milvus

import (
	"context"
	"encoding/json"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"strconv"
	"sync"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/oopslink/agent-go/pkg/support/document"
	"github.com/oopslink/agent-go/pkg/support/embedder"
	"github.com/oopslink/agent-go/pkg/support/vectordb"
)

const (
	// DefaultTimeout is the default timeout for Milvus client operations
	DefaultTimeout = 15 * time.Second
)

var (
	_ vectordb.VectorDB = &Store{}
)

// CollectionSchema holds configuration for Milvus collection schema.
type CollectionSchema struct {
	CollectionName string
	TextField      string
	MetaField      string
	PrimaryField   string
	VectorField    string
	MaxTextLength  int
	ShardNum       int32
	MetricType     entity.MetricType
	Index          entity.Index
}

// DefaultCollectionSchema returns a default collection schema configuration.
func DefaultCollectionSchema() *CollectionSchema {
	index, _ := entity.NewIndexHNSW(entity.L2, 8, 64)
	return &CollectionSchema{
		CollectionName: "documents",
		TextField:      "text",
		MetaField:      "metadata",
		PrimaryField:   "id",
		VectorField:    "vector",
		MaxTextLength:  65535,
		ShardNum:       1,
		MetricType:     entity.L2,
		Index:          index,
	}
}

// MilvusInsertOptions implements vectordb.InsertOptions with Milvus-specific fields.
type MilvusInsertOptions struct {
	collection       string
	embedder         embedder.Embedder
	partitionName    string
	skipFlushOnWrite bool
	collectionSchema *CollectionSchema
	dropOld          bool
	async            bool
}

// MilvusSearchOptions implements vectordb.SearchOptions with Milvus-specific fields.
type MilvusSearchOptions struct {
	collection       string
	scoreThreshold   float32
	filters          string // Milvus filter expression
	embedder         embedder.Embedder
	partitionNames   []string
	searchParameters entity.SearchParam
	consistencyLevel entity.ConsistencyLevel
}

// Implement vectordb.InsertOptions interface
func (o *MilvusInsertOptions) GetCollection() string {
	return o.collection
}

func (o *MilvusInsertOptions) GetEmbedder() embedder.Embedder {
	return o.embedder
}

// Implement vectordb.SearchOptions interface
func (o *MilvusSearchOptions) GetCollection() string {
	return o.collection
}

func (o *MilvusSearchOptions) GetScoreThreshold() float32 {
	return o.scoreThreshold
}

func (o *MilvusSearchOptions) GetFilters() any {
	return o.filters
}

func (o *MilvusSearchOptions) GetEmbedder() embedder.Embedder {
	return o.embedder
}

// Milvus-specific getters for SearchOptions
func (o *MilvusSearchOptions) GetPartitionNames() []string {
	return o.partitionNames
}

func (o *MilvusSearchOptions) GetSearchParameters() entity.SearchParam {
	return o.searchParameters
}

func (o *MilvusSearchOptions) GetConsistencyLevel() entity.ConsistencyLevel {
	return o.consistencyLevel
}

func (o *MilvusSearchOptions) GetFilterExpression() string {
	return o.filters
}

// Milvus-specific getters for InsertOptions
func (o *MilvusInsertOptions) GetPartitionName() string {
	return o.partitionName
}

func (o *MilvusInsertOptions) GetSkipFlushOnWrite() bool {
	return o.skipFlushOnWrite
}

func (o *MilvusInsertOptions) GetCollectionSchema() *CollectionSchema {
	return o.collectionSchema
}

func (o *MilvusInsertOptions) GetDropOld() bool {
	return o.dropOld
}

func (o *MilvusInsertOptions) GetAsync() bool {
	return o.async
}

// Setters for MilvusInsertOptions
func (o *MilvusInsertOptions) SetCollection(collection string) {
	o.collection = collection
}

func (o *MilvusInsertOptions) SetEmbedder(embedder embedder.Embedder) {
	o.embedder = embedder
}

func (o *MilvusInsertOptions) SetPartitionName(name string) {
	o.partitionName = name
}

func (o *MilvusInsertOptions) SetSkipFlushOnWrite(skip bool) {
	o.skipFlushOnWrite = skip
}

func (o *MilvusInsertOptions) SetCollectionSchema(schema *CollectionSchema) {
	o.collectionSchema = schema
}

func (o *MilvusInsertOptions) SetDropOld(dropOld bool) {
	o.dropOld = dropOld
}

func (o *MilvusInsertOptions) SetAsync(async bool) {
	o.async = async
}

// Setters for MilvusSearchOptions
func (o *MilvusSearchOptions) SetCollection(collection string) {
	o.collection = collection
}

func (o *MilvusSearchOptions) SetScoreThreshold(threshold float32) {
	o.scoreThreshold = threshold
}

func (o *MilvusSearchOptions) SetFilters(filters any) {
	if str, ok := filters.(string); ok {
		o.filters = str
	}
}

func (o *MilvusSearchOptions) SetEmbedder(embedder embedder.Embedder) {
	o.embedder = embedder
}

func (o *MilvusSearchOptions) SetPartitionNames(names []string) {
	o.partitionNames = names
}

func (o *MilvusSearchOptions) SetSearchParameters(params entity.SearchParam) {
	o.searchParameters = params
}

func (o *MilvusSearchOptions) SetConsistencyLevel(level entity.ConsistencyLevel) {
	o.consistencyLevel = level
}

func (o *MilvusSearchOptions) SetFilterExpression(filter string) {
	o.filters = filter
}

// NewMilvusInsertOptions creates a new MilvusInsertOptions instance.
func NewMilvusInsertOptions() *MilvusInsertOptions {
	return &MilvusInsertOptions{
		collectionSchema: DefaultCollectionSchema(),
		dropOld:          false,
		async:            false,
	}
}

// NewMilvusSearchOptions creates a new MilvusSearchOptions instance.
func NewMilvusSearchOptions() *MilvusSearchOptions {
	searchParams, _ := entity.NewIndexHNSWSearchParam(64)
	return &MilvusSearchOptions{
		consistencyLevel: entity.ClBounded,
		searchParameters: searchParams,
	}
}

// Milvus-specific option functions for Insert
func WithMilvusInsertPartition(partitionName string) vectordb.InsertOption {
	return func(o vectordb.InsertOptions) {
		if milvusOpts, ok := o.(*MilvusInsertOptions); ok {
			milvusOpts.SetPartitionName(partitionName)
		}
	}
}

func WithMilvusSkipFlush(skip bool) vectordb.InsertOption {
	return func(o vectordb.InsertOptions) {
		if milvusOpts, ok := o.(*MilvusInsertOptions); ok {
			milvusOpts.SetSkipFlushOnWrite(skip)
		}
	}
}

func WithMilvusCollectionSchema(schema *CollectionSchema) vectordb.InsertOption {
	return func(o vectordb.InsertOptions) {
		if milvusOpts, ok := o.(*MilvusInsertOptions); ok {
			milvusOpts.SetCollectionSchema(schema)
		}
	}
}

func WithMilvusDropOld(dropOld bool) vectordb.InsertOption {
	return func(o vectordb.InsertOptions) {
		if milvusOpts, ok := o.(*MilvusInsertOptions); ok {
			milvusOpts.SetDropOld(dropOld)
		}
	}
}

func WithMilvusAsync(async bool) vectordb.InsertOption {
	return func(o vectordb.InsertOptions) {
		if milvusOpts, ok := o.(*MilvusInsertOptions); ok {
			milvusOpts.SetAsync(async)
		}
	}
}

// Milvus-specific option functions for Search
func WithMilvusSearchPartitions(partitionNames ...string) vectordb.SearchOption {
	return func(o vectordb.SearchOptions) {
		if milvusOpts, ok := o.(*MilvusSearchOptions); ok {
			milvusOpts.SetPartitionNames(partitionNames)
		}
	}
}

func WithMilvusSearchParams(params entity.SearchParam) vectordb.SearchOption {
	return func(o vectordb.SearchOptions) {
		if milvusOpts, ok := o.(*MilvusSearchOptions); ok {
			milvusOpts.SetSearchParameters(params)
		}
	}
}

func WithMilvusConsistencyLevel(level entity.ConsistencyLevel) vectordb.SearchOption {
	return func(o vectordb.SearchOptions) {
		if milvusOpts, ok := o.(*MilvusSearchOptions); ok {
			milvusOpts.SetConsistencyLevel(level)
		}
	}
}

func WithMilvusFilterExpression(filter string) vectordb.SearchOption {
	return func(o vectordb.SearchOptions) {
		if milvusOpts, ok := o.(*MilvusSearchOptions); ok {
			milvusOpts.SetFilterExpression(filter)
		}
	}
}

// MilvusUpdateOptions implements vectordb.UpdateOptions with Milvus-specific fields.
type MilvusUpdateOptions struct {
	collection       string
	embedder         embedder.Embedder
	partitionName    string
	skipFlushOnWrite bool
	collectionSchema *CollectionSchema
	async            bool
}

// Implement vectordb.UpdateOptions interface
func (o *MilvusUpdateOptions) GetCollection() string {
	return o.collection
}

func (o *MilvusUpdateOptions) GetEmbedder() embedder.Embedder {
	return o.embedder
}

// Milvus-specific getters for UpdateOptions
func (o *MilvusUpdateOptions) GetPartitionName() string {
	return o.partitionName
}

func (o *MilvusUpdateOptions) GetSkipFlushOnWrite() bool {
	return o.skipFlushOnWrite
}

func (o *MilvusUpdateOptions) GetCollectionSchema() *CollectionSchema {
	return o.collectionSchema
}

func (o *MilvusUpdateOptions) GetAsync() bool {
	return o.async
}

// Setters for MilvusUpdateOptions
func (o *MilvusUpdateOptions) SetCollection(collection string) {
	o.collection = collection
}

func (o *MilvusUpdateOptions) SetEmbedder(embedder embedder.Embedder) {
	o.embedder = embedder
}

func (o *MilvusUpdateOptions) SetPartitionName(name string) {
	o.partitionName = name
}

func (o *MilvusUpdateOptions) SetSkipFlushOnWrite(skip bool) {
	o.skipFlushOnWrite = skip
}

func (o *MilvusUpdateOptions) SetCollectionSchema(schema *CollectionSchema) {
	o.collectionSchema = schema
}

func (o *MilvusUpdateOptions) SetAsync(async bool) {
	o.async = async
}

// NewMilvusUpdateOptions creates a new MilvusUpdateOptions instance.
func NewMilvusUpdateOptions() *MilvusUpdateOptions {
	return &MilvusUpdateOptions{
		collectionSchema: DefaultCollectionSchema(),
		async:            false,
	}
}

// Milvus-specific option functions for Update
func WithMilvusUpdatePartition(partitionName string) vectordb.UpdateOption {
	return func(o vectordb.UpdateOptions) {
		if milvusOpts, ok := o.(*MilvusUpdateOptions); ok {
			milvusOpts.SetPartitionName(partitionName)
		}
	}
}

func WithMilvusUpdateSkipFlush(skip bool) vectordb.UpdateOption {
	return func(o vectordb.UpdateOptions) {
		if milvusOpts, ok := o.(*MilvusUpdateOptions); ok {
			milvusOpts.SetSkipFlushOnWrite(skip)
		}
	}
}

func WithMilvusUpdateCollectionSchema(schema *CollectionSchema) vectordb.UpdateOption {
	return func(o vectordb.UpdateOptions) {
		if milvusOpts, ok := o.(*MilvusUpdateOptions); ok {
			milvusOpts.SetCollectionSchema(schema)
		}
	}
}

func WithMilvusUpdateAsync(async bool) vectordb.UpdateOption {
	return func(o vectordb.UpdateOptions) {
		if milvusOpts, ok := o.(*MilvusUpdateOptions); ok {
			milvusOpts.SetAsync(async)
		}
	}
}

type collectionInfo struct {
	schema           *entity.Schema
	loaded           bool
	collectionExists bool
	collectionSchema *CollectionSchema
}

func NewClientConfig(endpoint, username, password string) (*client.Config, error) {
	if endpoint == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeCreateVectorClientFailed, "endpoint cannot be empty")
	}
	if username == "" || password == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeCreateVectorClientFailed, "username and password cannot be empty")
	}
	return &client.Config{
		Address:  endpoint,
		Username: username,
		Password: password,
	}, nil
}

// NewClientConfigWithTimeout creates a new Milvus client configuration with custom timeout
func NewClientConfigWithTimeout(endpoint, username, password string, timeout time.Duration) (*client.Config, error) {
	if endpoint == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeCreateVectorClientFailed, "endpoint cannot be empty")
	}
	if username == "" || password == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeCreateVectorClientFailed, "username and password cannot be empty")
	}
	return &client.Config{
		Address:  endpoint,
		Username: username,
		Password: password,
	}, nil
}

// New creates an active client connection to the Milvus server.
func New(ctx context.Context, config client.Config) (*Store, error) {
	// Validate config before attempting connection
	if config.Address == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeCreateVectorStoreFailed, "milvus endpoint cannot be empty")
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	client, err := client.NewClient(ctxWithTimeout, config)
	if err != nil {
		return nil, err
	}

	return &Store{
		client:      client,
		collections: make(map[string]*collectionInfo),
	}, nil
}

// NewWithTimeout creates an active client connection to the Milvus server with custom timeout.
func NewWithTimeout(ctx context.Context, config client.Config, timeout time.Duration) (*Store, error) {
	// Validate config before attempting connection
	if config.Address == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeCreateVectorStoreFailed, "milvus endpoint cannot be empty")
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client, err := client.NewClient(ctxWithTimeout, config)
	if err != nil {
		return nil, err
	}

	return &Store{
		client:      client,
		collections: make(map[string]*collectionInfo),
	}, nil
}

// Store is a wrapper around the milvus client.
type Store struct {
	client      client.Client
	collections map[string]*collectionInfo
	mu          sync.RWMutex // 保护collections map的读写锁
}

// AddDocuments adds the text and metadata from the documents to the Milvus collection.
func (s *Store) AddDocuments(ctx context.Context, documents []*document.Document, opts ...vectordb.InsertOption) ([]document.DocumentId, error) {
	options := s.parseInsertOptions(opts...)

	// Use provided embedder or default
	embedder := options.GetEmbedder()
	if embedder == nil {
		return nil, errors.Errorf(vectordb.ErrorCodeAddDocumentFailed, "no embedder provided")
	}

	collectionName := options.GetCollection()
	if collectionName == "" {
		collectionName = options.GetCollectionSchema().CollectionName
	}

	// Get or create collection info
	info, err := s.getOrCreateCollection(ctx, collectionName, options.GetCollectionSchema(), options.GetDropOld())
	if err != nil {
		return nil, err
	}

	// Extract texts from documents
	texts := make([]string, 0, len(documents))
	for _, doc := range documents {
		texts = append(texts, doc.Content)
	}

	// Generate embeddings
	vectors, err := embedder.Embed(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(documents) {
		return nil, errors.Errorf(vectordb.ErrorCodeAddDocumentFailed,
			"number of vectors from embedder does not match number of documents")
	}

	if err := s.initCollection(ctx, collectionName, info, len(vectors[0]), options.GetAsync()); err != nil {
		return nil, err
	}

	// Prepare data for insertion
	colsData := make([]interface{}, 0, len(documents))
	docIds := make([]document.DocumentId, 0, len(documents))

	schema := info.collectionSchema
	for i, doc := range documents {
		// Convert float64 vector to float32 vector for Milvus
		vector32 := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			vector32[j] = float32(v)
		}

		docMap := map[string]any{
			schema.MetaField:   doc.Metadata,
			schema.TextField:   doc.Content,
			schema.VectorField: vector32,
		}
		colsData = append(colsData, docMap)
		docIds = append(docIds, doc.Id)
	}

	// Insert data into Milvus
	_, err = s.client.InsertRows(ctx, collectionName, options.GetPartitionName(), colsData)
	if err != nil {
		return nil, err
	}

	if !options.GetSkipFlushOnWrite() {
		if err = s.client.Flush(ctx, collectionName, false); err != nil {
			return nil, err
		}
	}

	return docIds, nil
}

// Search performs a similarity search in the vector database using the provided query.
func (s *Store) Search(ctx context.Context,
	query string, maxDocuments int, opts ...vectordb.SearchOption) ([]*vectordb.ScoredDocument, error) {
	options := s.parseSearchOptions(opts...)

	// Use provided embedder or default
	embedder := options.GetEmbedder()
	if embedder == nil {
		return nil, errors.Errorf(vectordb.ErrorCodeSearchDocumentFailed, "no embedder provided")
	}

	collectionName := options.GetCollection()
	if collectionName == "" {
		return nil, errors.Errorf(vectordb.ErrorCodeSearchDocumentFailed, "collection name is required")
	}

	// Get collection info with read lock
	info, err := s.getOrLoadCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	// Generate embedding for query
	vector, err := embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vector) == 0 {
		return nil, errors.Errorf(vectordb.ErrorCodeSearchDocumentFailed, "failed to generate embedding for query")
	}

	// Convert embedder.FloatVector to entity.FloatVector
	floatVector := make([]float32, len(vector[0]))
	for i, v := range vector[0] {
		floatVector[i] = float32(v)
	}
	vectors := []entity.Vector{
		entity.FloatVector(floatVector),
	}

	// Use configured partitions or default
	partitions := options.GetPartitionNames()

	// Use configured filters
	filter := options.GetFilterExpression()

	// Use configured search parameters
	sp := options.GetSearchParameters()

	searchResult, err := s.client.Search(ctx,
		collectionName,
		partitions,
		filter,
		s.getSearchFields(info),
		vectors,
		info.collectionSchema.VectorField,
		info.collectionSchema.MetricType,
		maxDocuments,
		sp,
		client.WithSearchQueryConsistencyLevel(options.GetConsistencyLevel()),
	)
	if err != nil {
		return nil, err
	}

	docs, err := s.convertResultToDocument(searchResult, info)
	if err != nil {
		return nil, err
	}

	// Apply score threshold if specified
	if threshold := options.GetScoreThreshold(); threshold > 0 {
		filteredDocs := make([]*vectordb.ScoredDocument, 0)
		for _, doc := range docs {
			if doc.Score >= threshold {
				filteredDocs = append(filteredDocs, doc)
			}
		}
		docs = filteredDocs
	}

	return docs, nil
}

// UpdateDocuments updates existing documents in the Milvus collection.
func (s *Store) UpdateDocuments(ctx context.Context, documents []*document.Document, opts ...vectordb.UpdateOption) error {
	options := s.parseUpdateOptions(opts...)

	// Use provided embedder or default
	embedder := options.GetEmbedder()
	if embedder == nil {
		return errors.Errorf(vectordb.ErrorCodeUpdateDocumentFailed, "no embedder provided")
	}

	collectionName := options.GetCollection()
	if collectionName == "" {
		collectionName = options.GetCollectionSchema().CollectionName
	}

	// Get collection info
	info, err := s.getOrLoadCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	// Extract texts from documents
	texts := make([]string, 0, len(documents))
	for _, doc := range documents {
		texts = append(texts, doc.Content)
	}

	// Generate embeddings
	vectors, err := embedder.Embed(ctx, texts)
	if err != nil {
		return err
	}

	if len(vectors) != len(documents) {
		return errors.Errorf(vectordb.ErrorCodeUpdateDocumentFailed,
			"number of vectors from embedder does not match number of documents")
	}

	// Prepare data for update
	colsData := make([]interface{}, 0, len(documents))

	schema := info.collectionSchema
	for i, doc := range documents {
		// Convert float64 vector to float32 vector for Milvus
		vector32 := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			vector32[j] = float32(v)
		}

		docMap := map[string]any{
			schema.PrimaryField: doc.Id,
			schema.MetaField:    doc.Metadata,
			schema.TextField:    doc.Content,
			schema.VectorField:  vector32,
		}
		colsData = append(colsData, docMap)
	}

	// Update data in Milvus using insert (Milvus will handle upsert based on primary key)
	_, err = s.client.InsertRows(ctx, collectionName, options.GetPartitionName(), colsData)
	if err != nil {
		return err
	}

	if !options.GetSkipFlushOnWrite() {
		if err = s.client.Flush(ctx, collectionName, false); err != nil {
			return err
		}
	}

	return nil
}

// Get retrieves a document by its ID from the Milvus collection.
func (s *Store) Get(ctx context.Context, documentId document.DocumentId) (*document.Document, error) {
	// For Milvus, we need to search by document ID
	// Since we don't have a direct get method, we'll use a search with filter
	// This is a simplified implementation - in practice, you might want to use a different approach

	// Create a simple embedder for the search (this might not be the best approach)
	// In a real implementation, you might want to store document IDs separately or use a different approach

	// For now, we'll return an error indicating this method needs a different implementation
	return nil, errors.Errorf(errors.NotImplemented,
		"get operation not directly supported by Milvus - consider using search with document ID filter")
}

// Delete removes a document by its ID from the Milvus collection.
func (s *Store) Delete(ctx context.Context, documentId document.DocumentId) error {
	// For Milvus, we need to delete by document ID
	// This requires knowing the collection name, which we don't have in the interface
	// We'll need to make some assumptions or modify the interface

	// For now, we'll return an error indicating this method needs a different implementation
	return errors.Errorf(errors.NotImplemented,
		"delete operation requires collection name - consider using Milvus-specific delete methods")
}

// parseInsertOptions parses both standard and Milvus-specific insert options.
func (s *Store) parseInsertOptions(opts ...vectordb.InsertOption) *MilvusInsertOptions {
	options := NewMilvusInsertOptions()

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// parseSearchOptions parses both standard and Milvus-specific search options.
func (s *Store) parseSearchOptions(opts ...vectordb.SearchOption) *MilvusSearchOptions {
	options := NewMilvusSearchOptions()

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// parseUpdateOptions parses both standard and Milvus-specific update options.
func (s *Store) parseUpdateOptions(opts ...vectordb.UpdateOption) *MilvusUpdateOptions {
	options := NewMilvusUpdateOptions()

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// getOrCreateCollection gets or creates collection info for the given collection name.
func (s *Store) getOrCreateCollection(ctx context.Context,
	collectionName string, schema *CollectionSchema, dropOld bool) (*collectionInfo, error) {
	// 先尝试读锁获取collection
	s.mu.RLock()
	if info, exists := s.collections[collectionName]; exists {
		s.mu.RUnlock()
		return info, nil
	}
	s.mu.RUnlock()

	// 需要创建collection，获取写锁
	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查，防止在获取写锁期间其他goroutine已经创建了collection
	if info, exists := s.collections[collectionName]; exists {
		return info, nil
	}

	// Check if collection exists in Milvus
	exists, err := s.client.HasCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	info := &collectionInfo{
		collectionExists: exists,
		collectionSchema: schema,
	}

	if exists && dropOld {
		if err := s.client.DropCollection(ctx, collectionName); err != nil {
			return nil, err
		}
		info.collectionExists = false
	}

	s.collections[collectionName] = info
	return info, nil
}

// getOrLoadCollection safely gets collection info with read lock, or loads it from remote database.
func (s *Store) getOrLoadCollection(ctx context.Context, collectionName string) (*collectionInfo, error) {
	// First try to get from memory with read lock
	s.mu.RLock()
	if info, exists := s.collections[collectionName]; exists {
		s.mu.RUnlock()
		return info, nil
	}
	s.mu.RUnlock()

	// Not found in memory, try to load from remote database
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check in case another goroutine loaded it while we were waiting for write lock
	if info, exists := s.collections[collectionName]; exists {
		return info, nil
	}

	// Check if collection exists in remote Milvus database
	exists, err := s.client.HasCollection(ctx, collectionName)
	if err != nil {
		return nil, errors.Errorf(vectordb.ErrorCodeLoadCollectionFailed,
			"failed to check collection existence: %s", err.Error())
	}

	if !exists {
		return nil, errors.Errorf(vectordb.ErrorCodeLoadCollectionFailed,
			"collection %s not found", collectionName)
	}

	// Collection exists in remote database, load its schema and create info
	collection, err := s.client.DescribeCollection(ctx, collectionName)
	if err != nil {
		return nil, errors.Errorf(vectordb.ErrorCodeLoadCollectionFailed,
			"failed to describe collection: %s", err.Error())
	}

	// Create default collection schema from the remote schema
	schema := DefaultCollectionSchema()
	schema.CollectionName = collectionName

	// Try to identify fields from the remote schema
	for _, field := range collection.Schema.Fields {
		switch field.DataType {
		case entity.FieldTypeVarChar:
			if field.Name != schema.PrimaryField {
				schema.TextField = field.Name
			}
		case entity.FieldTypeJSON:
			schema.MetaField = field.Name
		case entity.FieldTypeFloatVector, entity.FieldTypeBinaryVector:
			schema.VectorField = field.Name
		case entity.FieldTypeInt64:
			if field.PrimaryKey {
				schema.PrimaryField = field.Name
			}
		}
	}

	info := &collectionInfo{
		schema:           collection.Schema,
		loaded:           false, // We'll check load status separately if needed
		collectionExists: true,
		collectionSchema: schema,
	}

	// Store in memory for future use
	s.collections[collectionName] = info
	return info, nil
}

// initCollection initializes a collection with the given dimension.
func (s *Store) initCollection(ctx context.Context, collectionName string, info *collectionInfo, dim int, async bool) error {
	if info.loaded {
		return nil
	}

	if err := s.createCollection(ctx, collectionName, info, dim); err != nil {
		return err
	}
	if err := s.extractFields(ctx, collectionName, info); err != nil {
		return err
	}
	if err := s.createIndex(ctx, collectionName, info, async); err != nil {
		return err
	}

	return s.loadCollection(ctx, collectionName, info, async)
}

// createCollection creates a new collection with the specified schema.
func (s *Store) createCollection(ctx context.Context, collectionName string, info *collectionInfo, dim int) error {
	if dim == 0 || info.collectionExists {
		return nil
	}

	schema := info.collectionSchema
	info.schema = &entity.Schema{
		CollectionName: collectionName,
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       schema.PrimaryField,
				DataType:   entity.FieldTypeInt64,
				AutoID:     true,
				PrimaryKey: true,
			},
			{
				Name:     schema.TextField,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: strconv.Itoa(schema.MaxTextLength),
				},
			},
			{
				Name:     schema.MetaField,
				DataType: entity.FieldTypeJSON,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: strconv.Itoa(schema.MaxTextLength),
				},
			},
			{
				Name:     schema.VectorField,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: strconv.Itoa(dim),
				},
			},
		},
	}

	err := s.client.CreateCollection(ctx, info.schema, schema.ShardNum, client.WithMetricsType(schema.MetricType))
	if err != nil {
		return err
	}
	info.collectionExists = true
	return nil
}

// extractFields extracts field information from an existing collection.
func (s *Store) extractFields(ctx context.Context, collectionName string, info *collectionInfo) error {
	if !info.collectionExists || info.schema != nil {
		return nil
	}
	collection, err := s.client.DescribeCollection(ctx, collectionName)
	if err != nil {
		return err
	}
	info.schema = collection.Schema
	return nil
}

// createIndex creates an index for the vector field.
func (s *Store) createIndex(ctx context.Context, collectionName string, info *collectionInfo, async bool) error {
	if !info.collectionExists {
		return nil
	}

	return s.client.CreateIndex(ctx, collectionName, info.collectionSchema.VectorField, info.collectionSchema.Index, async)
}

// loadCollection loads a collection into memory.
func (s *Store) loadCollection(ctx context.Context, collectionName string, info *collectionInfo, async bool) error {
	if info.loaded || !info.collectionExists {
		return nil
	}

	err := s.client.LoadCollection(ctx, collectionName, async)
	if err == nil {
		info.loaded = true
	}
	return err
}

// getSearchFields returns the non-vector fields for search results.
func (s *Store) getSearchFields(info *collectionInfo) []string {
	fields := []string{}
	for _, f := range info.schema.Fields {
		if f.DataType == entity.FieldTypeBinaryVector || f.DataType == entity.FieldTypeFloatVector {
			continue
		}
		fields = append(fields, f.Name)
	}
	return fields
}

// convertResultToDocument converts Milvus search results to ScoredDocument.
func (s *Store) convertResultToDocument(searchResult []client.SearchResult, info *collectionInfo) ([]*vectordb.ScoredDocument, error) {
	docs := []*vectordb.ScoredDocument{}
	var err error

	schema := info.collectionSchema
	for _, res := range searchResult {
		if res.ResultCount == 0 {
			continue
		}
		textcol, ok := res.Fields.GetColumn(schema.TextField).(*entity.ColumnVarChar)
		if !ok {
			return nil, errors.Errorf(vectordb.ErrorCodeInvalidVectorDataSchema,
				"text column missing")
		}
		metacol, ok := res.Fields.GetColumn(schema.MetaField).(*entity.ColumnJSONBytes)
		if !ok {
			return nil, errors.Errorf(vectordb.ErrorCodeInvalidVectorDataSchema,
				"metadata column missing")
		}
		for i := 0; i < res.ResultCount; i++ {
			doc := &vectordb.ScoredDocument{}

			doc.Content, err = textcol.ValueByIdx(i)
			if err != nil {
				return nil, err
			}
			metaStr, err := metacol.ValueByIdx(i)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(metaStr, &doc.Metadata); err != nil {
				return nil, err
			}
			doc.Score = float32(res.Scores[i])
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

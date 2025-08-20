package knowledge

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/oopslink/agent-go/pkg/support/document"
	"github.com/oopslink/agent-go/pkg/support/embedder"
	"github.com/oopslink/agent-go/pkg/support/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorDBStorage_Add(t *testing.T) {
	t.Run("successful add", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		doc := &document.Document{
			Id:      "test_doc_1",
			Name:    "test_document",
			Content: "test content",
			Metadata: map[string]any{
				"category": "test",
			},
		}

		mockVectorDB.EXPECT().
			AddDocuments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]document.DocumentId{"test_doc_1"}, nil)

		err := storage.Add(context.Background(), doc)
		assert.NoError(t, err)
	})

	t.Run("vector db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		doc := &document.Document{
			Id:      "test_doc_2",
			Name:    "test_document",
			Content: "test content",
		}

		mockVectorDB.EXPECT().
			AddDocuments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("vector db error"))

		err := storage.Add(context.Background(), doc)
		assert.Error(t, err)
	})
}

func TestVectorDBStorage_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		doc := &document.Document{
			Id:      "test_doc_1",
			Name:    "updated_document",
			Content: "updated content",
		}

		mockVectorDB.EXPECT().
			AddDocuments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]document.DocumentId{"test_doc_1"}, nil)

		err := storage.Update(context.Background(), doc.Id, doc)
		assert.NoError(t, err)
	})

	t.Run("vector db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		doc := &document.Document{
			Id:      "test_doc_2",
			Name:    "updated_document",
			Content: "updated content",
		}

		mockVectorDB.EXPECT().
			AddDocuments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("vector db error"))

		err := storage.Update(context.Background(), doc.Id, doc)
		assert.Error(t, err)
	})
}

func TestVectorDBStorage_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		expectedDoc := &document.Document{
			Id:      "test_doc_1",
			Name:    "test_document",
			Content: "test content",
		}

		mockVectorDB.EXPECT().
			Get(gomock.Any(), document.DocumentId("test_doc_1")).
			Return(expectedDoc, nil)

		doc, err := storage.Get(context.Background(), "test_doc_1")
		assert.NoError(t, err)
		assert.Equal(t, expectedDoc, doc)
	})

	t.Run("document not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		mockVectorDB.EXPECT().
			Get(gomock.Any(), document.DocumentId("non_existent_doc")).
			Return(nil, ErrDocumentNotFound)

		doc, err := storage.Get(context.Background(), "non_existent_doc")
		assert.Error(t, err)
		assert.Nil(t, doc)
		assert.Equal(t, ErrDocumentNotFound, err)
	})
}

func TestVectorDBStorage_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		mockVectorDB.EXPECT().
			Delete(gomock.Any(), document.DocumentId("test_doc_1")).
			Return(nil)

		err := storage.Delete(context.Background(), "test_doc_1")
		assert.NoError(t, err)
	})

	t.Run("vector db error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		mockVectorDB.EXPECT().
			Delete(gomock.Any(), document.DocumentId("test_doc_2")).
			Return(errors.New("vector db error"))

		err := storage.Delete(context.Background(), "test_doc_2")
		assert.Error(t, err)
	})
}

func TestVectorDBStorage_Search(t *testing.T) {
	t.Run("successful search", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		expectedResults := []*vectordb.ScoredDocument{
			{
				Document: document.Document{
					Id:      "result_1",
					Name:    "result_document_1",
					Content: "result content 1",
				},
				Score: 0.95,
			},
			{
				Document: document.Document{
					Id:      "result_2",
					Name:    "result_document_2",
					Content: "result content 2",
				},
				Score: 0.85,
			},
		}

		mockVectorDB.EXPECT().
			Search(gomock.Any(), "test query", 10, gomock.Any()).
			Return(expectedResults, nil)

		results, err := storage.Search(context.Background(), "test query")
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, document.DocumentId("result_1"), results[0].Id)
		assert.Equal(t, document.DocumentId("result_2"), results[1].Id)
	})

	t.Run("vector db search error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		mockVectorDB.EXPECT().
			Search(gomock.Any(), "test query", 10, gomock.Any()).
			Return(nil, errors.New("vector db search error"))

		results, err := storage.Search(context.Background(), "test query")
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestVectorDBStorage_MakeOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVectorDB := vectordb.NewMockVectorDB(ctrl)
	mockEmbedder := embedder.NewMockEmbedder(ctrl)
	storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

	t.Run("makeSearchOptions", func(t *testing.T) {
		maxResults, opts := storage.makeSearchOptions(
			WithMaxResults(5),
			WithScoreThreshold(0.8),
		)

		assert.Equal(t, 5, maxResults)
		assert.Len(t, opts, 3) // embedder + collection + score threshold
	})
}

func TestVectorDBStorage_Integration(t *testing.T) {
	t.Run("integration workflow", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockVectorDB := vectordb.NewMockVectorDB(ctrl)
		mockEmbedder := embedder.NewMockEmbedder(ctrl)
		storage := NewVectorDBStorage("test-collection", mockVectorDB, mockEmbedder)

		doc := &document.Document{
			Id:      "integration_test_doc",
			Name:    "integration_test",
			Content: "integration test content",
			Metadata: map[string]any{
				"test": true,
			},
		}

		// Setup mocks for the entire workflow
		mockVectorDB.EXPECT().
			AddDocuments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]document.DocumentId{"integration_test_doc"}, nil)

		mockVectorDB.EXPECT().
			Get(gomock.Any(), document.DocumentId("integration_test_doc")).
			Return(doc, nil)

		mockVectorDB.EXPECT().
			AddDocuments(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]document.DocumentId{"integration_test_doc"}, nil)

		searchResults := []*vectordb.ScoredDocument{
			{
				Document: *doc,
				Score:    0.95,
			},
		}

		mockVectorDB.EXPECT().
			Search(gomock.Any(), "search query", 10, gomock.Any()).
			Return(searchResults, nil)

		mockVectorDB.EXPECT().
			Delete(gomock.Any(), document.DocumentId("integration_test_doc")).
			Return(nil)

		// Execute workflow
		// 1. Add document
		err := storage.Add(context.Background(), doc)
		require.NoError(t, err)

		// 2. Get document
		retrieved, err := storage.Get(context.Background(), doc.Id)
		require.NoError(t, err)
		assert.Equal(t, doc, retrieved)

		// 3. Update document
		updatedDoc := &document.Document{
			Id:      doc.Id,
			Name:    "updated_integration_test",
			Content: "updated integration test content",
		}
		err = storage.Update(context.Background(), doc.Id, updatedDoc)
		require.NoError(t, err)

		// 4. Search documents
		results, err := storage.Search(context.Background(), "search query")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, doc.Id, results[0].Id)

		// 5. Delete document
		err = storage.Delete(context.Background(), doc.Id)
		require.NoError(t, err)
	})
}

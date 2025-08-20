package document

import (
	"encoding/json"
	"github.com/oopslink/agent-go/pkg/support/embedder"
)

type DocumentId string

func NewDocument(id DocumentId, name string, metadata map[string]any, content string) *Document {
	return &Document{
		Id:       id,
		Name:     name,
		Metadata: metadata,
		Content:  content,
	}
}

type Document struct {
	Id DocumentId `json:"id"`

	Name     string         `json:"name"`
	Metadata map[string]any `json:"metadata"`

	Content string `json:"content"`

	// Embedding is the vector representation of the document content.
	Embedding embedder.FloatVector `json:"_"`
}

func (d *Document) MarshalJSON() ([]byte, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

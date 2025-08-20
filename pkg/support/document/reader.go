package document

import (
	"io"
	"net/http"
	"os"
	"strings"
)

type ReaderOption func(*ReaderOptions)

type ReaderOptions struct {
	chunker Chunker // Optional chunker to split the document into smaller parts
}

func WithChunker(chunker Chunker) ReaderOption {
	return func(opts *ReaderOptions) {
		opts.chunker = chunker
	}
}

type Reader interface {

	// Read reads the content of the document and returns it as a string.
	Read(documentName string, reader io.Reader, options ...ReaderOption) ([]*Document, error)
}

var _ Reader = (*defaultReader)(nil)

type defaultReader struct{}

// NewDefaultReader creates a new default reader instance
func NewDefaultReader() Reader {
	return &defaultReader{}
}

func (r *defaultReader) Read(documentName string, reader io.Reader, options ...ReaderOption) ([]*Document, error) {
	// Apply options
	opts := &ReaderOptions{}
	for _, option := range options {
		option(opts)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		Id:       DocumentId("default"),
		Name:     documentName,
		Metadata: make(map[string]any),
		Content:  string(content),
	}

	if opts.chunker == nil {
		return []*Document{doc}, nil
	}

	return opts.chunker.Chunk(doc)
}

// OfFile, creates a io.Reader that reads from a file.
func OfFile(filePath string) (io.Reader, error) {
	return os.Open(filePath)
}

// OfString creates a io.Reader from a string.
func OfString(content string) (io.Reader, error) {
	return io.NopCloser(strings.NewReader(content)), nil
}

// OfUrl creates a io.Reader that reads from a URL.
func OfUrl(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

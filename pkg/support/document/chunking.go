package document

type Chunker interface {
	// Chunk takes a document and splits it into smaller chunks.
	Chunk(doc *Document) ([]*Document, error)
}

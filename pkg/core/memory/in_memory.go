package memory

// NewInMemoryMemory keep backward compatibility
func NewInMemoryMemory() Memory {
	return NewSimpleMemory()
}

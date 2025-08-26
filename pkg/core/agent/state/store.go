package state

// StateStore defines the storage interface
type StateStore interface {
	Get(key string) (any, error)
	Set(key string, value any) error
	Delete(key string) error
	Clear() error
}

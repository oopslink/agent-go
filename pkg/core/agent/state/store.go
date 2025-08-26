package state

// StateStore 定义存储接口
type StateStore interface {
	Get(key string) (any, error)
	Set(key string, value any) error
	Delete(key string) error
	Clear() error
}

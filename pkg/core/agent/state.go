package agent

type AgentState interface {
	Get(key string) (any, error)
	Put(key string, value any) error
}

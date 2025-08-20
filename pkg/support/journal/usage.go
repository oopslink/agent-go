package journal

import (
	"sync"
)

// UsageAccumulator 用量累加器
type UsageAccumulator struct {
	usages map[string]map[string]float64 // sessionId -> usageKey -> value
	mu     sync.RWMutex
}

func NewUsageAccumulator() *UsageAccumulator {
	return &UsageAccumulator{
		usages: make(map[string]map[string]float64),
	}
}

// Accumulate 累加用量
func (u *UsageAccumulator) Accumulate(sessionId string, usage map[string]float64) map[string]float64 {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.usages[sessionId] == nil {
		u.usages[sessionId] = make(map[string]float64)
	}

	for k, v := range usage {
		u.usages[sessionId][k] += v
	}

	// 返回当前累加后的结果
	result := make(map[string]float64)
	for k, v := range u.usages[sessionId] {
		result[k] = v
	}
	return result
}

// GetUsage 获取指定 session 的用量
func (u *UsageAccumulator) GetUsage(sessionId string) map[string]float64 {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if usage, exists := u.usages[sessionId]; exists {
		result := make(map[string]float64)
		for k, v := range usage {
			result[k] = v
		}
		return result
	}
	return nil
}

// GetAllUsages 获取所有用量数据
func (u *UsageAccumulator) GetAllUsages() map[string]map[string]float64 {
	u.mu.RLock()
	defer u.mu.RUnlock()

	result := make(map[string]map[string]float64)
	for sessionId, usage := range u.usages {
		result[sessionId] = make(map[string]float64)
		for k, v := range usage {
			result[sessionId][k] = v
		}
	}
	return result
}

// Reset 重置指定 session 的用量
func (u *UsageAccumulator) Reset(sessionId string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.usages, sessionId)
}

// ResetAll 重置所有用量
func (u *UsageAccumulator) ResetAll() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.usages = make(map[string]map[string]float64)
}

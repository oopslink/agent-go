package journal

import (
	"sync"
)

var (
	globalJournal Journal
	globalMu      sync.RWMutex
)

// SetGlobalJournal sets the global journal instance
func SetGlobalJournal(journal Journal) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalJournal = journal
}

// GetGlobalJournal returns the global journal instance
func GetGlobalJournal() Journal {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalJournal
}

// Debug logs a debug message using the global journal
func Debug(category, source, message string, data ...any) error {
	if j := GetGlobalJournal(); j != nil {
		return j.Debug(category, source, message, data...)
	}
	return nil
}

// Info logs an info message using the global journal
func Info(category, source, message string, data ...any) error {
	if j := GetGlobalJournal(); j != nil {
		return j.Info(category, source, message, data...)
	}
	return nil
}

// Warning logs a warning message using the global journal
func Warning(category, source, message string, data ...any) error {
	if j := GetGlobalJournal(); j != nil {
		return j.Warning(category, source, message, data...)
	}
	return nil
}

// Error logs an error message using the global journal
func Error(category, source, message string, data ...any) error {
	if j := GetGlobalJournal(); j != nil {
		return j.Error(category, source, message, data...)
	}
	return nil
}

// AccumulateUsage accumulates usage using the global journal
func AccumulateUsage(usageSessionId string, usage map[string]float64) {
	if j := GetGlobalJournal(); j != nil {
		j.AccumulateUsage(usageSessionId, usage)
	}
}

// GetUsage get usage
func GetUsage(usageSessionId string) map[string]float64 {
	if j := GetGlobalJournal(); j != nil {
		return j.GetUsage(usageSessionId)
	}
	return map[string]float64{}
}

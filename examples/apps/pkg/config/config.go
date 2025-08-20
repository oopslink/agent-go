package config

import (
	"encoding/json"
	"os"

	"github.com/oopslink/agent-go/pkg/commons/errors"
)

var (
	ErrorCodeLoadConfigFailed = errors.ErrorCode{
		Code:           92000,
		Name:           "LoadConfigFailed",
		DefaultMessage: "Failed to load config",
	}

	ErrorCodeInvalidConfig = errors.ErrorCode{
		Code:           92001,
		Name:           "InvalidConfig",
		DefaultMessage: "Config is not valid",
	}
)

// Config holds the complete configuration for the chat application
type Config struct {
	Provider            string `json:"provider"`
	APIKey              string `json:"api_key"`
	BaseURL             string `json:"base_url"`
	ModelName           string `json:"model_name"`
	OpenAICompatibility bool   `json:"openai_compatibility"`
	SystemPrompt        string `json:"system_prompt"`

	// Runtime directory for storing application state
	RuntimeDir string `json:"runtime_dir"`

	// Vector database configuration
	VectorDB VectorDBConfig `json:"vector_db"`

	// Embedder configuration
	Embedder EmbedderConfig `json:"embedder"`

	// Tool call suppression configuration
	ToolSuppression ToolSuppressionConfig `json:"tool_suppression"`
}

// ToolSuppressionConfig holds tool call suppression configuration
type ToolSuppressionConfig struct {
	Enabled        bool `json:"enabled"`         // Enable tool call suppression
	MaxCount       int  `json:"max_count"`       // Maximum consecutive calls for the same tool
	CooldownRounds int  `json:"cooldown_rounds"` // Number of rounds to skip tools after suppression
}

// VectorDBConfig holds vector database configuration
type VectorDBConfig struct {
	Type     string `json:"type"` // "milvus" or "mock"
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout"` // seconds
}

// EmbedderConfig holds embedder configuration
type EmbedderConfig struct {
	Type     string `json:"type"`     // "llm" or "mock"
	Provider string `json:"provider"` // "openai", "anthropic", etc.
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"` // embedding model name
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Errorf(ErrorCodeLoadConfigFailed, "failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Errorf(ErrorCodeLoadConfigFailed, "failed to parse config file: %v", err)
	}

	return &config, nil
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	if c.Provider == "" {
		return errors.Errorf(ErrorCodeInvalidConfig, "provider is required")
	}
	if c.APIKey == "" {
		return errors.Errorf(ErrorCodeInvalidConfig, "api_key is required")
	}
	if c.ModelName == "" {
		return errors.Errorf(ErrorCodeInvalidConfig, "model_name is required")
	}

	// Validate vector database config
	if c.VectorDB.Type == "" {
		c.VectorDB.Type = "mock" // Default to mock
	}
	if c.VectorDB.Type == "milvus" && c.VectorDB.Endpoint == "" {
		return errors.Errorf(ErrorCodeInvalidConfig, "vector_db.endpoint is required for milvus")
	}

	// Validate embedder config
	if c.Embedder.Type == "" {
		c.Embedder.Type = "llm" // Default to LLM embedder
	}
	if c.Embedder.Type == "llm" {
		if c.Embedder.Provider == "" {
			return errors.Errorf(ErrorCodeInvalidConfig, "embedder.provider is required for llm type")
		}
		if c.Embedder.APIKey == "" {
			return errors.Errorf(ErrorCodeInvalidConfig, "embedder.api_key is required for llm type")
		}
	}

	// Validate tool suppression config
	if !c.ToolSuppression.Enabled {
		// Set default values when not enabled
		c.ToolSuppression.MaxCount = 3
		c.ToolSuppression.CooldownRounds = 1
	} else {
		// Validate enabled config
		if c.ToolSuppression.MaxCount <= 0 {
			c.ToolSuppression.MaxCount = 3 // Default max count
		}
		if c.ToolSuppression.CooldownRounds <= 0 {
			c.ToolSuppression.CooldownRounds = 1 // Default cooldown rounds
		}
	}

	return nil
}

package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configData := `{
        "provider": "openai",
        "api_key": "test-key",
        "base_url": "https://api.openai.com/v1",
        "model_name": "gpt-4o-mini",
        "openai_compatibility": false,
        "system_prompt": "You are a helpful assistant",
        "vector_db": {
            "type": "milvus",
            "endpoint": "localhost:19530",
            "username": "",
            "password": "",
            "timeout": 15
        },
        "embedder": {
            "type": "llm",
            "provider": "openai",
            "api_key": "test-key",
            "base_url": "https://api.openai.com/v1",
            "model": "text-embedding-3-small"
        }
    }`

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configData); err != nil {
		t.Fatalf("Failed to write config data: %v", err)
	}
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", cfg.Provider)
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("Expected api_key 'test-key', got '%s'", cfg.APIKey)
	}
	if cfg.ModelName != "gpt-4o-mini" {
		t.Errorf("Expected model_name 'gpt-4o-mini', got '%s'", cfg.ModelName)
	}
	if cfg.VectorDB.Type != "milvus" {
		t.Errorf("Expected vector_db.type 'milvus', got '%s'", cfg.VectorDB.Type)
	}
	if cfg.Embedder.Type != "llm" {
		t.Errorf("Expected embedder.type 'llm', got '%s'", cfg.Embedder.Type)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Provider:  "openai",
				APIKey:    "test-key",
				ModelName: "gpt-4o-mini",
				VectorDB: VectorDBConfig{
					Type:     "milvus",
					Endpoint: "localhost:19530",
				},
				Embedder: EmbedderConfig{
					Type:     "llm",
					Provider: "openai",
					APIKey:   "test-key",
				},
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			config: &Config{
				APIKey:    "test-key",
				ModelName: "gpt-4o-mini",
			},
			wantErr: true,
		},
		{
			name: "missing api_key",
			config: &Config{
				Provider:  "openai",
				ModelName: "gpt-4o-mini",
			},
			wantErr: true,
		},
		{
			name: "missing model_name",
			config: &Config{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantErr: true,
		},
		{
			name: "milvus without endpoint",
			config: &Config{
				Provider:  "openai",
				APIKey:    "test-key",
				ModelName: "gpt-4o-mini",
				VectorDB: VectorDBConfig{
					Type: "milvus",
				},
			},
			wantErr: true,
		},
		{
			name: "llm embedder without provider",
			config: &Config{
				Provider:  "openai",
				APIKey:    "test-key",
				ModelName: "gpt-4o-mini",
				VectorDB: VectorDBConfig{
					Type: "milvus",
				},
				Embedder: EmbedderConfig{
					Type: "llm",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent-file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`{ invalid json }`); err != nil {
		t.Fatalf("Failed to write config data: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

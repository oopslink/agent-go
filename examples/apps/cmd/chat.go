package cmd

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	"github.com/oopslink/agent-go-apps/chat"
	"github.com/oopslink/agent-go-apps/pkg/config"
)

var (
	configFile string
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start the chat demo application",
	Long: `Start an interactive chat demo application that showcases how to use the agent-go framework.

Supports multiple AI model providers including OpenAI, Anthropic, Gemini, and more.
Configuration is loaded from a JSON file.

Examples:
  # Using configuration file
  demo chat --config config.json

  # Using default config path
  demo chat`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Set default config file if not provided
		if configFile == "" {
			configFile = "config.json"
		}

		// Check if config file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return errors.Errorf(ErrorCodeCmdExecError, "configuration file not found: %s", configFile)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			return errors.Errorf(ErrorCodeCmdExecError, "failed to load configuration: %v", err)
		}

		// Validate configuration
		if err := cfg.ValidateConfig(); err != nil {
			return errors.Errorf(ErrorCodeCmdExecError, "configuration validation failed: %v", err)
		}

		// Display current configuration
		fmt.Printf("# Starting chat with configuration:\n")
		fmt.Printf("  - Config File: %s\n", configFile)
		fmt.Printf("  - Provider: %s\n", cfg.Provider)
		fmt.Printf("  - Model: %s\n", cfg.ModelName)
		fmt.Printf("  - Base URL: %s\n", cfg.BaseURL)
		fmt.Printf("  - OpenAI Compatible: %t\n", cfg.OpenAICompatibility)
		fmt.Printf("  - Vector DB Type: %s\n", cfg.VectorDB.Type)
		fmt.Printf("  - Embedder Type: %s\n", cfg.Embedder.Type)
		fmt.Println()

		// Start chat application
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		return chat.RunChat(ctx, cfg)
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Command line flags
	chatCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path (default: config.json)")
}

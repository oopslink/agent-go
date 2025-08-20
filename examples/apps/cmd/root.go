package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/oopslink/agent-go/pkg/commons/errors"
)

var ErrorCodeCmdExecError = errors.ErrorCode{
	Code:           90000,
	Name:           "CmdExecError",
	DefaultMessage: "failed to exec",
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "demo",
	Short: "Agent-Go framework demo application",
	Long: `Agent-Go framework demo application

This application demonstrates how to use the agent-go framework to build intelligent conversational applications.
Features include:
- OpenAI-based chatbot
- Memory management and conversation history
- Configurable models and parameters`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}

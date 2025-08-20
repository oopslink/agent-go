package main

import (
	"fmt"
	"os"

	"github.com/oopslink/agent-go/pkg/support/journal"

	"github.com/oopslink/agent-go-apps/cmd"
)

func main() {
	initJournal()
	cmd.Execute()
}

func initJournal() {
	tmpfile, err := os.CreateTemp("", "journal_demo_*.log")
	if err != nil {
		fmt.Printf("Failed to create temporary journal file: %v\n", err)
		os.Exit(1)
	}

	j, err := journal.NewFileJournal(tmpfile.Name())
	if err != nil {
		fmt.Printf("Failed to create journal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Journal created at: %s\n", tmpfile.Name())
	journal.SetGlobalJournal(j)
}

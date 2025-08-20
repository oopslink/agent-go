package utils

import (
	"fmt"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"os"
)

func init() {
	InitJournal()
}

func InitJournal() {
	tmpfile, err := os.CreateTemp("", "journal_*.log")
	if err != nil {
		fmt.Printf("Failed to create temporary journal file: %v\n", err)
		os.Exit(1)
	}

	j, err := journal.NewFileJournal(tmpfile.Name())
	if err != nil {
		fmt.Printf("Failed to create journal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("# Journal created at: %s\n", tmpfile.Name())
	journal.SetGlobalJournal(j)
}

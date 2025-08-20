package ui

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"
	"time"
)

var _inputReader *bufio.Reader
var _outputWriter *os.File

func init() {
	_inputReader = bufio.NewReader(os.Stdin)
	_outputWriter = os.Stdout
}

// Input print prompt and return user input in stdio
func Input(prompt string) string {
	for {
		fmt.Println()
		fmt.Print(prompt + " ")
		input, err := _inputReader.ReadString('\n')
		if err != nil {
			PrintWarning("Cannot get input, err: " + err.Error())
			continue
		}
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}
		return input
	}
}

// Print print content to stdout
func Print(content string) {
	_, _ = _outputWriter.WriteString(content)
}

func PrintKVs(title string, kv map[string]float64) {
	if len(kv) == 0 {
		return
	}
	if len(title) >= 45 {
		title = title[0:45] + "..."
	}
	_, _ = _outputWriter.WriteString(fmt.Sprintf("\n- %s %s\n", title, strings.Repeat("-", 50-len(title))))
	keys := maps.Keys(kv)
	for key := range keys {
		_, _ = _outputWriter.WriteString(fmt.Sprintf("%25s: %v\n", key, kv[key]))
	}
}

// PrintWarning print a line with warn icon to stdout
func PrintWarning(content string) {
	_, _ = _outputWriter.WriteString(fmt.Sprintf("\n> [WARNING] %s\n", content))
}

// PrintSeparator print separator
func PrintSeparator(c string) {
	_, _ = _outputWriter.WriteString(fmt.Sprintf("%s\n", strings.Repeat(c, 60)))
}

// Confirm print choices and return the index of choice user chosen
func Confirm(tip string, choices []string) int {
	for {
		fmt.Println()
		fmt.Print(tip)
		for idx, choice := range choices {
			fmt.Printf(" * [%d] %s\n", idx+1, choice)
		}
		fmt.Print(fmt.Sprintf("> Which would you like to do (1 ... %d): ", len(choices)))
		input, err := _inputReader.ReadString('\n')
		if err != nil {
			PrintWarning("Cannot get input, err: " + err.Error())
			continue
		}
		input = strings.TrimSpace(input)
		if len(input) == 0 {
			continue
		}
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(choices) {
			PrintWarning(fmt.Sprintf("Invalid choice, should be number between: [%d ... %d]", 1, len(choices)))
			continue
		}
		return idx
	}
}

// Loading print progress loading animation
func Loading(info string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := 0

	go func() {
		for {
			select {
			case <-ctx.Done():
				// clear the loading line
				fmt.Print("\r\033[K")
				_outputWriter.Sync()
				return
			default:
				fmt.Printf("\r%s %s ...", spinner[idx], info)
				_outputWriter.Sync()
				idx = (idx + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return cancel
}

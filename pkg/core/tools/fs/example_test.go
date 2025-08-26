package fs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// Example demonstrates how to use the file system tools
func ExampleFileSystemTools() {
	// Create temporary directory for this example
	tempDir, err := os.MkdirTemp("", "fs_example_*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create file system tools instance
	fst, err := NewFileSystemTools(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	// Get all available tools
	allTools := fst.GetTools()
	
	// Create a tool collection for easy access
	toolCollection := tools.OfTools(allTools...)

	ctx := context.Background()

	// Example 1: Create a directory
	fmt.Println("=== Creating a directory ===")
	createDirResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_1",
		Name:       "fs_create_directory",
		Arguments: map[string]any{
			"path": "example_project",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Create directory result: %+v\n", createDirResult.Result)

	// Example 2: Create a metadata file for the directory
	fmt.Println("\n=== Creating directory metadata ===")
	metaContent := `name: Example Project
description: A sample project for demonstrating file system tools
tags:
  - example
  - demo
  - filesystem
properties:
  author: file-system-tools
  version: "1.0.0"
  language: go
`
	writeMetaResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_2",
		Name:       "fs_write_file",
		Arguments: map[string]any{
			"path":    "example_project/.meta.yaml",
			"content": metaContent,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Write metadata result: %+v\n", writeMetaResult.Result)

	// Example 3: Create some example files
	fmt.Println("\n=== Creating example files ===")
	files := map[string]string{
		"example_project/main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
		"example_project/README.md": `# Example Project

This is a sample project created using the file system tools.

## Features

- File creation
- Directory listing
- Metadata support

## Usage

Run the main.go file to see the output.
`,
		"example_project/config.json": `{
	"name": "example-project",
	"version": "1.0.0",
	"description": "A sample project",
	"author": "file-system-tools"
}
`,
	}

	for filePath, content := range files {
		result, err := toolCollection.Call(ctx, &llms.ToolCall{
			ToolCallId: fmt.Sprintf("create_%s", filepath.Base(filePath)),
			Name:       "fs_write_file",
			Arguments: map[string]any{
				"path":    filePath,
				"content": content,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created file %s: %+v\n", filePath, result.Result)
	}

	// Example 4: List directory with metadata
	fmt.Println("\n=== Listing directory with metadata ===")
	listResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_list",
		Name:       "fs_list_directory",
		Arguments: map[string]any{
			"path":  "example_project",
			"depth": 0,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Directory listing: %+v\n", listResult.Result)

	// Example 5: Get file statistics
	fmt.Println("\n=== Getting file statistics ===")
	statResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_stat",
		Name:       "fs_get_file_stat",
		Arguments: map[string]any{
			"path": "example_project/main.go",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File stat: %+v\n", statResult.Result)

	// Example 6: Read a file
	fmt.Println("\n=== Reading file content ===")
	readResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_read",
		Name:       "fs_read_file",
		Arguments: map[string]any{
			"path": "example_project/README.md",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Read file result: %+v\n", readResult.Result)

	// Example 7: Create subdirectory and files
	fmt.Println("\n=== Creating subdirectory structure ===")
	subDirResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_subdir",
		Name:       "fs_create_directory",
		Arguments: map[string]any{
			"path": "example_project/src",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Create subdirectory result: %+v\n", subDirResult.Result)

	// Create a file in subdirectory
	writeSubFileResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_subfile",
		Name:       "fs_write_file",
		Arguments: map[string]any{
			"path": "example_project/src/utils.go",
			"content": `package main

import "fmt"

func printMessage(msg string) {
	fmt.Println(msg)
}
`,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Create subfile result: %+v\n", writeSubFileResult.Result)

	// Example 8: List directory recursively
	fmt.Println("\n=== Listing directory recursively ===")
	recursiveListResult, err := toolCollection.Call(ctx, &llms.ToolCall{
		ToolCallId: "example_recursive",
		Name:       "fs_list_directory",
		Arguments: map[string]any{
			"path":  "example_project",
			"depth": 2, // List up to 2 levels deep
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Recursive directory listing: %+v\n", recursiveListResult.Result)

	fmt.Println("\n=== Example completed successfully! ===")
}

// ExampleGetToolDescriptors shows how to get tool descriptors for AI integration
func ExampleGetToolDescriptors() {
	tempDir, err := os.MkdirTemp("", "fs_descriptors_*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create file system tools instance
	fst, err := NewFileSystemTools(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	// Get all tools and their descriptors
	allTools := fst.GetTools()
	toolCollection := tools.OfTools(allTools...)

	fmt.Println("Available File System Tools:")
	fmt.Println("============================")

	descriptors := toolCollection.Descriptors()
	for i, descriptor := range descriptors {
		fmt.Printf("\n%d. %s\n", i+1, descriptor.Name)
		fmt.Printf("   Description: %s\n", descriptor.Description)
		
		if descriptor.Parameters != nil {
			fmt.Printf("   Parameters:\n")
			for paramName, paramSchema := range descriptor.Parameters.Properties {
				required := ""
				for _, req := range descriptor.Parameters.Required {
					if req == paramName {
						required = " (required)"
						break
					}
				}
				fmt.Printf("     - %s: %s%s - %s\n", 
					paramName, 
					paramSchema.Type, 
					required,
					paramSchema.Description)
			}
		}
	}

	fmt.Println("\nTool descriptors can be used to integrate with AI agents and LLMs.")
}

package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemTools(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs_tools_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create file system tools instance
	fst, err := NewFileSystemTools(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("CreateDirectory", func(t *testing.T) {
		tool := NewCreateDirectoryTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_1",
			Name:       "fs_create_directory",
			Arguments: map[string]any{
				"path": "test_dir",
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// Verify directory was created
		dirPath := filepath.Join(tempDir, "test_dir")
		stat, err := os.Stat(dirPath)
		require.NoError(t, err)
		assert.True(t, stat.IsDir())
	})

	t.Run("CreateFile", func(t *testing.T) {
		tool := NewCreateFileTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_2",
			Name:       "fs_create_file",
			Arguments: map[string]any{
				"path": "test_dir/test_file.txt",
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// Verify file was created
		filePath := filepath.Join(tempDir, "test_dir", "test_file.txt")
		stat, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.False(t, stat.IsDir())
	})

	t.Run("WriteFile", func(t *testing.T) {
		tool := NewWriteFileTool(fst.rootPath)

		content := "Hello, World!"
		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_3",
			Name:       "fs_write_file",
			Arguments: map[string]any{
				"path":    "test_dir/write_test.txt",
				"content": content,
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// Handle the size field which might be int or float64
		sizeValue := result.Result["size"]
		var size int
		switch v := sizeValue.(type) {
		case int:
			size = v
		case float64:
			size = int(v)
		default:
			t.Fatalf("Unexpected size type: %T", v)
		}
		assert.Equal(t, len(content), size)

		// Verify file content
		filePath := filepath.Join(tempDir, "test_dir", "write_test.txt")
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("ReadFile", func(t *testing.T) {
		tool := NewReadFileTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_4",
			Name:       "fs_read_file",
			Arguments: map[string]any{
				"path": "test_dir/write_test.txt",
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))
		assert.Equal(t, "Hello, World!", result.Result["content"].(string))
	})

	t.Run("GetFileStat", func(t *testing.T) {
		tool := NewGetFileStatTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_5",
			Name:       "fs_get_file_stat",
			Arguments: map[string]any{
				"path": "test_dir/write_test.txt",
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// The stat field should be accessible through the result map
		statData := result.Result["stat"]

		// Convert to map for easier testing
		statBytes, err := json.Marshal(statData)
		require.NoError(t, err)

		var stat map[string]interface{}
		err = json.Unmarshal(statBytes, &stat)
		require.NoError(t, err)

		assert.True(t, stat["exists"].(bool))
		assert.False(t, stat["is_dir"].(bool))
		assert.Equal(t, "write_test.txt", stat["name"].(string))
	})

	t.Run("ListDirectory", func(t *testing.T) {
		tool := NewListDirectoryTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_6",
			Name:       "fs_list_directory",
			Arguments: map[string]any{
				"path":  "test_dir",
				"depth": 0,
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// Convert listing data to map for easier testing
		listingData := result.Result["listing"]
		listingBytes, err := json.Marshal(listingData)
		require.NoError(t, err)

		var listing map[string]interface{}
		err = json.Unmarshal(listingBytes, &listing)
		require.NoError(t, err)

		assert.Equal(t, "test_dir", listing["path"].(string))

		files := listing["files"].([]interface{})
		assert.Len(t, files, 2) // test_file.txt and write_test.txt
	})

	t.Run("DeleteFile", func(t *testing.T) {
		tool := NewDeleteFileTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_7",
			Name:       "fs_delete_file",
			Arguments: map[string]any{
				"path": "test_dir/test_file.txt",
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// Verify file was deleted
		filePath := filepath.Join(tempDir, "test_dir", "test_file.txt")
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("PathValidation", func(t *testing.T) {
		tool := NewReadFileTool(fst.rootPath)

		// Try to access a path outside the root directory
		_, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_8",
			Name:       "fs_read_file",
			Arguments: map[string]any{
				"path": "../../../etc/passwd",
			},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "outside the allowed root directory")
	})
}

func TestDirectoryMeta(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs_meta_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create file system tools instance
	fst, err := NewFileSystemTools(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Create a directory with metadata
	dirPath := filepath.Join(tempDir, "meta_test")
	err = os.MkdirAll(dirPath, 0755)
	require.NoError(t, err)

	// Create .meta.yaml file
	metaContent := `name: Test Directory
description: A directory for testing metadata
tags:
  - test
  - example
properties:
  owner: test_user
  version: "1.0"
`
	metaPath := filepath.Join(dirPath, ".meta.yaml")
	err = os.WriteFile(metaPath, []byte(metaContent), 0644)
	require.NoError(t, err)

	t.Run("ListDirectoryWithMeta", func(t *testing.T) {
		tool := NewListDirectoryTool(fst.rootPath)

		result, err := tool.Call(ctx, &llms.ToolCall{
			ToolCallId: "test_meta",
			Name:       "fs_list_directory",
			Arguments: map[string]any{
				"path":  "meta_test",
				"depth": 0,
			},
		})

		require.NoError(t, err)
		assert.True(t, result.Result["success"].(bool))

		// Convert listing data to map for easier testing
		listingData := result.Result["listing"]
		listingBytes, err := json.Marshal(listingData)
		require.NoError(t, err)

		var listing map[string]interface{}
		err = json.Unmarshal(listingBytes, &listing)
		require.NoError(t, err)

		// Check if meta exists
		metaData, hasKey := listing["meta"]
		require.True(t, hasKey, "meta key should exist in listing")
		require.NotNil(t, metaData, "meta should not be nil")

		meta := metaData.(map[string]interface{})

		// Use the correct field names (capitalized in JSON)
		assert.Equal(t, "Test Directory", meta["Name"].(string))
		assert.Equal(t, "A directory for testing metadata", meta["Description"].(string))

		tags := meta["Tags"].([]interface{})
		assert.Len(t, tags, 2)
		assert.Contains(t, tags, "test")
		assert.Contains(t, tags, "example")

		properties := meta["Properties"].(map[string]interface{})
		assert.Equal(t, "test_user", properties["owner"].(string))
		assert.Equal(t, "1.0", properties["version"].(string))
	})
}

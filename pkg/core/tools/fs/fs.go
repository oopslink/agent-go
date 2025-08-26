package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
	"gopkg.in/yaml.v3"
)

// DirectoryMeta represents metadata for a directory from .meta.yaml file
type DirectoryMeta struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Properties  map[string]string `yaml:"properties,omitempty"`
}

// FileSystemTools provides a collection of file system tools with a root path restriction
type FileSystemTools struct {
	rootPath string
}

// NewFileSystemTools creates a new file system tools instance with the specified root path
func NewFileSystemTools(rootPath string) (*FileSystemTools, error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for root: %w", err)
	}

	// Ensure the root directory exists
	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	return &FileSystemTools{
		rootPath: absRoot,
	}, nil
}

// GetTools returns all available file system tools
func (fst *FileSystemTools) GetTools() []tools.Tool {
	return []tools.Tool{
		NewListDirectoryTool(fst.rootPath),
		NewGetFileStatTool(fst.rootPath),
		NewReadFileTool(fst.rootPath),
		NewWriteFileTool(fst.rootPath),
		NewCreateFileTool(fst.rootPath),
		NewDeleteFileTool(fst.rootPath),
		NewCreateDirectoryTool(fst.rootPath),
	}
}

// validatePath ensures the path is within the root directory
func (fst *FileSystemTools) validatePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	absPath := filepath.Join(fst.rootPath, cleanPath)

	// Resolve any symlinks to prevent directory traversal
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// If file doesn't exist, use the original absolute path for validation
	if os.IsNotExist(err) {
		resolvedPath = absPath
	}

	// Ensure the resolved path is still within the root
	if !strings.HasPrefix(resolvedPath, fst.rootPath) {
		return "", fmt.Errorf("path '%s' is outside the allowed root directory", path)
	}

	return absPath, nil
}

// loadDirectoryMeta loads metadata from .meta.yaml file in the directory
func (fst *FileSystemTools) loadDirectoryMeta(dirPath string) (*DirectoryMeta, error) {
	metaPath := filepath.Join(dirPath, ".meta.yaml")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata file exists
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta DirectoryMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	return &meta, nil
}

// ===== List Directory Tool =====

type ListDirectoryTool struct {
	rootPath string
}

func NewListDirectoryTool(rootPath string) *ListDirectoryTool {
	return &ListDirectoryTool{rootPath: rootPath}
}

var _ tools.Tool = &ListDirectoryTool{}

type ListDirectoryParams struct {
	Path  string `json:"path"`
	Depth int    `json:"depth"`
}

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type DirectoryListing struct {
	Path     string             `json:"path"`
	Meta     *DirectoryMeta     `json:"meta,omitempty"`
	Files    []FileInfo         `json:"files"`
	Children []DirectoryListing `json:"children,omitempty"`
}

func (t *ListDirectoryTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_list_directory",
		Description: "List directory contents with support for recursive listing. Can load .meta.yaml metadata if present.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to list (default: '.' for root)",
				},
				"depth": {
					Type:        llms.TypeInteger,
					Description: "Maximum depth for recursive listing (0 = current directory only, -1 = unlimited)",
				},
			},
		},
	}
}

func (t *ListDirectoryTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var listParams ListDirectoryParams
	if err := mapToStruct(params.Arguments, &listParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if listParams.Path == "" {
		listParams.Path = "."
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	listing, err := t.listDirectory(fst, listParams.Path, listParams.Depth)
	if err != nil {
		return nil, err
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"listing": listing,
		},
	}, nil
}

func (t *ListDirectoryTool) listDirectory(fst *FileSystemTools, path string, maxDepth int) (*DirectoryListing, error) {
	absPath, err := fst.validatePath(path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory: %w", err)
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("path '%s' is not a directory", path)
	}

	listing := &DirectoryListing{
		Path:  path,
		Files: make([]FileInfo, 0),
	}

	// Load metadata if available
	meta, _ := fst.loadDirectoryMeta(absPath)
	listing.Meta = meta

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		// Skip .meta.yaml files in listing
		if entry.Name() == ".meta.yaml" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format(time.RFC3339),
		}

		listing.Files = append(listing.Files, fileInfo)

		// Recurse into subdirectories if depth allows
		if entry.IsDir() && (maxDepth < 0 || maxDepth > 0) {
			nextDepth := maxDepth
			if maxDepth > 0 {
				nextDepth = maxDepth - 1
			}

			childListing, err := t.listDirectory(fst, fileInfo.Path, nextDepth)
			if err == nil {
				if listing.Children == nil {
					listing.Children = make([]DirectoryListing, 0)
				}
				listing.Children = append(listing.Children, *childListing)
			}
		}
	}

	return listing, nil
}

// ===== Get File Stat Tool =====

type GetFileStatTool struct {
	rootPath string
}

func NewGetFileStatTool(rootPath string) *GetFileStatTool {
	return &GetFileStatTool{rootPath: rootPath}
}

var _ tools.Tool = &GetFileStatTool{}

type FileStatParams struct {
	Path string `json:"path"`
}

type FileStatResult struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	Exists  bool   `json:"exists"`
}

func (t *GetFileStatTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_get_file_stat",
		Description: "Get detailed information about a file or directory.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to the file/directory",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *GetFileStatTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var statParams FileStatParams
	if err := mapToStruct(params.Arguments, &statParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	absPath, err := fst.validatePath(statParams.Path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(absPath)
	result := FileStatResult{
		Path:   statParams.Path,
		Exists: err == nil,
	}

	if err != nil {
		if os.IsNotExist(err) {
			return &llms.ToolCallResult{
				ToolCallId: params.ToolCallId,
				Name:       params.Name,
				Result: map[string]any{
					"success": true,
					"stat":    result,
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	result.Name = stat.Name()
	result.IsDir = stat.IsDir()
	result.Size = stat.Size()
	result.Mode = stat.Mode().String()
	result.ModTime = stat.ModTime().Format(time.RFC3339)

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"stat":    result,
		},
	}, nil
}

// ===== Read File Tool =====

type ReadFileTool struct {
	rootPath string
}

func NewReadFileTool(rootPath string) *ReadFileTool {
	return &ReadFileTool{rootPath: rootPath}
}

var _ tools.Tool = &ReadFileTool{}

type ReadFileParams struct {
	Path string `json:"path"`
}

func (t *ReadFileTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_read_file",
		Description: "Read the contents of a file.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to the file to read",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *ReadFileTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var readParams ReadFileParams
	if err := mapToStruct(params.Arguments, &readParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	absPath, err := fst.validatePath(readParams.Path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"path":    readParams.Path,
			"content": string(content),
			"size":    len(content),
		},
	}, nil
}

// ===== Write File Tool =====

type WriteFileTool struct {
	rootPath string
}

func NewWriteFileTool(rootPath string) *WriteFileTool {
	return &WriteFileTool{rootPath: rootPath}
}

var _ tools.Tool = &WriteFileTool{}

type WriteFileParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    string `json:"mode,omitempty"`
}

func (t *WriteFileTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_write_file",
		Description: "Write content to a file. Creates directories as needed.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to the file to write",
				},
				"content": {
					Type:        llms.TypeString,
					Description: "Content to write to the file",
				},
				"mode": {
					Type:        llms.TypeString,
					Description: "File permissions in octal format (default: '0644')",
				},
			},
			Required: []string{"path", "content"},
		},
	}
}

func (t *WriteFileTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var writeParams WriteFileParams
	if err := mapToStruct(params.Arguments, &writeParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	absPath, err := fst.validatePath(writeParams.Path)
	if err != nil {
		return nil, err
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Parse file mode
	mode := fs.FileMode(0644)
	if writeParams.Mode != "" {
		var modeInt uint64
		if _, err := fmt.Sscanf(writeParams.Mode, "%o", &modeInt); err != nil {
			return nil, fmt.Errorf("invalid file mode '%s': %w", writeParams.Mode, err)
		}
		mode = fs.FileMode(modeInt)
	}

	if err := os.WriteFile(absPath, []byte(writeParams.Content), mode); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"path":    writeParams.Path,
			"size":    len(writeParams.Content),
		},
	}, nil
}

// ===== Create File Tool =====

type CreateFileTool struct {
	rootPath string
}

func NewCreateFileTool(rootPath string) *CreateFileTool {
	return &CreateFileTool{rootPath: rootPath}
}

var _ tools.Tool = &CreateFileTool{}

type CreateFileParams struct {
	Path string `json:"path"`
	Mode string `json:"mode,omitempty"`
}

func (t *CreateFileTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_create_file",
		Description: "Create an empty file. Creates directories as needed.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to the file to create",
				},
				"mode": {
					Type:        llms.TypeString,
					Description: "File permissions in octal format (default: '0644')",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *CreateFileTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var createParams CreateFileParams
	if err := mapToStruct(params.Arguments, &createParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	absPath, err := fst.validatePath(createParams.Path)
	if err != nil {
		return nil, err
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Parse file mode
	mode := fs.FileMode(0644)
	if createParams.Mode != "" {
		var modeInt uint64
		if _, err := fmt.Sscanf(createParams.Mode, "%o", &modeInt); err != nil {
			return nil, fmt.Errorf("invalid file mode '%s': %w", createParams.Mode, err)
		}
		mode = fs.FileMode(modeInt)
	}

	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("file already exists: %s", createParams.Path)
		}
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	file.Close()

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"path":    createParams.Path,
		},
	}, nil
}

// ===== Delete File Tool =====

type DeleteFileTool struct {
	rootPath string
}

func NewDeleteFileTool(rootPath string) *DeleteFileTool {
	return &DeleteFileTool{rootPath: rootPath}
}

var _ tools.Tool = &DeleteFileTool{}

type DeleteFileParams struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
}

func (t *DeleteFileTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_delete_file",
		Description: "Delete a file or directory.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to the file/directory to delete",
				},
				"recursive": {
					Type:        llms.TypeBoolean,
					Description: "Whether to delete directories recursively (default: false)",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *DeleteFileTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var deleteParams DeleteFileParams
	if err := mapToStruct(params.Arguments, &deleteParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	absPath, err := fst.validatePath(deleteParams.Path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &llms.ToolCallResult{
				ToolCallId: params.ToolCallId,
				Name:       params.Name,
				Result: map[string]any{
					"success": true,
					"path":    deleteParams.Path,
					"message": "file does not exist",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if stat.IsDir() && !deleteParams.Recursive {
		// Check if directory is empty
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}
		if len(entries) > 0 {
			return nil, fmt.Errorf("directory is not empty, use recursive=true to delete")
		}
	}

	var deleteErr error
	if deleteParams.Recursive {
		deleteErr = os.RemoveAll(absPath)
	} else {
		deleteErr = os.Remove(absPath)
	}

	if deleteErr != nil {
		return nil, fmt.Errorf("failed to delete: %w", deleteErr)
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"path":    deleteParams.Path,
		},
	}, nil
}

// ===== Create Directory Tool =====

type CreateDirectoryTool struct {
	rootPath string
}

func NewCreateDirectoryTool(rootPath string) *CreateDirectoryTool {
	return &CreateDirectoryTool{rootPath: rootPath}
}

var _ tools.Tool = &CreateDirectoryTool{}

type CreateDirectoryParams struct {
	Path string `json:"path"`
	Mode string `json:"mode,omitempty"`
}

func (t *CreateDirectoryTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "fs_create_directory",
		Description: "Create a directory. Creates parent directories as needed.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"path": {
					Type:        llms.TypeString,
					Description: "Relative path from root directory to the directory to create",
				},
				"mode": {
					Type:        llms.TypeString,
					Description: "Directory permissions in octal format (default: '0755')",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (t *CreateDirectoryTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var createParams CreateDirectoryParams
	if err := mapToStruct(params.Arguments, &createParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	fst := &FileSystemTools{rootPath: t.rootPath}
	absPath, err := fst.validatePath(createParams.Path)
	if err != nil {
		return nil, err
	}

	// Parse directory mode
	mode := fs.FileMode(0755)
	if createParams.Mode != "" {
		var modeInt uint64
		if _, err := fmt.Sscanf(createParams.Mode, "%o", &modeInt); err != nil {
			return nil, fmt.Errorf("invalid directory mode '%s': %w", createParams.Mode, err)
		}
		mode = fs.FileMode(modeInt)
	}

	if err := os.MkdirAll(absPath, mode); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"path":    createParams.Path,
		},
	}, nil
}

// ===== Utility Functions =====

// mapToStruct converts a map[string]any to a struct using JSON marshaling/unmarshaling
func mapToStruct(m map[string]any, target interface{}) error {
	if m == nil {
		return nil
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return nil
}

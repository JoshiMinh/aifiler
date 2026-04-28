package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileInfo represents information about a file.
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	IsDir       bool      `json:"is_dir"`
	ModTime     time.Time `json:"mod_time"`
	IsSymlink   bool      `json:"is_symlink"`
	SymlinkDest string    `json:"symlink_dest,omitempty"`
}

// InitFileOperationsTools registers file system operation tools on the executor.
func InitFileOperationsTools(te *ToolExecutor) {
	// read_file
	te.RegisterTool(
		ToolDefinition{
			Name:        "read_file",
			Description: "Read content from a file with optional line range support",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "Absolute or relative file path to read",
					Required:    true,
				},
				{
					Name:        "start_line",
					Type:        "int",
					Description: "Start reading from this line (1-indexed). Optional, defaults to 1.",
					Required:    false,
					Default:     1,
				},
				{
					Name:        "end_line",
					Type:        "int",
					Description: "End reading at this line (1-indexed). Optional, defaults to end of file.",
					Required:    false,
				},
				{
					Name:        "max_size",
					Type:        "int",
					Description: "Maximum bytes to read. Optional, defaults to 1MB.",
					Required:    false,
					Default:     1048576,
				},
			},
		},
		te.executeReadFile,
	)

	// write_file
	te.RegisterTool(
		ToolDefinition{
			Name:        "write_file",
			Description: "Write content to a file (creates if not exists, overwrites if exists)",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "Absolute or relative file path to write to",
					Required:    true,
				},
				{
					Name:        "content",
					Type:        "string",
					Description: "Content to write to the file",
					Required:    true,
				},
				{
					Name:        "mode",
					Type:        "string",
					Description: "File permissions (e.g., '0644'). Optional, defaults to '0644'.",
					Required:    false,
					Default:     "0644",
				},
			},
		},
		te.executeWriteFile,
	)

	// list_directory
	te.RegisterTool(
		ToolDefinition{
			Name:        "list_directory",
			Description: "List files and subdirectories in a directory",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "Directory path to list",
					Required:    true,
				},
				{
					Name:        "recursive",
					Type:        "bool",
					Description: "Recursively list subdirectories. Optional, defaults to false.",
					Required:    false,
					Default:     false,
				},
				{
					Name:        "max_depth",
					Type:        "int",
					Description: "Maximum recursion depth. Optional, defaults to -1 (unlimited).",
					Required:    false,
					Default:     -1,
				},
			},
		},
		te.executeListDirectory,
	)

	// check_file_exists
	te.RegisterTool(
		ToolDefinition{
			Name:        "check_file_exists",
			Description: "Check if a file or directory exists",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "File or directory path to check",
					Required:    true,
				},
			},
		},
		te.executeCheckFileExists,
	)

	// get_file_metadata
	te.RegisterTool(
		ToolDefinition{
			Name:        "get_file_metadata",
			Description: "Get detailed metadata about a file (size, mtime, permissions, etc.)",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "File or directory path",
					Required:    true,
				},
			},
		},
		te.executeGetFileMetadata,
	)

	// get_file_type
	te.RegisterTool(
		ToolDefinition{
			Name:        "get_file_type",
			Description: "Detect file type by extension and content (basic detection)",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "File path to analyze",
					Required:    true,
				},
			},
		},
		te.executeGetFileType,
	)

	// copy_file
	te.RegisterTool(
		ToolDefinition{
			Name:        "copy_file",
			Description: "Copy a file from source to destination",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "src",
					Type:        "string",
					Description: "Source file path",
					Required:    true,
				},
				{
					Name:        "dst",
					Type:        "string",
					Description: "Destination file path",
					Required:    true,
				},
				{
					Name:        "overwrite",
					Type:        "bool",
					Description: "Overwrite destination if it exists. Optional, defaults to false.",
					Required:    false,
					Default:     false,
				},
			},
		},
		te.executeCopyFile,
	)

	// delete_file
	te.RegisterTool(
		ToolDefinition{
			Name:        "delete_file",
			Description: "Delete a file or directory recursively",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "File or directory path to delete",
					Required:    true,
				},
			},
		},
		te.executeDeleteFile,
	)

	// create_directory
	te.RegisterTool(
		ToolDefinition{
			Name:        "create_directory",
			Description: "Create a directory with all parent directories if needed",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "Directory path to create",
					Required:    true,
				},
			},
		},
		te.executeCreateDirectory,
	)

	// rename_file
	te.RegisterTool(
		ToolDefinition{
			Name:        "rename_file",
			Description: "Rename or move a file to a new path",
			Category:    "file_ops",
			Parameters: []ToolParam{
				{
					Name:        "from",
					Type:        "string",
					Description: "Current file path",
					Required:    true,
				},
				{
					Name:        "to",
					Type:        "string",
					Description: "New file path",
					Required:    true,
				},
			},
		},
		te.executeRenameFile,
	)
}

// ============================================================================
// File System Operation Executors
// ============================================================================

func (te *ToolExecutor) executeReadFile(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)

	startLine := 1
	if sl, ok := params["start_line"].(float64); ok {
		startLine = int(sl)
	}

	endLine := -1
	if el, ok := params["end_line"].(float64); ok {
		endLine = int(el)
	}

	maxSize := 1048576 // 1MB
	if ms, ok := params["max_size"].(float64); ok {
		maxSize = int(ms)
	}

	content, err := readFileRange(fullPath, startLine, endLine, maxSize)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to read file: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":    path,
			"content": content,
			"lines":   strings.Count(content, "\n") + 1,
		},
	}
}

func (te *ToolExecutor) executeWriteFile(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	content, ok := params["content"].(string)
	if !ok {
		return ToolResult{Success: false, Error: "content parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create directories: %v", err)}
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to write file: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":   path,
			"size":   len(content),
			"status": "written successfully",
		},
	}
}

func (te *ToolExecutor) executeListDirectory(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		path = "."
	}

	fullPath := resolvePath(te.cwd, path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to list directory: %v", err)}
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		}
		files = append(files, fileInfo)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":  path,
			"count": len(files),
			"files": files,
		},
	}
}

func (te *ToolExecutor) executeCheckFileExists(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)
	_, err := os.Stat(fullPath)

	exists := err == nil
	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":   path,
			"exists": exists,
			"error":  getErrorString(err),
		},
	}
}

func (te *ToolExecutor) executeGetFileMetadata(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to stat file: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"name":       info.Name(),
			"size":       info.Size(),
			"mode":       info.Mode().String(),
			"is_dir":     info.IsDir(),
			"mod_time":   info.ModTime(),
			"is_symlink": info.Mode()&os.ModeSymlink != 0,
		},
	}
}

func (te *ToolExecutor) executeGetFileType(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)

	// Get file extension
	ext := strings.ToLower(filepath.Ext(fullPath))

	// Try to read first few bytes for magic number detection
	file, err := os.Open(fullPath)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to open file: %v", err)}
	}
	defer file.Close()

	header := make([]byte, 512)
	n, _ := file.Read(header)

	fileType := detectFileType(ext, header[:n])

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"extension": ext,
			"type":      fileType,
			"detected":  true,
		},
	}
}

func (te *ToolExecutor) executeCopyFile(ctx context.Context, params map[string]interface{}) ToolResult {
	src, ok := params["src"].(string)
	if !ok || src == "" {
		return ToolResult{Success: false, Error: "src parameter is required"}
	}

	dst, ok := params["dst"].(string)
	if !ok || dst == "" {
		return ToolResult{Success: false, Error: "dst parameter is required"}
	}

	overwrite := false
	if ow, ok := params["overwrite"].(bool); ok {
		overwrite = ow
	}

	fullSrc := resolvePath(te.cwd, src)
	fullDst := resolvePath(te.cwd, dst)

	// Check if destination exists
	if _, err := os.Stat(fullDst); err == nil && !overwrite {
		return ToolResult{Success: false, Error: "destination file already exists (set overwrite=true to overwrite)"}
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(fullDst), 0o755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create destination directory: %v", err)}
	}

	// Copy file
	srcFile, err := os.Open(fullSrc)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to open source file: %v", err)}
	}
	defer srcFile.Close()

	dstFile, err := os.Create(fullDst)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create destination file: %v", err)}
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to copy file: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"src":    src,
			"dst":    dst,
			"status": "copied successfully",
		},
	}
}

func (te *ToolExecutor) executeDeleteFile(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)

	if err := os.RemoveAll(fullPath); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to delete: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":   path,
			"status": "deleted successfully",
		},
	}
}

func (te *ToolExecutor) executeCreateDirectory(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)

	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create directory: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":   path,
			"status": "created successfully",
		},
	}
}

func (te *ToolExecutor) executeRenameFile(ctx context.Context, params map[string]interface{}) ToolResult {
	from, ok := params["from"].(string)
	if !ok || from == "" {
		return ToolResult{Success: false, Error: "from parameter is required"}
	}

	to, ok := params["to"].(string)
	if !ok || to == "" {
		return ToolResult{Success: false, Error: "to parameter is required"}
	}

	fullFrom := resolvePath(te.cwd, from)
	fullTo := resolvePath(te.cwd, to)

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(fullTo), 0o755); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to create destination directory: %v", err)}
	}

	if err := os.Rename(fullFrom, fullTo); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to rename file: %v", err)}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"from":   from,
			"to":     to,
			"status": "renamed successfully",
		},
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func readFileRange(filePath string, startLine, endLine, maxSize int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Limit read size
	reader := &io.LimitedReader{R: file, N: int64(maxSize)}

	content, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return "", err
	}

	lines := strings.Split(string(content), "\n")

	// Adjust for 1-indexing
	if startLine < 1 {
		startLine = 1
	}
	if endLine < 0 {
		endLine = len(lines)
	}
	if startLine > len(lines) {
		startLine = len(lines)
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	selectedLines := lines[startLine-1 : endLine]
	return strings.Join(selectedLines, "\n"), nil
}

func resolvePath(cwd, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(cwd, relPath)
}

func detectFileType(ext string, header []byte) string {
	// Magic number detection
	if len(header) >= 4 {
		// PNG: 89 50 4E 47
		if bytes.HasPrefix(header, []byte{0x89, 0x50, 0x4E, 0x47}) {
			return "image/png"
		}
		// JPEG: FF D8 FF
		if bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}) {
			return "image/jpeg"
		}
		// GIF: 47 49 46 38
		if bytes.HasPrefix(header, []byte{0x47, 0x49, 0x46, 0x38}) {
			return "image/gif"
		}
		// PDF: 25 50 44 46
		if bytes.HasPrefix(header, []byte{0x25, 0x50, 0x44, 0x46}) {
			return "application/pdf"
		}
	}

	// Extension-based detection
	typeMap := map[string]string{
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".yaml": "application/yaml",
		".yml":  "application/yaml",
		".xml":  "application/xml",
		".csv":  "text/csv",
		".html": "text/html",
		".htm":  "text/html",
		".js":   "text/javascript",
		".css":  "text/css",
		".go":   "text/plain",
		".py":   "text/plain",
		".sh":   "text/plain",
		".mp4":  "video/mp4",
		".avi":  "video/avi",
		".mov":  "video/quicktime",
		".mkv":  "video/x-matroska",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
	}

	if t, exists := typeMap[ext]; exists {
		return t
	}

	return "application/octet-stream"
}

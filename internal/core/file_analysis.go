package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileMetadata contains extracted metadata from a file.
type FileMetadata struct {
	Name          string                 `json:"name"`
	Path          string                 `json:"path"`
	Size          int64                  `json:"size"`
	Extension     string                 `json:"extension"`
	MimeType      string                 `json:"mime_type"`
	ModTime       time.Time              `json:"mod_time"`
	IsDir         bool                   `json:"is_dir"`
	CanRead       bool                   `json:"can_read"`
	CanWrite      bool                   `json:"can_write"`
	ExtraMetadata map[string]interface{} `json:"extra_metadata,omitempty"`
}

// DirectoryStats contains statistics about a directory.
type DirectoryStats struct {
	Path         string                   `json:"path"`
	FileCount    int                      `json:"file_count"`
	DirCount     int                      `json:"dir_count"`
	TotalSize    int64                    `json:"total_size"`
	FilesByType  map[string]int           `json:"files_by_type"`
	LargestFiles []map[string]interface{} `json:"largest_files"`
	LastModified time.Time                `json:"last_modified"`
	AnalysisTime string                   `json:"analysis_time"`
}

// PatternMatch represents a file matching a pattern.
type PatternMatch struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// InitFileAnalysisTools registers file analysis tools on the executor.
func InitFileAnalysisTools(te *ToolExecutor) {
	// extract_metadata
	te.RegisterTool(
		ToolDefinition{
			Name:        "extract_metadata",
			Description: "Extract detailed metadata from a file (size, type, timestamps, permissions)",
			Category:    "file_analysis",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "File path to analyze",
					Required:    true,
				},
				{
					Name:        "include_extra",
					Type:        "bool",
					Description: "Include extra metadata (permissions, owner, etc). Optional, defaults to false.",
					Required:    false,
					Default:     false,
				},
			},
		},
		te.executeExtractMetadata,
	)

	// scan_directory_content
	te.RegisterTool(
		ToolDefinition{
			Name:        "scan_directory_content",
			Description: "Analyze directory structure, file count, size distribution, and file types",
			Category:    "file_analysis",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "Directory path to analyze",
					Required:    true,
				},
				{
					Name:        "max_depth",
					Type:        "int",
					Description: "Maximum recursion depth. Optional, defaults to 2.",
					Required:    false,
					Default:     2,
				},
				{
					Name:        "largest_n",
					Type:        "int",
					Description: "Return top N largest files. Optional, defaults to 10.",
					Required:    false,
					Default:     10,
				},
			},
		},
		te.executeScanDirectoryContent,
	)

	// get_file_patterns
	te.RegisterTool(
		ToolDefinition{
			Name:        "get_file_patterns",
			Description: "Find files matching a pattern (glob or regex)",
			Category:    "file_analysis",
			Parameters: []ToolParam{
				{
					Name:        "directory",
					Type:        "string",
					Description: "Directory to search in",
					Required:    true,
				},
				{
					Name:        "pattern",
					Type:        "string",
					Description: "Pattern to match (glob: *.jpg or regex: .*\\.jpe?g$)",
					Required:    true,
				},
				{
					Name:        "pattern_type",
					Type:        "string",
					Description: "Pattern type: 'glob' or 'regex'. Optional, defaults to 'glob'.",
					Required:    false,
					Default:     "glob",
				},
				{
					Name:        "recursive",
					Type:        "bool",
					Description: "Search recursively. Optional, defaults to true.",
					Required:    false,
					Default:     true,
				},
				{
					Name:        "max_results",
					Type:        "int",
					Description: "Maximum results to return. Optional, defaults to 1000.",
					Required:    false,
					Default:     1000,
				},
			},
		},
		te.executeGetFilePatterns,
	)

	// analyze_file_types
	te.RegisterTool(
		ToolDefinition{
			Name:        "analyze_file_types",
			Description: "Analyze and categorize files by type in a directory",
			Category:    "file_analysis",
			Parameters: []ToolParam{
				{
					Name:        "path",
					Type:        "string",
					Description: "Directory path to analyze",
					Required:    true,
				},
				{
					Name:        "depth",
					Type:        "int",
					Description: "Recursion depth. Optional, defaults to 1.",
					Required:    false,
					Default:     1,
				},
			},
		},
		te.executeAnalyzeFileTypes,
	)
}

// ============================================================================
// File Analysis Executors
// ============================================================================

func (te *ToolExecutor) executeExtractMetadata(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Success: false, Error: "path parameter is required"}
	}

	fullPath := resolvePath(te.cwd, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to stat file: %v", err)}
	}

	meta := FileMetadata{
		Name:      info.Name(),
		Path:      path,
		Size:      info.Size(),
		Extension: strings.ToLower(filepath.Ext(fullPath)),
		MimeType:  detectFileTypeByPath(fullPath),
		ModTime:   info.ModTime(),
		IsDir:     info.IsDir(),
		CanRead:   true,
		CanWrite:  (info.Mode() & 0o200) != 0,
	}

	// Check if readable
	if f, err := os.Open(fullPath); err == nil {
		f.Close()
		meta.CanRead = true
	} else {
		meta.CanRead = false
	}

	// Add extra metadata if requested
	if extra, ok := params["include_extra"].(bool); ok && extra {
		meta.ExtraMetadata = map[string]interface{}{
			"mode":        info.Mode().String(),
			"is_symlink":  (info.Mode() & os.ModeSymlink) != 0,
			"permissions": fmt.Sprintf("%o", info.Mode().Perm()),
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"metadata": meta,
		},
	}
}

func (te *ToolExecutor) executeScanDirectoryContent(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		path = "."
	}

	fullPath := resolvePath(te.cwd, path)

	maxDepth := 2
	if md, ok := params["max_depth"].(float64); ok {
		maxDepth = int(md)
	}

	largestN := 10
	if ln, ok := params["largest_n"].(float64); ok {
		largestN = int(ln)
	}

	start := time.Now()

	stats := &DirectoryStats{
		Path:         path,
		FilesByType:  make(map[string]int),
		LargestFiles: make([]map[string]interface{}, 0),
	}

	// Walk directory
	var files []struct {
		path string
		info os.FileInfo
	}

	filepath.Walk(fullPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Check depth
		depth := strings.Count(strings.TrimPrefix(p, fullPath), string(os.PathSeparator))
		if maxDepth > 0 && depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			stats.DirCount++
		} else {
			stats.FileCount++
			stats.TotalSize += info.Size()
			files = append(files, struct {
				path string
				info os.FileInfo
			}{p, info})

			ext := filepath.Ext(info.Name())
			if ext == "" {
				ext = "no-extension"
			}
			stats.FilesByType[ext]++
		}

		if info.ModTime().After(stats.LastModified) {
			stats.LastModified = info.ModTime()
		}

		return nil
	})

	// Get largest files
	if len(files) > 0 {
		// Simple sort to find largest files
		for i := 0; i < len(files) && i < largestN; i++ {
			largest := i
			for j := i + 1; j < len(files); j++ {
				if files[j].info.Size() > files[largest].info.Size() {
					largest = j
					files[i], files[j] = files[j], files[i]
				}
			}
			if i < len(files) {
				stats.LargestFiles = append(stats.LargestFiles, map[string]interface{}{
					"name": files[i].info.Name(),
					"path": files[i].path,
					"size": files[i].info.Size(),
				})
			}
		}
	}

	stats.AnalysisTime = time.Since(start).String()

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"stats": stats,
		},
	}
}

func (te *ToolExecutor) executeGetFilePatterns(ctx context.Context, params map[string]interface{}) ToolResult {
	directory, ok := params["directory"].(string)
	if !ok || directory == "" {
		directory = "."
	}

	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return ToolResult{Success: false, Error: "pattern parameter is required"}
	}

	fullDir := resolvePath(te.cwd, directory)

	patternType := "glob"
	if pt, ok := params["pattern_type"].(string); ok {
		patternType = strings.ToLower(pt)
	}

	recursive := true
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	maxResults := 1000
	if mr, ok := params["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	matches := make([]PatternMatch, 0)

	walkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !recursive && p != fullDir && filepath.Dir(p) != fullDir {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		matched := false

		if patternType == "regex" {
			if re, err := regexp.Compile(pattern); err == nil {
				matched = re.MatchString(info.Name())
			}
		} else {
			// Glob pattern
			if m, err := filepath.Match(pattern, info.Name()); err == nil {
				matched = m
			}
		}

		if matched && !info.IsDir() {
			matches = append(matches, PatternMatch{
				Path: filepath.Join(directory, strings.TrimPrefix(p, fullDir)),
				Name: info.Name(),
				Size: info.Size(),
			})

			if len(matches) >= maxResults {
				return filepath.SkipDir
			}
		}

		return nil
	}

	if recursive {
		filepath.Walk(fullDir, walkFunc)
	} else {
		entries, err := os.ReadDir(fullDir)
		if err != nil {
			return ToolResult{Success: false, Error: fmt.Sprintf("failed to read directory: %v", err)}
		}
		for _, entry := range entries {
			info, _ := entry.Info()
			if info != nil {
				walkFunc(filepath.Join(fullDir, entry.Name()), info, nil)
			}
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"pattern":   pattern,
			"count":     len(matches),
			"matches":   matches,
			"truncated": len(matches) >= maxResults,
		},
	}
}

func (te *ToolExecutor) executeAnalyzeFileTypes(ctx context.Context, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		path = "."
	}

	depth := 1
	if d, ok := params["depth"].(float64); ok {
		depth = int(d)
	}

	fullPath := resolvePath(te.cwd, path)

	typeStats := make(map[string]map[string]interface{})

	filepath.Walk(fullPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Check depth
		relPath := strings.TrimPrefix(p, fullPath)
		currentDepth := strings.Count(relPath, string(os.PathSeparator))
		if depth > 0 && currentDepth > depth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			mimeType := detectFileTypeByPath(p)
			category := extractMimeTypeCategory(mimeType)

			if _, exists := typeStats[category]; !exists {
				typeStats[category] = map[string]interface{}{
					"count":      0,
					"total_size": int64(0),
					"mime_types": make(map[string]int),
				}
			}

			stats := typeStats[category]
			stats["count"] = stats["count"].(int) + 1
			stats["total_size"] = stats["total_size"].(int64) + info.Size()

			mimeTypes := stats["mime_types"].(map[string]int)
			mimeTypes[mimeType]++
			stats["mime_types"] = mimeTypes
		}

		return nil
	})

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":             path,
			"file_type_stats":  typeStats,
			"categories_found": len(typeStats),
		},
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func detectFileTypeByPath(fullPath string) string {
	ext := strings.ToLower(filepath.Ext(fullPath))
	typeMap := map[string]string{
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".yaml": "application/yaml",
		".yml":  "application/yaml",
		".xml":  "application/xml",
		".csv":  "text/csv",
		".html": "text/html",
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

func extractMimeTypeCategory(mimeType string) string {
	parts := strings.Split(mimeType, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "application"
}

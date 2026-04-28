package core

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
)

// ToolsDiscoveryReport contains a comprehensive report of available tools.
type ToolsDiscoveryReport struct {
	AvailableCount int        `json:"available_count"`
	MissingCount   int        `json:"missing_count"`
	AvailableTools []ToolInfo `json:"available_tools"`
	MissingTools   []ToolInfo `json:"missing_tools"`
	SystemInfo     string     `json:"system_info"`
	GeneratedAt    string     `json:"generated_at"`
}

// DiscoverAllToolsWithConfig discovers all known tools and returns a comprehensive report using config.
func DiscoverAllToolsWithConfig(ctx context.Context, cfg Config) ToolsDiscoveryReport {
	var available []ToolInfo
	var missing []ToolInfo

	for toolName := range cfg.Tools.Registry {
		info := DetectToolWithVersion(ctx, toolName)
		if info.Available {
			available = append(available, info)
		} else {
			missing = append(missing, info)
		}
	}

	// Sort by name for consistency
	sortToolsByName(available)
	sortToolsByName(missing)

	return ToolsDiscoveryReport{
		AvailableCount: len(available),
		MissingCount:   len(missing),
		AvailableTools: available,
		MissingTools:   missing,
		SystemInfo:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// DiscoverAllTools discovers all known tools and returns a comprehensive report (deprecated, use DiscoverAllToolsWithConfig).
func DiscoverAllTools(ctx context.Context) ToolsDiscoveryReport {
	cfg, _ := LoadOrDefault()
	return DiscoverAllToolsWithConfig(ctx, cfg)
}

// GetRequiredToolsWithConfig returns only the required tools that are missing using config.
func GetRequiredToolsWithConfig(ctx context.Context, cfg Config) []ToolInfo {
	var missing []ToolInfo
	for toolName, check := range cfg.Tools.Registry {
		if check.Required {
			info := DetectToolWithVersion(ctx, toolName)
			if !info.Available {
				missing = append(missing, info)
			}
		}
	}
	sortToolsByName(missing)
	return missing
}

// GetRequiredTools returns only the required tools that are missing (deprecated, use GetRequiredToolsWithConfig).
func GetRequiredTools(ctx context.Context) []ToolInfo {
	cfg, _ := LoadOrDefault()
	return GetRequiredToolsWithConfig(ctx, cfg)
}

// CheckToolGroup checks availability of a group of tools by category/purpose using config.
func CheckToolGroupWithConfig(ctx context.Context, purpose string, cfg Config) map[string]ToolInfo {
	tools, exists := cfg.Tools.Groups[strings.ToLower(purpose)]
	if !exists {
		return nil
	}

	result := make(map[string]ToolInfo)
	for _, toolName := range tools {
		result[toolName] = DetectToolWithVersion(ctx, toolName)
	}
	return result
}

// CheckToolGroup checks availability of a group of tools by category/purpose (deprecated, use CheckToolGroupWithConfig).
func CheckToolGroup(ctx context.Context, purpose string) map[string]ToolInfo {
	cfg, _ := LoadOrDefault()
	return CheckToolGroupWithConfig(ctx, purpose, cfg)
}

// SuggestToolsForTaskWithConfig suggests which tools to install for a specific task using config.
func SuggestToolsForTaskWithConfig(ctx context.Context, task string, cfg Config) []string {
	tools, exists := cfg.Tools.TaskMap[strings.ToLower(task)]
	if !exists {
		return []string{}
	}

	var missing []string
	for _, toolName := range tools {
		info := DetectToolWithVersion(ctx, toolName)
		if !info.Available {
			missing = append(missing, toolName)
		}
	}
	return missing
}

// SuggestToolsForTask suggests which tools to install for a specific task (deprecated, use SuggestToolsForTaskWithConfig).
func SuggestToolsForTask(ctx context.Context, task string) []string {
	cfg, _ := LoadOrDefault()
	return SuggestToolsForTaskWithConfig(ctx, task, cfg)
}

// FormatToolsReport formats a tools discovery report as a readable string.
func FormatToolsReport(report ToolsDiscoveryReport) string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║              CLI Tools Discovery Report                   ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("System: %s\n", report.SystemInfo))
	sb.WriteString(fmt.Sprintf("Available: %d | Missing: %d\n\n", report.AvailableCount, report.MissingCount))

	if len(report.AvailableTools) > 0 {
		sb.WriteString("✓ Available Tools:\n")
		for _, tool := range report.AvailableTools {
			sb.WriteString(fmt.Sprintf("  • %s", tool.Name))
			if tool.Version != "" {
				sb.WriteString(fmt.Sprintf(" v%s", tool.Version))
			}
			sb.WriteString(fmt.Sprintf(" (%s)\n", tool.Path))
		}
		sb.WriteString("\n")
	}

	if len(report.MissingTools) > 0 {
		sb.WriteString("✗ Missing Tools:\n")
		for _, tool := range report.MissingTools {
			sb.WriteString(fmt.Sprintf("  • %s - %s\n", tool.Name, tool.Error))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FilterToolsByCategoryWithConfig filters tools by category using config groups.
func FilterToolsByCategoryWithConfig(ctx context.Context, category string, cfg Config) []ToolInfo {
	// First try to find in config groups
	if tools, exists := cfg.Tools.Groups[strings.ToLower(category)]; exists {
		var result []ToolInfo
		for _, toolName := range tools {
			info := DetectToolWithVersion(ctx, toolName)
			result = append(result, info)
		}
		return result
	}

	// Fallback to predefined categories for backward compatibility
	categories := map[string][]string{
		"image":           {"exiftool", "magick"},
		"video":           {"ffmpeg"},
		"metadata":        {"exiftool"},
		"graphics":        {"magick"},
		"version_control": {"git"},
		"runtime":         {"node", "python", "python3", "go"},
	}

	toolNames, exists := categories[strings.ToLower(category)]
	if !exists {
		return []ToolInfo{}
	}

	var result []ToolInfo
	for _, toolName := range toolNames {
		info := DetectToolWithVersion(ctx, toolName)
		result = append(result, info)
	}
	return result
}

// FilterToolsByCategory filters tools by their typical use category (deprecated, use FilterToolsByCategoryWithConfig).
func FilterToolsByCategory(ctx context.Context, category string) []ToolInfo {
	cfg, _ := LoadOrDefault()
	return FilterToolsByCategoryWithConfig(ctx, category, cfg)
}

// GetInstallationScriptWithConfig generates a shell script to install missing tools using config.
func GetInstallationScriptWithConfig(ctx context.Context, cfg Config) string {
	missing := GetRequiredToolsWithConfig(ctx, cfg)
	if len(missing) == 0 {
		return "# All required tools are installed!"
	}

	osType := runtime.GOOS
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("#!/bin/bash\n# Install missing tools for %s\n\n", osType))

	for _, tool := range missing {
		hint := SuggestInstallation(tool.Name)
		// Extract just the install command from the hint
		lines := strings.Split(hint, "\n")
		if len(lines) > 1 {
			sb.WriteString(fmt.Sprintf("echo \"Installing %s...\"\n", tool.Name))
			sb.WriteString(strings.TrimSpace(lines[len(lines)-1]) + "\n")
		}
	}

	return sb.String()
}

// GetInstallationScript generates a shell script to install missing tools (deprecated, use GetInstallationScriptWithConfig).
func GetInstallationScript(ctx context.Context) string {
	cfg, _ := LoadOrDefault()
	return GetInstallationScriptWithConfig(ctx, cfg)
}

// sortToolsByName sorts tool info by name.
func sortToolsByName(tools []ToolInfo) {
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
}

// ToolsForAIAgentWithConfig returns tool information formatted for the AI agent context using config.
func ToolsForAIAgentWithConfig(ctx context.Context, cfg Config) map[string]interface{} {
	available := make([]string, 0)
	missing := make([]string, 0)

	for toolName := range cfg.Tools.Registry {
		info := DetectToolWithVersion(ctx, toolName)
		if info.Available {
			available = append(available, toolName)
		} else {
			missing = append(missing, toolName)
		}
	}

	sort.Strings(available)
	sort.Strings(missing)

	return map[string]interface{}{
		"available": available,
		"missing":   missing,
		"system":    runtime.GOOS,
		"groups":    cfg.Tools.Groups,
		"task_map":  cfg.Tools.TaskMap,
	}
}

// ToolsForAIAgent returns tool information formatted for the AI agent context (deprecated, use ToolsForAIAgentWithConfig).
func ToolsForAIAgent(ctx context.Context) map[string]interface{} {
	cfg, _ := LoadOrDefault()
	return ToolsForAIAgentWithConfig(ctx, cfg)
}

// ToolCapabilities provides information about what each tool can do.
func ToolCapabilities() map[string][]string {
	return map[string][]string{
		"exiftool": {
			"Extract metadata from images (EXIF, IPTC, XMP)",
			"Read/write file metadata",
			"Rename files based on metadata",
			"Organize by date/camera/location",
		},
		"ffmpeg": {
			"Transcode video and audio",
			"Extract metadata from media files",
			"Get video/audio duration and properties",
			"Create thumbnails from videos",
			"Batch process media files",
		},
		"magick": {
			"Image format conversion",
			"Image resizing and scaling",
			"Extract image metadata",
			"Batch image processing",
			"Image analysis (dimensions, colors)",
		},
		"git": {
			"Version control operations",
			"Repository management",
		},
		"node": {
			"JavaScript runtime",
			"Execute Node.js scripts",
		},
		"python": {
			"Python script execution",
			"Data processing",
		},
		"python3": {
			"Python 3 script execution",
			"Modern Python support",
		},
		"npm": {
			"Node package management",
			"Dependency installation",
		},
		"go": {
			"Go compilation and execution",
			"Build Go projects",
		},
	}
}

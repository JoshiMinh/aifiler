package core

import (
	"context"
	"fmt"
	"strings"
)

// SafetyValidator validates tool parameters before execution.
type SafetyValidator struct {
	allowedPaths      []string
	blockedExtensions []string
	maxFileSize       int64
	maxCommandLength  int
}

// NewSafetyValidator creates a new safety validator with defaults.
func NewSafetyValidator() SafetyValidator {
	return SafetyValidator{
		allowedPaths:      []string{},
		blockedExtensions: []string{".exe", ".bat", ".cmd", ".com", ".sh"},
		maxFileSize:       1073741824, // 1GB
		maxCommandLength:  10000,
	}
}

// ValidatePath checks if a path is safe to access.
func (sv SafetyValidator) ValidatePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	if strings.HasPrefix(strings.ToLower(path), "c:\\windows") {
		return fmt.Errorf("access to system directories not allowed: %s", path)
	}

	if strings.HasPrefix(strings.ToLower(path), "/system") {
		return fmt.Errorf("access to system directories not allowed: %s", path)
	}

	return nil
}

// ValidateCommand checks if a command is safe to execute.
func (sv SafetyValidator) ValidateCommand(cmd string) error {
	if len(cmd) > sv.maxCommandLength {
		return fmt.Errorf("command too long: %d > %d", len(cmd), sv.maxCommandLength)
	}

	dangerous := []string{
		"rm -rf /",
		"del /s /q",
		"format c:",
		":(){:|:&};:",
	}

	for _, pattern := range dangerous {
		if strings.Contains(strings.ToLower(cmd), strings.ToLower(pattern)) {
			return fmt.Errorf("dangerous command pattern detected")
		}
	}

	return nil
}

// ValidateFileSize checks if file size is acceptable.
func (sv SafetyValidator) ValidateFileSize(size int64) error {
	if size > sv.maxFileSize {
		return fmt.Errorf("file too large: %d > %d", size, sv.maxFileSize)
	}
	return nil
}

// ToolDefinitionRegistry provides a centralized registry of all available tool definitions.
type ToolDefinitionRegistry struct {
	tools map[string]ToolDefinition
}

// NewToolDefinitionRegistry creates a new tool registry.
func NewToolDefinitionRegistry(executor *ToolExecutor) *ToolDefinitionRegistry {
	registry := &ToolDefinitionRegistry{
		tools: make(map[string]ToolDefinition),
	}

	// Register all tools from executor
	for _, def := range executor.GetToolDefinitions() {
		registry.tools[def.Name] = def
	}

	return registry
}

// GetTool retrieves a tool definition by name.
func (tr *ToolDefinitionRegistry) GetTool(name string) (ToolDefinition, bool) {
	tool, exists := tr.tools[name]
	return tool, exists
}

// GetToolsByCategory returns all tools in a category.
func (tr *ToolDefinitionRegistry) GetToolsByCategory(category string) []ToolDefinition {
	var tools []ToolDefinition
	for _, tool := range tr.tools {
		if tool.Category == category {
			tools = append(tools, tool)
		}
	}
	return tools
}

// GetCategoriesWithTools returns all categories and their tool counts.
func (tr *ToolDefinitionRegistry) GetCategoriesWithTools() map[string]int {
	categories := make(map[string]int)
	for _, tool := range tr.tools {
		categories[tool.Category]++
	}
	return categories
}

// GetAllTools returns all tool definitions.
func (tr *ToolDefinitionRegistry) GetAllTools() []ToolDefinition {
	var tools []ToolDefinition
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolDescriptionForAI returns tool definitions formatted for AI prompts.
func (tr *ToolDefinitionRegistry) GetToolDescriptionForAI() string {
	categories := tr.GetCategoriesWithTools()
	var sb strings.Builder

	sb.WriteString("# Available Tools\n\n")

	for category, count := range categories {
		sb.WriteString(fmt.Sprintf("## %s (%d tools)\n\n", formatCategory(category), count))

		for _, tool := range tr.GetToolsByCategory(category) {
			sb.WriteString(fmt.Sprintf("### %s\n", tool.Name))
			sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", tool.Description))

			if len(tool.Parameters) > 0 {
				sb.WriteString("**Parameters:**\n")
				for _, param := range tool.Parameters {
					required := "optional"
					if param.Required {
						required = "required"
					}
					sb.WriteString(fmt.Sprintf("- `%s` (%s, %s): %s", param.Name, param.Type, required, param.Description))
					if param.Default != nil {
						sb.WriteString(fmt.Sprintf(" (default: %v)", param.Default))
					}
					sb.WriteString("\n")
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// GetToolsByUsagePattern returns tools suitable for specific tasks.
func (tr *ToolDefinitionRegistry) GetToolsByUsagePattern(pattern string) []ToolDefinition {
	taskToTools := map[string][]string{
		"file_reading":    {"read_file", "list_directory", "get_file_metadata", "get_file_type"},
		"file_writing":    {"write_file", "create_directory", "copy_file", "rename_file", "delete_file"},
		"file_analysis":   {"extract_metadata", "scan_directory_content", "get_file_patterns", "analyze_file_types"},
		"system_commands": {"run_command", "run_command_safe", "check_command_available", "get_command_help"},
		"tool_discovery":  {"check_available_tools", "check_tool_availability", "get_tool_info", "get_installation_hint"},
	}

	toolNames, exists := taskToTools[strings.ToLower(pattern)]
	if !exists {
		return []ToolDefinition{}
	}

	var tools []ToolDefinition
	for _, name := range toolNames {
		if tool, exists := tr.GetTool(name); exists {
			tools = append(tools, tool)
		}
	}
	return tools
}

// ValidateToolCall validates a tool call before execution.
func (tr *ToolDefinitionRegistry) ValidateToolCall(toolName string, params map[string]interface{}) error {
	tool, exists := tr.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool '%s' not found", toolName)
	}

	// Check required parameters
	for _, param := range tool.Parameters {
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				return fmt.Errorf("required parameter missing: %s", param.Name)
			}
		}
	}

	// Additional safety checks
	validator := NewSafetyValidator()

	// Check path parameters
	if path, exists := params["path"].(string); exists {
		if err := validator.ValidatePath(path); err != nil {
			return err
		}
	}

	// Check command parameters
	if cmd, exists := params["command"].(string); exists {
		if err := validator.ValidateCommand(cmd); err != nil {
			return err
		}
	}

	return nil
}

// formatCategory formats a category name for display.
func formatCategory(category string) string {
	words := strings.Split(category, "_")
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

// ToolUsageExample provides usage examples for a tool.
type ToolUsageExample struct {
	Description    string                 `json:"description"`
	Parameters     map[string]interface{} `json:"parameters"`
	ExpectedResult string                 `json:"expected_result"`
}

// GetToolExamples returns usage examples for tools.
func GetToolExamples() map[string][]ToolUsageExample {
	return map[string][]ToolUsageExample{
		"read_file": {
			{
				Description: "Read entire file",
				Parameters: map[string]interface{}{
					"path": "config.yaml",
				},
				ExpectedResult: "File content returned",
			},
			{
				Description: "Read lines 10-20",
				Parameters: map[string]interface{}{
					"path":       "main.go",
					"start_line": 10,
					"end_line":   20,
				},
				ExpectedResult: "Lines 10-20 returned",
			},
		},
		"run_command": {
			{
				Description: "Check tool version",
				Parameters: map[string]interface{}{
					"command": "ffmpeg -version",
					"timeout": 10,
				},
				ExpectedResult: "Version information printed",
			},
		},
		"scan_directory_content": {
			{
				Description: "Analyze current directory",
				Parameters: map[string]interface{}{
					"path":      ".",
					"max_depth": 2,
					"largest_n": 10,
				},
				ExpectedResult: "Directory statistics and largest files",
			},
		},
	}
}

// ExecuteToolWithValidation executes a tool with safety validation.
func (te *ToolExecutor) ExecuteToolWithValidation(ctx context.Context, toolName string, params map[string]interface{}) ToolResult {
	registry := NewToolDefinitionRegistry(te)

	// Validate before execution
	if err := registry.ValidateToolCall(toolName, params); err != nil {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("validation error: %v", err),
		}
	}

	// Execute tool
	return te.ExecuteTool(ctx, toolName, params)
}

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ToolParam defines a parameter for a tool.
type ToolParam struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, bool, int, array
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// ToolDefinition describes an available tool that the AI agent can use.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"` // cli_tools, file_ops, command_exec, etc.
	Parameters  []ToolParam `json:"parameters"`
}

// ToolResult is the result of executing a tool.
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ToolExecutorFunc is the signature for tool execution functions.
type ToolExecutorFunc func(ctx context.Context, params map[string]interface{}) ToolResult

// ToolExecutor manages tool definitions and execution.
type ToolExecutor struct {
	definitions map[string]ToolDefinition
	executors   map[string]ToolExecutorFunc
	cwd         string
	config      Config
}

// NewToolExecutor creates a new tool executor for the AI agent.
func NewToolExecutor(cwd string) *ToolExecutor {
	cfg, _ := LoadOrDefault()
	return NewToolExecutorWithConfig(cwd, cfg)
}

// NewToolExecutorWithConfig creates a new tool executor with custom configuration.
func NewToolExecutorWithConfig(cwd string, cfg Config) *ToolExecutor {
	te := &ToolExecutor{
		definitions: make(map[string]ToolDefinition),
		executors:   make(map[string]ToolExecutorFunc),
		cwd:         cwd,
		config:      cfg,
	}
	te.registerCLIToolsDiscovery()
	InitFileOperationsTools(te)
	InitCommandExecutionTools(te)
	InitFileAnalysisTools(te)
	return te
}

// RegisterTool registers a tool with its definition and executor.
func (te *ToolExecutor) RegisterTool(def ToolDefinition, executor ToolExecutorFunc) {
	te.definitions[def.Name] = def
	te.executors[def.Name] = executor
}

// ExecuteTool executes a registered tool with the given parameters.
func (te *ToolExecutor) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) ToolResult {
	executor, exists := te.executors[toolName]
	if !exists {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool '%s' not found", toolName),
		}
	}
	return executor(ctx, params)
}

// GetToolDefinitions returns all registered tool definitions.
func (te *ToolExecutor) GetToolDefinitions() []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(te.definitions))
	for _, def := range te.definitions {
		defs = append(defs, def)
	}
	return defs
}

// GetToolDefinitionsByCategory returns tool definitions for a specific category.
func (te *ToolExecutor) GetToolDefinitionsByCategory(category string) []ToolDefinition {
	var defs []ToolDefinition
	for _, def := range te.definitions {
		if def.Category == category {
			defs = append(defs, def)
		}
	}
	return defs
}

// ToJSON converts a ToolResult to JSON string.
func (tr ToolResult) ToJSON() string {
	data, _ := json.MarshalIndent(tr, "", "  ")
	return string(data)
}

// ToJSONString converts a ToolResult to formatted JSON string.
func (tr ToolResult) ToJSONString() string {
	return tr.ToJSON()
}

// ============================================================================
// CLI Tools Discovery Tool Implementations
// ============================================================================

// registerCLIToolsDiscovery registers CLI tools discovery tools.
func (te *ToolExecutor) registerCLIToolsDiscovery() {
	// check_available_tools
	te.RegisterTool(
		ToolDefinition{
			Name:        "check_available_tools",
			Description: "List all available CLI tools detected on the system with their versions and paths",
			Category:    "cli_tools",
			Parameters: []ToolParam{
				{
					Name:        "tools",
					Type:        "array",
					Description: "Optional list of specific tool names to check. If empty, checks all known tools.",
					Required:    false,
					Default:     []string{},
				},
			},
		},
		te.executeCheckAvailableTools,
	)

	// check_tool_availability
	te.RegisterTool(
		ToolDefinition{
			Name:        "check_tool_availability",
			Description: "Check if a specific CLI tool is available on the system",
			Category:    "cli_tools",
			Parameters: []ToolParam{
				{
					Name:        "tool_name",
					Type:        "string",
					Description: "Name of the tool to check (e.g., 'ffmpeg', 'exiftool', 'git')",
					Required:    true,
				},
			},
		},
		te.executeCheckToolAvailability,
	)

	// get_tool_info
	te.RegisterTool(
		ToolDefinition{
			Name:        "get_tool_info",
			Description: "Get detailed information about a CLI tool including version, path, and capabilities",
			Category:    "cli_tools",
			Parameters: []ToolParam{
				{
					Name:        "tool_name",
					Type:        "string",
					Description: "Name of the tool to get info for (e.g., 'ffmpeg', 'exiftool')",
					Required:    true,
				},
			},
		},
		te.executeGetToolInfo,
	)

	// get_installation_hint
	te.RegisterTool(
		ToolDefinition{
			Name:        "get_installation_hint",
			Description: "Get OS-specific installation command for a missing CLI tool",
			Category:    "cli_tools",
			Parameters: []ToolParam{
				{
					Name:        "tool_name",
					Type:        "string",
					Description: "Name of the tool to get installation command for",
					Required:    true,
				},
			},
		},
		te.executeGetInstallationHint,
	)

	// run_sanity_check
	te.RegisterTool(
		ToolDefinition{
			Name:        "run_sanity_check",
			Description: "Run a sanity check on a CLI tool to verify it works correctly",
			Category:    "cli_tools",
			Parameters: []ToolParam{
				{
					Name:        "tool_name",
					Type:        "string",
					Description: "Name of the tool to run sanity check on",
					Required:    true,
				},
			},
		},
		te.executeRunSanityCheck,
	)
}

// executeCheckAvailableTools lists all available tools.
func (te *ToolExecutor) executeCheckAvailableTools(ctx context.Context, params map[string]interface{}) ToolResult {
	var toolNames []string
	if tools, exists := params["tools"]; exists {
		if toolsArr, ok := tools.([]interface{}); ok {
			for _, t := range toolsArr {
				if toolStr, ok := t.(string); ok {
					toolNames = append(toolNames, toolStr)
				}
			}
		}
	}

	// If no specific tools provided, use all known tools
	if len(toolNames) == 0 {
		for toolName := range ToolRegistry {
			toolNames = append(toolNames, toolName)
		}
	}

	availableTools := make([]ToolInfo, 0)
	for _, name := range toolNames {
		info := DetectToolWithVersion(ctx, name)
		if info.Available {
			availableTools = append(availableTools, info)
		}
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"count":   len(availableTools),
			"tools":   availableTools,
			"message": fmt.Sprintf("Found %d available tools", len(availableTools)),
		},
	}
}

// executeCheckToolAvailability checks if a specific tool is available.
func (te *ToolExecutor) executeCheckToolAvailability(ctx context.Context, params map[string]interface{}) ToolResult {
	toolName, ok := params["tool_name"].(string)
	if !ok || toolName == "" {
		return ToolResult{
			Success: false,
			Error:   "tool_name parameter is required and must be a string",
		}
	}

	info := DetectToolWithVersion(ctx, toolName)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"tool":      info.Name,
			"available": info.Available,
			"version":   info.Version,
			"path":      info.Path,
			"error":     info.Error,
		},
	}
}

// executeGetToolInfo gets detailed tool information.
func (te *ToolExecutor) executeGetToolInfo(ctx context.Context, params map[string]interface{}) ToolResult {
	toolName, ok := params["tool_name"].(string)
	if !ok || toolName == "" {
		return ToolResult{
			Success: false,
			Error:   "tool_name parameter is required and must be a string",
		}
	}

	info := DetectToolWithVersion(ctx, toolName)

	if !info.Available {
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool '%s' not found: %s", toolName, info.Error),
		}
	}

	// Run sanity check if tool is available
	passed, output, checkErr := RunSanityCheck(ctx, toolName)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"name":              info.Name,
			"available":         info.Available,
			"version":           info.Version,
			"path":              info.Path,
			"sanity_check_pass": passed,
			"sanity_check_out":  strings.TrimSpace(output),
			"sanity_check_err":  getErrorString(checkErr),
		},
	}
}

// executeGetInstallationHint gets OS-specific installation hint.
func (te *ToolExecutor) executeGetInstallationHint(ctx context.Context, params map[string]interface{}) ToolResult {
	toolName, ok := params["tool_name"].(string)
	if !ok || toolName == "" {
		return ToolResult{
			Success: false,
			Error:   "tool_name parameter is required and must be a string",
		}
	}

	hint := SuggestInstallation(toolName)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"tool":         toolName,
			"hint":         hint,
			"installation": hint,
		},
	}
}

// executeRunSanityCheck runs a sanity check on a tool.
func (te *ToolExecutor) executeRunSanityCheck(ctx context.Context, params map[string]interface{}) ToolResult {
	toolName, ok := params["tool_name"].(string)
	if !ok || toolName == "" {
		return ToolResult{
			Success: false,
			Error:   "tool_name parameter is required and must be a string",
		}
	}

	passed, output, err := RunSanityCheck(ctx, toolName)

	return ToolResult{
		Success: !passed, // Success if check passed
		Data: map[string]interface{}{
			"tool":   toolName,
			"passed": passed,
			"output": strings.TrimSpace(output),
			"error":  getErrorString(err),
		},
	}
}

// ============================================================================
// File System Operations Tool Implementations
// ============================================================================

// registerFileSystemOps is no longer used - tools are registered in InitFileOperationsTools

// ============================================================================
// Command Execution Tool Implementations
// ============================================================================

// registerCommandExecution is no longer used - tools are registered in InitCommandExecutionTools

// ============================================================================
// File Analysis Tool Implementations
// ============================================================================

// registerFileAnalysis is no longer used - tools are registered in InitFileAnalysisTools

// ============================================================================
// Helper Functions
// ============================================================================

// getErrorString safely converts an error to a string.
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

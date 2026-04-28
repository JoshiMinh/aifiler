package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"aifiler/internal/core"
)

// ToolDefinitionForAI defines how tools are presented to AI models.
type ToolDefinitionForAI struct {
	Type     string          `json:"type"`
	Function ToolFunctionDef `json:"function"`
}

// ToolFunctionDef defines a function for AI models.
type ToolFunctionDef struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Parameters  ToolParametersSchema `json:"parameters"`
}

// ToolParametersSchema defines parameters schema for AI models.
type ToolParametersSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropertyDef `json:"properties"`
	Required   []string               `json:"required"`
}

// PropertyDef defines a property in parameters schema.
type PropertyDef struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// ConvertToolsForAI converts internal tool definitions to AI provider format.
func ConvertToolsForAI(toolDefs []core.ToolDefinition) []ToolDefinitionForAI {
	var aiTools []ToolDefinitionForAI

	for _, toolDef := range toolDefs {
		aiTool := ToolDefinitionForAI{
			Type: "function",
			Function: ToolFunctionDef{
				Name:        toolDef.Name,
				Description: toolDef.Description,
				Parameters: ToolParametersSchema{
					Type:       "object",
					Properties: make(map[string]PropertyDef),
					Required:   []string{},
				},
			},
		}

		for _, param := range toolDef.Parameters {
			prop := PropertyDef{
				Type:        param.Type,
				Description: param.Description,
				Default:     param.Default,
			}

			aiTool.Function.Parameters.Properties[param.Name] = prop

			if param.Required {
				aiTool.Function.Parameters.Required = append(aiTool.Function.Parameters.Required, param.Name)
			}
		}

		aiTools = append(aiTools, aiTool)
	}

	return aiTools
}

// ConvertToolsForAIJSON returns tools as JSON string for AI prompts.
func ConvertToolsForAIJSON(toolDefs []core.ToolDefinition) string {
	aiTools := ConvertToolsForAI(toolDefs)
	data, _ := json.MarshalIndent(aiTools, "", "  ")
	return string(data)
}

// GetAISystemPromptWithTools generates a system prompt that includes tool definitions.
func GetAISystemPromptWithTools(executor *core.ToolExecutor, basePrompt string) string {
	registry := core.NewToolDefinitionRegistry(executor)

	categories := registry.GetCategoriesWithTools()

	var sb strings.Builder
	sb.WriteString(basePrompt)
	sb.WriteString("\n\n")
	sb.WriteString("## Available Tools\n\n")
	sb.WriteString("You have access to the following tools to help organize and manage files:\n\n")

	for category := range categories {
		tools := registry.GetToolsByCategory(category)
		categoryName := formatCategoryName(category)
		sb.WriteString(fmt.Sprintf("### %s\n", categoryName))

		for _, tool := range tools {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", tool.Name, tool.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Tool Usage Instructions\n\n")
	sb.WriteString("1. Analyze the user's request to determine which tools are needed\n")
	sb.WriteString("2. Use the tool parameters schema to construct valid tool calls\n")
	sb.WriteString("3. All file paths should be relative to the working directory\n")
	sb.WriteString("4. Tool calls should be made with accurate parameter names and types\n")
	sb.WriteString("5. Always validate paths and commands before execution\n")

	return sb.String()
}

// ToolExecutionRequest represents a request to execute a tool.
type ToolExecutionRequest struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolExecutionResponse represents the result of tool execution.
type ToolExecutionResponse struct {
	ToolName string      `json:"tool_name"`
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// ExecuteToolFromAIRequest executes a tool based on AI request.
func ExecuteToolFromAIRequest(ctx context.Context, executor *core.ToolExecutor, req ToolExecutionRequest) ToolExecutionResponse {
	result := executor.ExecuteToolWithValidation(ctx, req.ToolName, req.Parameters)

	return ToolExecutionResponse{
		ToolName: req.ToolName,
		Success:  result.Success,
		Data:     result.Data,
		Error:    result.Error,
	}
}

// ParseToolCallFromAI attempts to parse tool call from AI response.
func ParseToolCallFromAI(content string) (ToolExecutionRequest, error) {
	// Try to find JSON in the content
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return ToolExecutionRequest{}, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := content[startIdx : endIdx+1]

	var req ToolExecutionRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return ToolExecutionRequest{}, fmt.Errorf("failed to parse tool call: %v", err)
	}

	return req, nil
}

// formatCategoryName formats category name for display.
func formatCategoryName(category string) string {
	words := strings.Split(category, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// ToolContextBuilder helps build AI context with tools.
type ToolContextBuilder struct {
	executor *core.ToolExecutor
	registry *core.ToolDefinitionRegistry
}

// NewToolContextBuilder creates a new tool context builder.
func NewToolContextBuilder(executor *core.ToolExecutor) *ToolContextBuilder {
	return &ToolContextBuilder{
		executor: executor,
		registry: core.NewToolDefinitionRegistry(executor),
	}
}

// BuildToolContext builds complete tool context for AI.
func (tcb *ToolContextBuilder) BuildToolContext() map[string]interface{} {
	categories := tcb.registry.GetCategoriesWithTools()
	toolsByCategory := make(map[string][]map[string]interface{})

	for category := range categories {
		tools := tcb.registry.GetToolsByCategory(category)
		var toolsList []map[string]interface{}

		for _, tool := range tools {
			toolsList = append(toolsList, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"params":      len(tool.Parameters),
			})
		}

		toolsByCategory[category] = toolsList
	}

	return map[string]interface{}{
		"total_tools":       len(tcb.registry.GetAllTools()),
		"categories":        categories,
		"tools_by_category": toolsByCategory,
		"usage_patterns": []string{
			"file_reading",
			"file_writing",
			"file_analysis",
			"system_commands",
			"tool_discovery",
		},
	}
}

// GetToolsForPattern returns tools suitable for a pattern.
func (tcb *ToolContextBuilder) GetToolsForPattern(pattern string) []core.ToolDefinition {
	return tcb.registry.GetToolsByUsagePattern(pattern)
}

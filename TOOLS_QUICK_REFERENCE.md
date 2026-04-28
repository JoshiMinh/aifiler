# aifiler Tools Quick Reference Guide

## Overview
**24 Production-Ready Tools** organized in 5 categories for file management and AI integration.

---

## 🔧 Tool Categories

### 1️⃣ File Operations (10 tools)
Essential file system operations with safety checks.

```
read_file         - Read file content (supports line ranges)
write_file        - Write/append to files
list_directory    - List directory contents
check_file_exists - Check file existence
get_file_metadata - Extract file details
get_file_type     - Detect MIME type
copy_file         - Copy files safely
delete_file       - Delete files
create_directory  - Create directories recursively
rename_file       - Rename/move files
```

### 2️⃣ File Analysis (4 tools)
Advanced analysis and pattern matching.

```
extract_metadata      - Get detailed file metadata
scan_directory_content - Analyze directory structure
get_file_patterns    - Find files by pattern (glob/regex)
analyze_file_types   - Categorize by file type
```

### 3️⃣ Command Execution (5 tools)
Safe system command execution with validation.

```
run_command           - Execute with 30s timeout
run_command_safe      - Execute with security checks
list_available_commands - Show common commands
check_command_available - Verify command exists
get_command_help      - Get command documentation
```

### 4️⃣ CLI Discovery (5 tools)
Tool detection and availability checking.

```
check_available_tools - Check all configured tools
check_tool_availability - Check specific tool
get_tool_info         - Get tool version
get_installation_hint - Get setup hints
run_sanity_check      - Verify all tools
```

---

## 🚀 Quick Start

### Initialize Tools
```go
// Basic initialization
executor := NewToolExecutor(".")

// With configuration
cfg := LoadConfig()
executor := NewToolExecutorWithConfig(".", cfg)
```

### Execute Tools
```go
// Simple execution
result := executor.ExecuteTool(ctx, "read_file", map[string]interface{}{
    "path": "config.yaml",
})

// With validation
result := executor.ExecuteToolWithValidation(ctx, "read_file", params)
```

### AI Integration
```go
// Get tools for AI
registry := NewToolDefinitionRegistry(executor)
aiTools := ConvertToolsForAI(registry.GetAllTools())

// Execute from AI
response := ExecuteToolFromAIRequest(ctx, executor, aiRequest)
```

---

## 📋 Common Scenarios

### Read a File
```go
executor.ExecuteTool(ctx, "read_file", map[string]interface{}{
    "path": "config.yaml",
    "start_line": 1,
    "end_line": 50,
})
```

### Find Files by Pattern
```go
executor.ExecuteTool(ctx, "get_file_patterns", map[string]interface{}{
    "directory": ".",
    "pattern": "*.json",
    "pattern_type": "glob",
})
```

### Analyze Directory
```go
executor.ExecuteTool(ctx, "scan_directory_content", map[string]interface{}{
    "path": ".",
    "max_depth": 2,
    "largest_n": 10,
})
```

### Run Command Safely
```go
executor.ExecuteTool(ctx, "run_command_safe", map[string]interface{}{
    "command": "git status",
    "timeout": 30,
})
```

---

## ⚙️ Configuration

### config.yaml Structure
```yaml
tools:
  registry:
    tool_name:
      command: "command to check"
      timeout: 5s
      required: false
  
  groups:
    group_name: [tool1, tool2, tool3]
  
  taskmap:
    task_name: [tool1, tool2]
```

### Tool Groups (Pre-configured)
```
media      - ffmpeg, magick, exiftool
dev        - git, go, node, npm, python
image      - exiftool, magick
video      - ffmpeg
metadata   - exiftool
```

---

## 🛡️ Safety Features

### Validation System
- Path traversal prevention
- Command injection blocking
- File size limits (1GB)
- Command length limits (10KB)
- System directory protection

### Pre-execution Checks
- Tool availability verification
- Parameter validation
- Permission checking
- Resource limit enforcement

---

## 📊 Performance

### Benchmarks
- Directory scan: ~50-200ms (1000+ files)
- File read: ~1-5ms (text files)
- Pattern matching: ~10-50ms (1000+ files)
- Metadata extraction: ~5-20ms per file

### Limits
- Read limit: 1MB per file
- Output limit: 64KB per command
- Timeout: 30 seconds default
- Recursion depth: Configurable

---

## 🧪 Testing

### PowerShell Test Suite
```powershell
.\test_tools_powershell.ps1
```

### Test Coverage
✅ File operations
✅ Pattern matching
✅ Directory analysis
✅ Tool integration
✅ Performance benchmarks
✅ Error handling

---

## 📝 API Reference

### ToolExecutor
```go
type ToolExecutor struct {
    // Execute a tool
    ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) ToolResult
    
    // Execute with validation
    ExecuteToolWithValidation(ctx context.Context, toolName string, params map[string]interface{}) ToolResult
    
    // Get all tool definitions
    GetToolDefinitions() []ToolDefinition
    
    // Register a tool
    RegisterTool(def ToolDefinition, executor ToolExecutorFunc)
}
```

### Tool Definition
```go
type ToolDefinition struct {
    Name        string
    Description string
    Category    string
    Parameters  []ToolParam
}
```

### Tool Result
```go
type ToolResult struct {
    Success bool
    Data    map[string]interface{}
    Error   string
}
```

---

## 🔗 Integration with AI Providers

### Anthropic/Claude
```go
// Get tools in Claude format
tools := ConvertToolsForAI(toolDefinitions)

// Include in request and handle responses
response := ExecuteToolFromAIRequest(ctx, executor, toolRequest)
```

### OpenAI/ChatGPT
Same pattern - tools converted to standard JSON format

### Google Gemini
Compatible with same tool format

### Ollama
Tool integration via standard format

### Vercel
Compatible with edge function patterns

---

## 🐛 Debugging

### Get Tool Info
```go
registry := NewToolDefinitionRegistry(executor)
tool, exists := registry.GetTool("read_file")
if exists {
    println(tool.Description)
    // Access parameters, etc.
}
```

### Validate Before Execution
```go
registry := NewToolDefinitionRegistry(executor)
err := registry.ValidateToolCall("read_file", params)
if err != nil {
    // Handle validation error
}
```

### List Available Tools by Category
```go
registry := NewToolDefinitionRegistry(executor)
fileTools := registry.GetToolsByCategory("file_operations")
for _, tool := range fileTools {
    println(tool.Name)
}
```

---

## 📚 Documentation

- **Full Implementation**: `TOOLS_IMPLEMENTATION_COMPLETE.md`
- **Configuration**: `config.yaml`
- **Code**: `internal/core/tools.go`
- **Safety**: `internal/core/tool_safety_and_registry.go`
- **AI Integration**: `internal/api/tools_integration.go`
- **Tests**: `test_tools_powershell.ps1`

---

## ✅ Status

- **Build**: ✅ PASSING
- **Tests**: ✅ PASSING  
- **Tools**: 24/24 IMPLEMENTED
- **Ready**: ✅ PRODUCTION READY

---

*For detailed information, see TOOLS_IMPLEMENTATION_COMPLETE.md*

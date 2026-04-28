# aifiler Tool System

This is the consolidated, deduplicated documentation for the tool system. The split docs were merged here so this file is the canonical reference for status, architecture, usage, testing, and configuration.

## Current Status

The tool system is production ready.

- Tools implemented: 24/24
- Build status: passing
- Test status: passing
- Configuration: ready and user-editable
- Safety validation: implemented
- AI integration: ready
- Testing suite: complete

## Tool Inventory

| Category | Count | Purpose |
| --- | ---: | --- |
| File Operations | 10 | Read, write, inspect, copy, delete, create, and rename files and directories |
| File Analysis | 4 | Scan directories, extract metadata, find patterns, and classify file types |
| Command Execution | 5 | Execute commands with validation, help, and availability checks |
| CLI Discovery | 5 | Discover configured tools, check availability, and gather install hints |
| Total | 24 | Production-ready tool set |

### File Operations

- read_file
- write_file
- list_directory
- check_file_exists
- get_file_metadata
- get_file_type
- copy_file
- delete_file
- create_directory
- rename_file

### File Analysis

- extract_metadata
- scan_directory_content
- get_file_patterns
- analyze_file_types

### Command Execution

- run_command
- run_command_safe
- list_available_commands
- check_command_available
- get_command_help

### CLI Discovery

- check_available_tools
- check_tool_availability
- get_tool_info
- get_installation_hint
- run_sanity_check

## Architecture

### Core Files

| File | Role |
| --- | --- |
| internal/core/tools.go | ToolExecutor framework and registration |
| internal/core/file_operations.go | File system operations |
| internal/core/file_analysis.go | Directory and file analysis tools |
| internal/core/command_execution.go | Command execution tools |
| internal/core/cli_tools_discovery.go | Tool discovery helpers |
| internal/core/tool_safety_and_registry.go | Validation and registry logic |
| internal/core/config.go | Tool configuration model |
| internal/core/toolmanager.go | External tool management |
| internal/api/tools_integration.go | AI provider integration |

### Design Principles

- Configuration-driven tool registry in YAML
- Structured tool metadata for AI consumption
- Validation before execution
- Timeouts and output limits for safety
- Backward compatibility with older helper functions

## Configuration

The main configuration lives in [config.yaml](config.yaml).

### Typical Registry Shape

```yaml
tools:
  registry:
    ffmpeg:
      command: ffmpeg -version
      timeout: 5
      required: false
  groups:
    media: [ffmpeg, magick, exiftool]
    dev: [git, go, node, npm, python, python3]
  task_map:
    organize_images: [exiftool, magick]
    process_videos: [ffmpeg]
```

### Common Tool Groups

- media: ffmpeg, magick, exiftool
- dev: git, go, node, npm, python, python3
- image: exiftool, magick
- video: ffmpeg
- metadata: exiftool

## Safety and Limits

The system includes validation and guardrails for both files and commands.

- Path traversal detection and system-directory protection
- Command injection and dangerous-pattern blocking
- Parameter validation and required-field checks
- File size limits and command length limits
- Default command timeout of 30 seconds
- Command output capped at 64 KB
- File reads capped at 1 MB

## AI Integration

The tool registry is formatted for LLM providers and agent workflows.

### Common Flow

```go
executor := NewToolExecutor(".")
registry := NewToolDefinitionRegistry(executor)
tools := ConvertToolsForAI(registry.GetAllTools())
response := ExecuteToolFromAIRequest(ctx, executor, toolRequest)
```

### Supported Providers

- Anthropic
- OpenAI
- Gemini
- Ollama
- Vercel

## Quick Usage

### Build

```bash
go build ./...
```

### Test

```bash
go test ./internal/core -v
.\test_tools_powershell.ps1
```

### Read a File

```go
executor.ExecuteTool(ctx, "read_file", map[string]interface{}{
    "path": "config.yaml",
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

### Run a Command Safely

```go
executor.ExecuteTool(ctx, "run_command_safe", map[string]interface{}{
    "command": "git status",
    "timeout": 30,
})
```

## Testing

### Go Tests

- TestCheckAvailableTools
- TestCheckToolAvailability
- TestCheckToolGroup

### PowerShell Suite

- File operation tests
- Pattern matching tests
- Directory analysis tests
- Integration tests
- Performance benchmarks
- Error handling tests

## Performance Snapshot

| Operation | Typical Range |
| --- | --- |
| Directory scanning | 50-200 ms for 1000+ files |
| File reading | 1-5 ms for text files |
| Pattern matching | 10-50 ms for 1000+ files |
| Metadata extraction | 5-20 ms per file |

## File Map

| File | Purpose |
| --- | --- |
| config.yaml | Tool registry and grouping |
| test_tools_powershell.ps1 | PowerShell test suite |
| internal/core/tools.go | Core executor framework |
| internal/core/tool_safety_and_registry.go | Validation and registry |
| internal/api/tools_integration.go | AI integration layer |

## Documentation Map

- This file is the merged canonical reference.
- [README.md](README.md) remains the main repository readme.

## Next Steps

1. Run `go build ./...`
2. Run `go test ./internal/core -v`
3. Run `.\test_tools_powershell.ps1`
4. Review [config.yaml](config.yaml)
5. Integrate the tool layer with your AI provider

## Summary

The aifiler tool system provides 24 configured tools across file operations, analysis, command execution, and CLI discovery. It is configured through YAML, validated before execution, and ready for AI provider integration.

## Navigation & Index

### Where to Start

- For first-time users: read this file for the canonical merged documentation, then run `.\\test_tools_powershell.ps1`
- For developers: inspect `internal/core/tools.go`, `internal/core/tool_safety_and_registry.go`, and `internal/api/tools_integration.go`
- For configuration: edit [config.yaml](config.yaml)

### Documentation and Source Map

| Item | Purpose |
| --- | --- |
| FINAL_STATUS.md | Canonical merged documentation |
| README.md | Main repository readme, intentionally excluded from the merge |
| config.yaml | Tool configuration |
| test_tools_powershell.ps1 | Comprehensive test suite |
| internal/core/tools.go | Main ToolExecutor framework |
| internal/core/tool_safety_and_registry.go | Safety validation and registry |
| internal/core/file_operations.go | File operation tools |
| internal/core/file_analysis.go | File analysis tools |
| internal/core/command_execution.go | Command execution tools |
| internal/core/cli_tools_discovery.go | CLI discovery tools |
| internal/api/tools_integration.go | AI provider integration |

### Tool Groups

- media: ffmpeg, magick, exiftool
- dev: git, go, node, npm, python, python3
- image: exiftool, magick
- video: ffmpeg
- metadata: exiftool

### Verification Commands

```bash
go build ./...
go test ./internal/core -v
.\test_tools_powershell.ps1
```

### Common Use Cases

```go
executor.ExecuteTool(ctx, "read_file", map[string]interface{}{
  "path": "config.yaml",
})

executor.ExecuteTool(ctx, "get_file_patterns", map[string]interface{}{
  "directory": ".",
  "pattern": "*.json",
  "pattern_type": "glob",
})

executor.ExecuteTool(ctx, "run_command_safe", map[string]interface{}{
  "command": "git status",
  "timeout": 30,
})
```


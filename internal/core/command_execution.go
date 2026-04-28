package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CommandResult represents the result of executing a command.
type CommandResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Duration string `json:"duration"`
	TimedOut bool   `json:"timed_out"`
}

const DefaultCommandTimeout = 30 * time.Second
const MaxCommandOutput = 65536 // 64KB

// InitCommandExecutionTools registers command execution tools on the executor.
func InitCommandExecutionTools(te *ToolExecutor) {
	// run_command
	te.RegisterTool(
		ToolDefinition{
			Name:        "run_command",
			Description: "Execute a shell command with timeout and output capture",
			Category:    "command_exec",
			Parameters: []ToolParam{
				{
					Name:        "command",
					Type:        "string",
					Description: "Shell command to execute (e.g., 'ffmpeg -version')",
					Required:    true,
				},
				{
					Name:        "timeout",
					Type:        "int",
					Description: "Timeout in seconds. Optional, defaults to 30.",
					Required:    false,
					Default:     30,
				},
				{
					Name:        "cwd",
					Type:        "string",
					Description: "Working directory for command execution. Optional, defaults to current directory.",
					Required:    false,
				},
				{
					Name:        "env",
					Type:        "array",
					Description: "Environment variables as KEY=VALUE strings. Optional.",
					Required:    false,
				},
			},
		},
		te.executeRunCommand,
	)

	// run_command_safe
	te.RegisterTool(
		ToolDefinition{
			Name:        "run_command_safe",
			Description: "Execute a command safely with enhanced validation and security checks",
			Category:    "command_exec",
			Parameters: []ToolParam{
				{
					Name:        "command",
					Type:        "string",
					Description: "Shell command to execute",
					Required:    true,
				},
				{
					Name:        "timeout",
					Type:        "int",
					Description: "Timeout in seconds. Optional, defaults to 30.",
					Required:    false,
					Default:     30,
				},
				{
					Name:        "allow_shell",
					Type:        "bool",
					Description: "Allow shell metacharacters. Optional, defaults to false for safety.",
					Required:    false,
					Default:     false,
				},
			},
		},
		te.executeRunCommandSafe,
	)

	// list_available_commands
	te.RegisterTool(
		ToolDefinition{
			Name:        "list_available_commands",
			Description: "List available commands that can be executed (based on PATH)",
			Category:    "command_exec",
			Parameters: []ToolParam{
				{
					Name:        "pattern",
					Type:        "string",
					Description: "Optional pattern to filter commands (e.g., 'ffmpeg', 'python*')",
					Required:    false,
				},
			},
		},
		te.executeListAvailableCommands,
	)

	// check_command_available
	te.RegisterTool(
		ToolDefinition{
			Name:        "check_command_available",
			Description: "Check if a command is available on the system",
			Category:    "command_exec",
			Parameters: []ToolParam{
				{
					Name:        "command",
					Type:        "string",
					Description: "Command name to check (e.g., 'ffmpeg', 'python')",
					Required:    true,
				},
			},
		},
		te.executeCheckCommandAvailable,
	)

	// get_command_help
	te.RegisterTool(
		ToolDefinition{
			Name:        "get_command_help",
			Description: "Get help/usage information for a command",
			Category:    "command_exec",
			Parameters: []ToolParam{
				{
					Name:        "command",
					Type:        "string",
					Description: "Command name (e.g., 'ffmpeg', 'git')",
					Required:    true,
				},
				{
					Name:        "arg",
					Type:        "string",
					Description: "Help argument (e.g., '--help', '-h', 'help'). Optional, defaults to '--help'.",
					Required:    false,
					Default:     "--help",
				},
			},
		},
		te.executeGetCommandHelp,
	)
}

// ============================================================================
// Command Execution Executors
// ============================================================================

func (te *ToolExecutor) executeRunCommand(ctx context.Context, params map[string]interface{}) ToolResult {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return ToolResult{Success: false, Error: "command parameter is required"}
	}

	command = strings.TrimSpace(command)

	timeout := DefaultCommandTimeout
	if t, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(int(t)) * time.Second
	}

	cwd := te.cwd
	if c, ok := params["cwd"].(string); ok && c != "" {
		cwd = c
	}

	// Parse environment variables
	var env []string
	if envList, ok := params["env"].([]interface{}); ok {
		for _, e := range envList {
			if envStr, ok := e.(string); ok {
				env = append(env, envStr)
			}
		}
	}

	result := executeCommandWithTimeout(ctx, command, timeout, cwd, env)
	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"command":   command,
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
			"duration":  result.Duration,
			"timed_out": result.TimedOut,
		},
	}
}

func (te *ToolExecutor) executeRunCommandSafe(ctx context.Context, params map[string]interface{}) ToolResult {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return ToolResult{Success: false, Error: "command parameter is required"}
	}

	command = strings.TrimSpace(command)
	allowShell := false
	if as, ok := params["allow_shell"].(bool); ok {
		allowShell = as
	}

	// Security check: prevent dangerous patterns if shell not allowed
	if !allowShell {
		dangerous := []string{";", "|", ">", "<", "&", "`", "$", "&&", "||"}
		for _, pattern := range dangerous {
			if strings.Contains(command, pattern) {
				return ToolResult{
					Success: false,
					Error:   fmt.Sprintf("command contains unsafe shell metacharacter '%s' (set allow_shell=true if intentional)", pattern),
				}
			}
		}
	}

	timeout := DefaultCommandTimeout
	if t, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(int(t)) * time.Second
	}

	result := executeCommandWithTimeout(ctx, command, timeout, te.cwd, nil)
	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"command":   command,
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
			"duration":  result.Duration,
			"timed_out": result.TimedOut,
		},
	}
}

func (te *ToolExecutor) executeListAvailableCommands(ctx context.Context, params map[string]interface{}) ToolResult {
	pattern, ok := params["pattern"].(string)
	if !ok {
		pattern = ""
	}

	commands := getCommonCommands(pattern)

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"pattern":  pattern,
			"count":    len(commands),
			"commands": commands,
		},
	}
}

func (te *ToolExecutor) executeCheckCommandAvailable(ctx context.Context, params map[string]interface{}) ToolResult {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return ToolResult{Success: false, Error: "command parameter is required"}
	}

	path, err := exec.LookPath(command)
	available := err == nil

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"command":   command,
			"available": available,
			"path":      path,
			"error":     getErrorString(err),
		},
	}
}

func (te *ToolExecutor) executeGetCommandHelp(ctx context.Context, params map[string]interface{}) ToolResult {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return ToolResult{Success: false, Error: "command parameter is required"}
	}

	arg := "--help"
	if a, ok := params["arg"].(string); ok && a != "" {
		arg = a
	}

	timeout := 10 * time.Second

	result := executeCommandWithTimeout(ctx, fmt.Sprintf("%s %s", command, arg), timeout, te.cwd, nil)

	// If exit code is non-zero but we got output, still return it
	output := result.Stdout
	if output == "" && result.Stderr != "" {
		output = result.Stderr
	}

	return ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"command":   command,
			"help_arg":  arg,
			"output":    output,
			"exit_code": result.ExitCode,
		},
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// executeCommandWithTimeout executes a command with timeout and output capture.
func executeCommandWithTimeout(ctx context.Context, command string, timeout time.Duration, cwd string, env []string) CommandResult {
	start := time.Now()

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Parse command - simple split on spaces for basic commands
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return CommandResult{
			Command:  command,
			ExitCode: -1,
			Stderr:   "empty command",
			Duration: "0s",
		}
	}

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)

	if cwd != "" {
		cmd.Dir = cwd
	}

	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Limit output size
	stdoutStr := truncateOutput(stdout.String(), MaxCommandOutput)
	stderrStr := truncateOutput(stderr.String(), MaxCommandOutput)

	timedOut := cmdCtx.Err() == context.DeadlineExceeded

	return CommandResult{
		Command:  command,
		ExitCode: exitCode,
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		Duration: duration.String(),
		TimedOut: timedOut,
	}
}

// truncateOutput limits output to a maximum size.
func truncateOutput(output string, maxSize int) string {
	if len(output) <= maxSize {
		return output
	}
	return output[:maxSize] + fmt.Sprintf("\n... (truncated, %d more bytes)", len(output)-maxSize)
}

// getCommonCommands returns common commands that might be available.
func getCommonCommands(pattern string) []string {
	common := []string{
		"ffmpeg", "exiftool", "magick", "convert", "identify",
		"git", "go", "node", "npm", "python", "python3",
		"java", "javac", "docker", "curl", "wget",
		"grep", "find", "sed", "awk", "cp", "mv", "rm",
		"mkdir", "ls", "pwd", "cd", "cat", "echo",
	}

	if pattern == "" {
		return common
	}

	// Simple pattern matching
	pattern = strings.ToLower(pattern)
	var filtered []string
	for _, cmd := range common {
		if strings.Contains(strings.ToLower(cmd), pattern) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ToolInfo holds metadata about an external CLI tool.
type ToolInfo struct {
	Name      string
	Available bool
	Version   string
	Path      string
	Error     string
}

// ToolSanityCheck contains a test command and expected output pattern.
type ToolSanityCheck struct {
	Command  string
	Timeout  time.Duration
	Required bool // If true, failure blocks execution
}

// ToolRegistry maps tool names to their sanity check configuration.
var ToolRegistry = map[string]ToolSanityCheck{
	"exiftool": {
		Command:  "exiftool -ver",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"magick": {
		Command:  "magick -version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"ffmpeg": {
		Command:  "ffmpeg -version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"git": {
		Command:  "git --version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"go": {
		Command:  "go version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"node": {
		Command:  "node --version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"npm": {
		Command:  "npm --version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"python": {
		Command:  "python --version",
		Timeout:  5 * time.Second,
		Required: false,
	},
	"python3": {
		Command:  "python3 --version",
		Timeout:  5 * time.Second,
		Required: false,
	},
}

// InstallationHints provides OS-specific installation commands for missing tools.
var InstallationHints = map[string]map[string]string{
	"windows": {
		"exiftool": "winget install OliverBetz.ExifTool",
		"magick":   "winget install ImageMagick.ImageMagick",
		"ffmpeg":   "winget install Gyan.FFmpeg",
		"git":      "winget install Git.Git",
		"go":       "winget install GoLang.Go",
		"node":     "winget install OpenJS.NodeJS",
	},
	"darwin": {
		"exiftool": "brew install exiftool",
		"magick":   "brew install imagemagick",
		"ffmpeg":   "brew install ffmpeg",
		"git":      "brew install git",
		"go":       "brew install go",
		"node":     "brew install node",
	},
	"linux": {
		"exiftool": "sudo apt install libimage-exiftool-perl",
		"magick":   "sudo apt install imagemagick",
		"ffmpeg":   "sudo apt install ffmpeg",
		"git":      "sudo apt install git",
		"go":       "sudo apt install golang-go",
		"node":     "sudo apt install nodejs npm",
	},
}

// DetectToolWithVersion detects a tool on PATH and captures its version.
func DetectToolWithVersion(ctx context.Context, toolName string) ToolInfo {
	info := ToolInfo{
		Name:      toolName,
		Available: false,
	}

	// Check if tool exists on PATH
	path, err := exec.LookPath(toolName)
	if err != nil {
		info.Error = fmt.Sprintf("not found on PATH: %v", err)
		return info
	}

	info.Path = path
	info.Available = true

	// Try to capture version
	check, exists := ToolRegistry[toolName]
	if !exists {
		// Tool not in registry, but it exists on PATH
		return info
	}

	version, verErr := getToolVersion(ctx, check.Command, check.Timeout)
	if verErr != nil {
		info.Error = fmt.Sprintf("version check failed: %v", verErr)
	} else {
		info.Version = strings.TrimSpace(version)
	}

	return info
}

// getToolVersion executes a version check command with timeout.
func getToolVersion(ctx context.Context, command string, timeout time.Duration) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("version check timeout")
		}
		// Some tools write version to stderr, so check both
		if errOut.Len() > 0 {
			return errOut.String(), nil
		}
		return "", fmt.Errorf("version command failed: %v", err)
	}

	return out.String(), nil
}

// RunSanityCheck runs a non-destructive test for a tool.
func RunSanityCheck(ctx context.Context, toolName string) (bool, string, error) {
	check, exists := ToolRegistry[toolName]
	if !exists {
		return true, "no sanity check defined", nil // Tool exists but no check defined
	}

	cmdCtx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	parts := strings.Fields(check.Command)
	if len(parts) == 0 {
		return false, "", fmt.Errorf("invalid sanity check command")
	}

	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return false, "", fmt.Errorf("sanity check timeout")
		}
		return false, errOut.String(), fmt.Errorf("sanity check failed: %v", err)
	}

	return true, out.String(), nil
}

// ReportToolAvailability provides detailed availability and version information.
func ReportToolAvailability(ctx context.Context, toolNames []string) string {
	var report strings.Builder
	report.WriteString("Tool Availability Report:\n")

	if len(toolNames) == 0 {
		report.WriteString("No tools detected on PATH\n")
		return report.String()
	}

	for _, name := range toolNames {
		info := DetectToolWithVersion(ctx, name)
		if info.Available {
			report.WriteString(fmt.Sprintf("✓ %s", info.Name))
			if info.Version != "" {
				report.WriteString(fmt.Sprintf(" (version: %s)", info.Version))
			}
			report.WriteString("\n")
		}
	}

	return report.String()
}

// SuggestInstallation provides installation hints for missing tools.
func SuggestInstallation(toolName string) string {
	osKey := runtime.GOOS
	hints, exists := InstallationHints[osKey]
	if !exists {
		return fmt.Sprintf("No installation hint available for %s on %s", toolName, osKey)
	}

	cmd, found := hints[toolName]
	if !found {
		return fmt.Sprintf("No installation hint available for %s on %s", toolName, osKey)
	}

	return fmt.Sprintf("To install %s on %s, run:\n  %s", toolName, osKey, cmd)
}

// ClassifyCommandError categorizes execution errors.
type CommandError struct {
	Type    string // "notfound", "timeout", "failed", "other"
	Message string
}

// ClassifyError categorizes a command execution error.
func ClassifyError(err error) CommandError {
	if err == nil {
		return CommandError{Type: "none"}
	}

	errStr := err.Error()

	// Check for timeout
	if strings.Contains(errStr, "deadline exceeded") || strings.Contains(errStr, "context deadline") {
		return CommandError{
			Type:    "timeout",
			Message: "Command execution timed out",
		}
	}

	// Check for not found
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "executable file not found") {
		return CommandError{
			Type:    "notfound",
			Message: "Command not found on PATH",
		}
	}

	// Check for exit code errors
	if strings.Contains(errStr, "exit status") {
		return CommandError{
			Type:    "failed",
			Message: errStr,
		}
	}

	return CommandError{
		Type:    "other",
		Message: errStr,
	}
}

// QuoteForShell properly quotes a string for shell execution.
// This handles file paths with spaces and special characters.
func QuoteForShell(s string) string {
	// On Windows, use double quotes
	// On Unix, use single quotes to prevent expansion
	if runtime.GOOS == "windows" {
		// Escape any double quotes in the string
		escaped := strings.ReplaceAll(s, `"`, `\"`)
		return `"` + escaped + `"`
	}
	// Unix: single quotes prevent all expansion
	escaped := strings.ReplaceAll(s, "'", "'\"'\"'")
	return "'" + escaped + "'"
}

// ParseCommandLineWithShellQuotes parses a command line, respecting quoted strings.
// Unlike strings.Fields(), this preserves quoted strings as single arguments.
func ParseCommandLineWithShellQuotes(command string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)
	escape := false

	for _, r := range command {
		if escape {
			current.WriteRune(r)
			escape = false
			continue
		}

		if r == '\\' {
			current.WriteRune(r)
			escape = true
			continue
		}

		if !inQuote {
			if r == '"' || r == '\'' {
				inQuote = true
				quoteChar = r
				continue
			}
			if r == ' ' || r == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
				continue
			}
		} else {
			if r == quoteChar {
				inQuote = false
				continue
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

package cmds

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"aifiler/internal/core"
)

type promptTurn struct {
	Role    string
	Content string
}

func (a *App) runDynamicPrompt(ctx context.Context, prompt string) int {
	currentPrompt := strings.TrimSpace(prompt)
	if currentPrompt == "" {
		a.printHelp()
		return 0
	}

	client, provider, model, err := a.newClient("", "")
	if err != nil {
		core.ErrorStyle.Printf("failed to initialize model client: %v\n", err)
		return 1
	}

	availableTools := detectAvailableToolsWithVersions(ctx, []string{"ffmpeg", "magick", "exiftool", "git", "go", "node", "npm", "python", "python3"})
	var conversation []promptTurn

	for {
		workspaceContext := core.BuildWorkspaceContext(a.maxDepth, a.showAll)
		workspaceContext = augmentWorkspaceContext(workspaceContext, availableTools, conversation)
		thinking := core.StartThinking("AI is thinking")

		finalPrompt := currentPrompt
		isExplain := false
		if strings.HasPrefix(currentPrompt, "/") {
			parts := strings.SplitN(currentPrompt[1:], " ", 2)
			intent := strings.ToLower(parts[0])
			rest := ""
			if len(parts) > 1 {
				rest = parts[1]
			}
			if intent == "explain" {
				isExplain = true
				finalPrompt = fmt.Sprintf("EXPLAIN THE FOLLOWING: %s\nProvide a clear, concise explanation. Do not return JSON actions.", rest)
			} else {
				finalPrompt = fmt.Sprintf("FORCE OPERATION TYPE: %s\nUSER REQUEST: %s", strings.ToUpper(intent), rest)
			}
		}

		response, err := client.Prompt(ctx, buildDynamicPrompt(finalPrompt, workspaceContext, a.force && !isExplain, a.agent))
		statusMessage := "AI response ready\nReviewing the next step"
		if a.agent {
			statusMessage = "AI response ready\nChecking installed CLI tools before proposing actions"
		}
		thinking.Stop(statusMessage)
		if err != nil {
			core.ErrorStyle.Printf("%s AI request failed: %v\n", core.ErrorIcon, err)
			return 1
		}

		conversation = append(conversation,
			promptTurn{Role: "user", Content: finalPrompt},
			promptTurn{Role: "assistant", Content: response},
		)

		var plan core.AIPlan
		var parseErr error
		if isExplain {
			parseErr = fmt.Errorf("explain mode")
		} else {
			plan, parseErr = core.ParsePlan(response)
		}

		if parseErr == nil && a.agent {
			plan = injectToolAvailabilityChecks(plan)
		}

		if parseErr != nil && !isExplain {
			coerceThinking := core.StartThinking("AI is restructuring response as plan")
			coerced, coerceErr := client.Prompt(ctx, buildPlanCoercionPrompt(currentPrompt, response))
			coerceThinking.Stop("")
			if coerceErr == nil {
				if repairedPlan, repairedErr := core.ParsePlan(coerced); repairedErr == nil {
					plan = repairedPlan
					parseErr = nil
					if a.agent {
						plan = injectToolAvailabilityChecks(plan)
					}
					conversation[len(conversation)-1].Content = coerced
				}
			}
		}

		core.MutedStyle.Printf("provider=%s model=%s\n", provider, model)
		if parseErr == nil && len(plan.Operations) > 0 {
			result := ApplyPlanWithApproval(plan)
			if strings.TrimSpace(result.NextPrompt) == "" {
				return result.ExitCode
			}
			currentPrompt = strings.TrimSpace(result.NextPrompt)
			continue
		}
		if parseErr == nil && len(plan.Operations) == 0 {
			if a.force {
				core.WarnStyle.Println("AI failed to propose operations even with -force flag.")
			} else {
				core.WarnStyle.Println("No operations proposed for this prompt.")
				fmt.Println("Try a more specific prompt, or use -force to insist on a suggestion.")
			}
			return 0
		}

		fmt.Println(response)
		return 0
	}
}

func detectAvailableToolsWithVersions(ctx context.Context, candidates []string) []core.ToolInfo {
	tools := make([]core.ToolInfo, 0, len(candidates))
	for _, candidate := range candidates {
		name := strings.TrimSpace(candidate)
		if name == "" {
			continue
		}
		info := core.DetectToolWithVersion(ctx, name)
		if info.Available {
			tools = append(tools, info)
		}
	}
	return tools
}

// Deprecated: use detectAvailableToolsWithVersions instead
func detectAvailableTools(candidates []string) []string {
	available := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		name := strings.TrimSpace(candidate)
		if name == "" {
			continue
		}
		if _, err := exec.LookPath(name); err == nil {
			available = append(available, name)
		}
	}
	return available
}

func augmentWorkspaceContext(workspaceContext string, availableTools []core.ToolInfo, conversation []promptTurn) string {
	var builder strings.Builder
	builder.WriteString(workspaceContext)

	if len(availableTools) > 0 {
		builder.WriteString("\n\nDetected CLI tools:\n")
		for _, tool := range availableTools {
			builder.WriteString("- ")
			builder.WriteString(tool.Name)
			if tool.Version != "" {
				builder.WriteString(" (")
				builder.WriteString(strings.TrimSpace(tool.Version))
				builder.WriteString(")")
			}
			builder.WriteString("\n")
		}
	} else {
		builder.WriteString("\n\nDetected CLI tools:\n- none of the common tools were found on PATH\n")
	}

	if len(conversation) > 0 {
		builder.WriteString("\n\nRecent conversation context:\n")
		start := 0
		if len(conversation) > 6 {
			start = len(conversation) - 6
		}
		for _, turn := range conversation[start:] {
			builder.WriteString("- ")
			builder.WriteString(turn.Role)
			builder.WriteString(": ")
			builder.WriteString(strings.TrimSpace(turn.Content))
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func buildDynamicPrompt(userPrompt, workspaceContext string, force bool, agentMode bool) string {
	forceText := ""
	if force {
		forceText = "\nIMPORTANT: You MUST propose at least one filesystem operation in the JSON format below. Do not return plain text."
	}
	agentText := ""
	if agentMode {
		agentText = "\nAgent mode is enabled. Prefer command-driven solutions and installed CLI tools over built-in assumptions. When a CLI tool is needed, verify it exists using the detected tool list or an explicit non-interactive check before relying on it."
	}
	return fmt.Sprintf(`You are a local agent operating inside a workspace.%s%s
Use run_command for non-interactive shell commands when they are necessary.
Before relying on a CLI tool, inspect the detected tool list in the workspace context and prefer tools that are already installed.
If a requested tool is not detected, do not assume it exists; either choose a different approach or include a command that checks availability first.
If the user request requires filesystem or command actions, return STRICT JSON only in this format:
{"summary":"brief explanation of plan","operations":[{"type":"create_dir|create_file|update_file|rename|delete|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
If the request is informational only, return a normal text response.%s
Rules for action plans:
- infer file/folder targets from workspace context; do not ask user to describe structure
- paths must be relative and within current directory
- use update_file when modifying existing files
- use run_command only when necessary and keep commands non-interactive
- when agent mode is enabled, prefer run_command for transformations that external tools can handle better than built-in file editing
- no markdown fences when returning JSON
- for text responses, DO NOT use markdown format (like bold, headers, or bullet lists); use plain text only
- for workspace context, lines starting with symbols (like ◆, ▸, ▫) denote types; the symbol is a label, NOT part of the path name
Workspace context:
%s
User request: %s`, forceText, agentText, forceText, workspaceContext, userPrompt)
}

func buildPlanCoercionPrompt(userPrompt, modelResponse string) string {
	return fmt.Sprintf(`Convert the following into STRICT JSON only in this exact format:
{"summary":"brief explanation of plan","operations":[{"type":"create_dir|create_file|update_file|rename|delete|run_command","path":"relative/path","from":"relative/path","to":"relative/path","content":"optional","command":"optional"}]}
Rules:
- no explanation text
- no markdown fences
- paths must be relative
User request: %s
Previous response to convert:
%s`, userPrompt, modelResponse)
}

func injectToolAvailabilityChecks(plan core.AIPlan) core.AIPlan {
	if len(plan.Operations) == 0 {
		return plan
	}

	updated := make([]core.Operation, 0, len(plan.Operations)*2)
	for _, op := range plan.Operations {
		if shouldPreflightToolCheck(op) {
			toolName := firstCommandToken(op.Command)
			if toolName != "" {
				// Add availability check
				updated = append(updated, core.Operation{
					Type:    "run_command",
					Command: availabilityCheckCommand(toolName),
				})
				// Add sanity check if defined
				updated = append(updated, core.Operation{
					Type:    "run_command",
					Command: sanitycheckCommand(toolName),
				})
			}
		}
		updated = append(updated, op)
	}
	plan.Operations = updated
	return plan
}

// sanitycheckCommand generates a sanity check command for the tool.
func sanitycheckCommand(toolName string) string {
	if check, exists := core.ToolRegistry[toolName]; exists {
		return check.Command
	}
	return ""
}

func shouldPreflightToolCheck(op core.Operation) bool {
	if strings.ToLower(strings.TrimSpace(op.Type)) != "run_command" {
		return false
	}
	command := strings.TrimSpace(op.Command)
	if command == "" {
		return false
	}
	toolName := firstCommandToken(command)
	if toolName == "" {
		return false
	}
	if isAvailabilityProbe(toolName) {
		return false
	}
	if strings.ContainsAny(toolName, "/\\") {
		return false
	}
	return true
}

func firstCommandToken(command string) string {
	fields := strings.Fields(strings.TrimSpace(command))
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], `"'`)
}

func availabilityCheckCommand(toolName string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("where.exe %s", toolName)
	}
	return fmt.Sprintf("command -v %s", toolName)
}

func isAvailabilityProbe(commandName string) bool {
	switch strings.ToLower(strings.TrimSpace(commandName)) {
	case "where", "where.exe", "which", "command", "get-command":
		return true
	default:
		return false
	}
}

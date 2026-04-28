package core

import (
	"context"
	"testing"
)

// TestCheckAvailableTools tests the check_available_tools tool.
func TestCheckAvailableTools(t *testing.T) {
	executor := NewToolExecutor(".")
	ctx := context.Background()

	result := executor.ExecuteTool(ctx, "check_available_tools", map[string]interface{}{})

	if !result.Success {
		t.Fatalf("ExecuteTool failed: %s", result.Error)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Result data is not a map")
	}

	if _, exists := data["tools"]; !exists {
		t.Fatalf("Result missing 'tools' key")
	}

	if _, exists := data["count"]; !exists {
		t.Fatalf("Result missing 'count' key")
	}

	t.Logf("Found %v available tools", data["count"])
}

// TestCheckToolAvailability tests the check_tool_availability tool.
func TestCheckToolAvailability(t *testing.T) {
	executor := NewToolExecutor(".")
	ctx := context.Background()

	// Test with a tool that may or may not be available
	result := executor.ExecuteTool(ctx, "check_tool_availability", map[string]interface{}{
		"tool_name": "git",
	})

	if !result.Success {
		t.Fatalf("ExecuteTool failed: %s", result.Error)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Result data is not a map")
	}

	if _, exists := data["available"]; !exists {
		t.Fatalf("Result missing 'available' key")
	}

	t.Logf("git available: %v", data["available"])
}

// TestGetToolInfo tests the get_tool_info tool.
func TestGetToolInfo(t *testing.T) {
	executor := NewToolExecutor(".")
	ctx := context.Background()

	// Test with git (commonly available)
	result := executor.ExecuteTool(ctx, "get_tool_info", map[string]interface{}{
		"tool_name": "git",
	})

	// Tool might not be available in test environment, but executor should work
	data, ok := result.Data.(map[string]interface{})
	if ok {
		t.Logf("Tool info retrieved: %v", data)
	}
}

// TestGetInstallationHint tests the get_installation_hint tool.
func TestGetInstallationHint(t *testing.T) {
	executor := NewToolExecutor(".")
	ctx := context.Background()

	result := executor.ExecuteTool(ctx, "get_installation_hint", map[string]interface{}{
		"tool_name": "ffmpeg",
	})

	if !result.Success {
		t.Fatalf("ExecuteTool failed: %s", result.Error)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Result data is not a map")
	}

	if hint, exists := data["hint"]; exists {
		t.Logf("Installation hint for ffmpeg: %v", hint)
	}
}

// TestDiscoverAllTools tests the DiscoverAllTools function.
func TestDiscoverAllTools(t *testing.T) {
	ctx := context.Background()
	report := DiscoverAllTools(ctx)

	if report.AvailableCount+report.MissingCount == 0 {
		t.Logf("No tools detected in test environment")
		return
	}

	t.Logf("Available tools: %d", report.AvailableCount)
	t.Logf("Missing tools: %d", report.MissingCount)

	formatted := FormatToolsReport(report)
	t.Logf("Report:\n%s", formatted)
}

// TestCheckToolGroup tests tool grouping functionality.
func TestCheckToolGroup(t *testing.T) {
	ctx := context.Background()

	mediaTools := CheckToolGroup(ctx, "media")
	if len(mediaTools) > 0 {
		t.Logf("Media tools found: %d", len(mediaTools))
		for name, info := range mediaTools {
			t.Logf("  %s: available=%v", name, info.Available)
		}
	}
}

// TestSuggestToolsForTask tests task-based tool suggestions.
func TestSuggestToolsForTask(t *testing.T) {
	ctx := context.Background()

	missing := SuggestToolsForTask(ctx, "organize_images")
	t.Logf("Missing tools for 'organize_images': %v", missing)

	missing = SuggestToolsForTask(ctx, "process_videos")
	t.Logf("Missing tools for 'process_videos': %v", missing)
}

// TestToolCapabilities tests capability information.
func TestToolCapabilities(t *testing.T) {
	caps := ToolCapabilities()

	if len(caps) == 0 {
		t.Fatal("ToolCapabilities returned empty map")
	}

	for tool, capabilities := range caps {
		if len(capabilities) == 0 {
			t.Logf("Warning: %s has no capabilities defined", tool)
		} else {
			t.Logf("%s can: %v", tool, capabilities)
		}
	}
}

// TestToolsForAIAgent tests AI agent formatted tool info.
func TestToolsForAIAgent(t *testing.T) {
	ctx := context.Background()
	info := ToolsForAIAgent(ctx)

	available, ok := info["available"].([]string)
	if !ok {
		t.Fatal("Missing 'available' key")
	}

	missing, ok := info["missing"].([]string)
	if !ok {
		t.Fatal("Missing 'missing' key")
	}

	t.Logf("Available for AI: %d tools", len(available))
	t.Logf("Missing for AI: %d tools", len(missing))
}

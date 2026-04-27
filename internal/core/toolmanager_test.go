package core

import (
	"testing"
)

func TestQuoteForShell(t *testing.T) {
	tests := []struct {
		input    string
		hasQuote bool
	}{
		{"simple", true},
		{"path with spaces", true},
		{"path/to/file", true},
	}

	for _, tt := range tests {
		result := QuoteForShell(tt.input)
		if tt.hasQuote && (result[0] == '"' || result[0] == '\'') {
			// OK
		} else {
			t.Errorf("QuoteForShell(%q) didn't add quotes", tt.input)
		}
	}
}

func TestParseCommandLineWithShellQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			`command -o "output file.txt" -i input.txt`,
			[]string{"command", "-o", "output file.txt", "-i", "input.txt"},
		},
		{
			`ffmpeg -i 'my file.mp4' output.mp4`,
			[]string{"ffmpeg", "-i", "my file.mp4", "output.mp4"},
		},
		{
			`simple command`,
			[]string{"simple", "command"},
		},
	}

	for _, tt := range tests {
		result := ParseCommandLineWithShellQuotes(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("ParseCommandLineWithShellQuotes(%q) got %d args, expected %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i, arg := range result {
			if arg != tt.expected[i] {
				t.Errorf("ParseCommandLineWithShellQuotes(%q)[%d] = %q, expected %q", tt.input, i, arg, tt.expected[i])
			}
		}
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		input    error
		expected string
	}{
		{nil, "none"},
	}

	for _, tt := range tests {
		result := ClassifyError(tt.input)
		if result.Type != tt.expected {
			t.Errorf("ClassifyError(%v) = %s, expected %s", tt.input, result.Type, tt.expected)
		}
	}
}

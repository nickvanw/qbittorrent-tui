// Package testutil provides testing utilities for UI components
package testutil

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// SendKey sends a key message to a Bubble Tea model and returns the updated model
func SendKey(t *testing.T, model tea.Model, key string) tea.Model {
	t.Helper()

	var msg tea.KeyMsg

	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "pgup":
		msg = tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		msg = tea.KeyMsg{Type: tea.KeyPgDown}
	case "home":
		msg = tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		msg = tea.KeyMsg{Type: tea.KeyEnd}
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+a":
		msg = tea.KeyMsg{Type: tea.KeyCtrlA}
	case "ctrl+r":
		msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	case " ", "space":
		msg = tea.KeyMsg{Type: tea.KeySpace}
	default:
		// Regular character key
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}

	newModel, _ := model.Update(msg)
	return newModel
}

// SendKeys sends multiple key messages in sequence
func SendKeys(t *testing.T, model tea.Model, keys ...string) tea.Model {
	t.Helper()

	for _, key := range keys {
		model = SendKey(t, model, key)
	}
	return model
}

// TypeString types a string character by character
func TypeString(t *testing.T, model tea.Model, text string) tea.Model {
	t.Helper()

	for _, char := range text {
		model = SendKey(t, model, string(char))
	}
	return model
}

// AssertViewContains checks if the view contains expected text
func AssertViewContains(t *testing.T, model interface{ View() string }, expected string) {
	t.Helper()

	view := model.View()
	assert.Contains(t, view, expected, "View should contain: %s", expected)
}

// AssertViewNotContains checks if the view does not contain text
func AssertViewNotContains(t *testing.T, model interface{ View() string }, unexpected string) {
	t.Helper()

	view := model.View()
	assert.NotContains(t, view, unexpected, "View should not contain: %s", unexpected)
}

// AssertViewLines checks if the view has the expected number of lines
func AssertViewLines(t *testing.T, model interface{ View() string }, expectedLines int) {
	t.Helper()

	view := model.View()
	lines := strings.Split(view, "\n")
	assert.Equal(t, expectedLines, len(lines), "View should have %d lines", expectedLines)
}

// StripANSI removes ANSI escape sequences from a string
func StripANSI(s string) string {
	// More comprehensive ANSI stripper for testing
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			// End escape sequence on various terminators
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '~' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// GetViewLines returns the lines of a view with ANSI codes stripped
func GetViewLines(model interface{ View() string }) []string {
	view := model.View()
	lines := strings.Split(view, "\n")

	cleanLines := make([]string, len(lines))
	for i, line := range lines {
		cleanLines[i] = StripANSI(line)
	}

	return cleanLines
}

// AssertLineWidth checks that all lines in the view fit within maxWidth
func AssertLineWidth(t *testing.T, model interface{ View() string }, maxWidth int) {
	t.Helper()

	lines := GetViewLines(model)
	for i, line := range lines {
		// Use rune count for display width, not byte count
		displayWidth := len([]rune(line))
		assert.LessOrEqual(t, displayWidth, maxWidth,
			"Line %d exceeds max width %d: %s", i+1, maxWidth, line)
	}
}

// FindLineContaining returns the first line containing the substring
func FindLineContaining(model interface{ View() string }, substr string) (string, bool) {
	lines := GetViewLines(model)
	for _, line := range lines {
		if strings.Contains(line, substr) {
			return line, true
		}
	}
	return "", false
}

// CountOccurrences counts how many times a substring appears in the view
func CountOccurrences(model interface{ View() string }, substr string) int {
	view := model.View()
	return strings.Count(view, substr)
}

// SimulateResize sends a window size message to the model
func SimulateResize(t *testing.T, model tea.Model, width, height int) tea.Model {
	t.Helper()

	msg := tea.WindowSizeMsg{
		Width:  width,
		Height: height,
	}

	newModel, _ := model.Update(msg)
	return newModel
}

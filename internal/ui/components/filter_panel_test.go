package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFilterPanel_BasicFunctionality(t *testing.T) {
	panel := NewFilterPanel()

	// Test initial state
	assert.Equal(t, FilterModeNone, panel.mode)
	assert.True(t, panel.filter.IsEmpty())

	// Test entering search mode with "/"
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.Equal(t, FilterModeSearch, panel.mode)

	// Test entering search mode with "f"
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	assert.Equal(t, FilterModeSearch, panel.mode)
}

func TestFilterPanel_SearchMode(t *testing.T) {
	panel := NewFilterPanel()

	// Enter search mode
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	// Type some text
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'e', 's', 't'}})

	// Apply search with Enter
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, FilterModeNone, panel.mode)
	assert.Equal(t, "test", panel.filter.Search)
}

func TestFilterPanel_ClearFilters(t *testing.T) {
	panel := NewFilterPanel()

	// Set some filters
	panel.filter.Search = "test"
	panel.filter.Category = "movies"
	panel.filter.States = []string{"downloading"}

	assert.False(t, panel.filter.IsEmpty())

	// Clear filters with "x"
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	assert.True(t, panel.filter.IsEmpty())
	assert.Equal(t, "", panel.filter.Search)
	assert.Equal(t, "", panel.filter.Category)
	assert.Empty(t, panel.filter.States)
}

func TestFilterPanel_StateFilterMode(t *testing.T) {
	panel := NewFilterPanel()

	// Enter state filter mode
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Equal(t, FilterModeState, panel.mode)
	assert.Equal(t, 0, panel.cursor)

	// Move cursor down
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assert.Equal(t, 1, panel.cursor)

	// Toggle selection
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	assert.Contains(t, panel.filter.States, panel.availableStates[1])

	// Exit with Escape
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, FilterModeNone, panel.mode)
}

func TestFilterPanel_ViewRendering(t *testing.T) {
	panel := NewFilterPanel()

	// Test normal mode view
	view := panel.View()
	assert.Contains(t, view, "No active filters")
	assert.Contains(t, view, "Press:")

	// Test search mode view
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	view = panel.View()
	assert.Contains(t, view, "Search:")
	assert.Contains(t, view, "Enter save")

	// Test state filter mode view
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	view = panel.View()
	assert.Contains(t, view, "Select State:")
	assert.Contains(t, view, "downloading")
}

func TestFilterPanel_InputMode(t *testing.T) {
	panel := NewFilterPanel()

	// Should not be in input mode initially
	assert.False(t, panel.IsInInputMode())

	// Should be in input mode when in search mode
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, panel.IsInInputMode())

	// Should not be in input mode when in state filter mode
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.False(t, panel.IsInInputMode())

	// Should exit input mode on escape
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, panel.IsInInputMode())
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, panel.IsInInputMode())

	// Should exit input mode on enter
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, panel.IsInInputMode())
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, panel.IsInInputMode())
}

func TestFilterPanel_InteractiveMode(t *testing.T) {
	panel := NewFilterPanel()

	// Should not be in interactive mode initially
	assert.False(t, panel.IsInInteractiveMode())

	// Should be in interactive mode when in search mode
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, panel.IsInInteractiveMode())

	// Should be in interactive mode when in state filter mode
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.True(t, panel.IsInInteractiveMode())

	// Should be in interactive mode when in category filter mode
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	assert.True(t, panel.IsInInteractiveMode())

	// Should exit interactive mode on escape
	panel = NewFilterPanel()
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.True(t, panel.IsInInteractiveMode())
	panel, _ = panel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, panel.IsInInteractiveMode())
}

func TestFilterPanel_ResponsiveLayout(t *testing.T) {
	fp := NewFilterPanel()

	// Test different widths
	testWidths := []int{0, 80, 120, 200, 300}

	for _, width := range testWidths {
		fp.SetDimensions(width, 3)

		// Test normal mode rendering
		view := fp.View()
		assert.NotEmpty(t, view, "Empty view for width %d", width)

		// Verify view content makes sense
		if width == 0 {
			// Should use fallback layout
			assert.Contains(t, view, "No active filters")
		} else {
			// Should use responsive layout
			assert.Contains(t, view, "No active filters")

			// Check that content doesn't obviously overflow
			lines := strings.Split(view, "\n")
			for _, line := range lines {
				cleanLine := stripANSI(line)
				// Allow some reasonable tolerance for styling
				if len(cleanLine) > width+20 {
					t.Errorf("Line length %d significantly exceeds width %d for line: %q",
						len(cleanLine), width, cleanLine)
				}
			}
		}
	}

	// Test with active filters for more complex layout
	fp.filter.Search = "test"
	fp.filter.States = []string{"downloading", "uploading"}
	fp.filter.Category = "movies"

	for _, width := range testWidths {
		fp.SetDimensions(width, 3)
		view := fp.View()
		assert.NotEmpty(t, view)

		// Should contain filter information
		if width > 50 {
			assert.Contains(t, view, "Active")
			assert.Contains(t, view, "test")
		}
	}
}

func TestFilterPanel_ResponsiveFilterOptions(t *testing.T) {
	fp := NewFilterPanel()

	// Set up some available states for testing
	fp.availableStates = []string{
		"active", "downloading", "uploading", "completed",
		"paused", "queued", "stalled", "checking", "error",
	}

	// Test different widths and their visible option counts
	testCases := []struct {
		width    int
		expected int
	}{
		{0, 3},   // No width set, use minimum
		{50, 3},  // Narrow, use minimum (< 120)
		{100, 3}, // Still narrow (< 120)
		{150, 4}, // Medium (120-159)
		{180, 5}, // Wide (160-199)
		{200, 6}, // Very wide (>= 200)
		{300, 6}, // Extra wide, still capped at maximum
	}

	for _, tc := range testCases {
		fp.SetDimensions(tc.width, 5)
		maxVisible := fp.calculateMaxVisibleOptions()

		if maxVisible != tc.expected {
			t.Errorf("Width %d: expected %d visible options, got %d",
				tc.width, tc.expected, maxVisible)
		}
	}
}

// stripANSI removes ANSI escape sequences for length calculation
func stripANSI(s string) string {
	// Simple ANSI stripper for testing
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

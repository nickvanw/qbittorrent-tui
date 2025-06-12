package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nickvanw/qbittorrent-tui/internal/filter"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/styles"
)

// FilterMode represents the current filter editing mode
type FilterMode int

const (
	FilterModeNone FilterMode = iota
	FilterModeSearch
	FilterModeState
	FilterModeCategory
	FilterModeTracker
	FilterModeTag
)

// FilterPanel handles torrent filtering
type FilterPanel struct {
	filter       filter.Filter
	backupFilter filter.Filter // Backup for cancel operation
	mode         FilterMode
	searchInput  textinput.Model
	width        int
	height       int

	// Available options for selection
	availableStates     []string
	availableCategories []string
	availableTrackers   []string
	availableTags       []string

	// Selection cursor for list modes
	cursor int
}

// NewFilterPanel creates a new filter panel
func NewFilterPanel() *FilterPanel {
	searchInput := textinput.New()
	searchInput.Placeholder = "Type to search..."
	searchInput.CharLimit = 100

	return &FilterPanel{
		searchInput: searchInput,
		mode:        FilterModeNone,
		availableStates: []string{
			"active",      // Custom state for actively transferring
			"downloading", // Actively downloading
			"uploading",   // Actively seeding/uploading
			"completed",   // Download completed
			"paused",      // Any paused state
			"queued",      // Any queued state
			"stalled",     // Stalled (no peers/seeds)
			"checking",    // Checking files
			"error",       // Error state
			"allocating",  // Allocating disk space
			"metaDL",      // Downloading metadata
			"moving",      // Moving files
		},
	}
}

// SetAvailableOptions updates the available filter options
func (f *FilterPanel) SetAvailableOptions(categories, trackers, tags []string) {
	f.availableCategories = categories
	f.availableTrackers = trackers
	f.availableTags = tags
}

// GetFilter returns the current filter
func (f *FilterPanel) GetFilter() filter.Filter {
	return f.filter
}

// IsInInputMode returns true if the filter panel is currently accepting text input
func (f *FilterPanel) IsInInputMode() bool {
	return f.mode == FilterModeSearch
}

// IsInInteractiveMode returns true if the filter panel needs to handle navigation keys
func (f *FilterPanel) IsInInteractiveMode() bool {
	return f.mode != FilterModeNone
}

// Update handles messages
func (f *FilterPanel) Update(msg tea.Msg) (*FilterPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch f.mode {
		case FilterModeSearch:
			switch msg.String() {
			case "esc":
				// Esc cancels search and restores previous value
				f.searchInput.SetValue(f.backupFilter.Search)
				f.mode = FilterModeNone
				f.searchInput.Blur()
			case "enter":
				f.filter.Search = f.searchInput.Value()
				f.mode = FilterModeNone
				f.searchInput.Blur()
			default:
				f.searchInput, cmd = f.searchInput.Update(msg)
			}

		case FilterModeState, FilterModeCategory, FilterModeTracker, FilterModeTag:
			switch msg.String() {
			case "esc":
				// Esc cancels and restores previous filter state
				f.filter = f.backupFilter
				f.mode = FilterModeNone
				f.cursor = 0
			case "up", "k":
				f.moveCursorUp()
			case "down", "j":
				f.moveCursorDown()
			case "enter":
				// Enter exits interactive mode (marking as "done")
				f.mode = FilterModeNone
				f.cursor = 0
			case " ":
				f.toggleSelection()
			case "a":
				f.selectAll()
			case "n":
				f.selectNone()
			}

		case FilterModeNone:
			switch msg.String() {
			case "/", "f":
				f.backupFilter = f.filter
				f.searchInput.SetValue(f.filter.Search)
				f.mode = FilterModeSearch
				f.searchInput.Focus()
				cmd = textinput.Blink
			case "s":
				f.backupFilter = f.filter
				f.mode = FilterModeState
				f.cursor = 0
			case "c":
				f.backupFilter = f.filter
				f.mode = FilterModeCategory
				f.cursor = 0
			case "t":
				f.backupFilter = f.filter
				f.mode = FilterModeTracker
				f.cursor = 0
			case "g":
				f.backupFilter = f.filter
				f.mode = FilterModeTag
				f.cursor = 0
			case "x":
				f.clearFilters()
			}
		}
	}

	return f, cmd
}

// View renders the filter panel
func (f *FilterPanel) View() string {
	switch f.mode {
	case FilterModeSearch:
		return f.renderSearchMode()
	case FilterModeState:
		return f.renderListMode("State", f.availableStates, f.filter.States)
	case FilterModeCategory:
		return f.renderListMode("Category", f.availableCategories, []string{f.filter.Category})
	case FilterModeTracker:
		return f.renderListMode("Tracker", f.availableTrackers, f.filter.Trackers)
	case FilterModeTag:
		return f.renderListMode("Tag", f.availableTags, f.filter.Tags)
	default:
		return f.renderNormalMode()
	}
}

// renderNormalMode renders the filter panel in normal mode
func (f *FilterPanel) renderNormalMode() string {
	if f.width == 0 {
		// Fallback to simple layout if width not set
		return f.renderSimpleNormalMode()
	}

	// Calculate how much space we have
	availableWidth := f.width

	// Build filter section
	var filterSection string
	if !f.filter.IsEmpty() {
		var filterParts []string
		filterParts = append(filterParts, styles.TitleStyle.Render("Active:"))

		if f.filter.Search != "" {
			filterParts = append(filterParts, fmt.Sprintf("Search=%s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(f.filter.Search)))
		}
		if len(f.filter.States) > 0 {
			states := strings.Join(f.filter.States, ",")
			filterParts = append(filterParts, fmt.Sprintf("States=%s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(states)))
		}
		if f.filter.Category != "" {
			filterParts = append(filterParts, fmt.Sprintf("Category=%s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(f.filter.Category)))
		}
		if len(f.filter.Trackers) > 0 {
			trackers := strings.Join(f.filter.Trackers, ",")
			filterParts = append(filterParts, fmt.Sprintf("Trackers=%s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(trackers)))
		}
		if len(f.filter.Tags) > 0 {
			tags := strings.Join(f.filter.Tags, ",")
			filterParts = append(filterParts, fmt.Sprintf("Tags=%s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(tags)))
		}

		filterSection = strings.Join(filterParts, " ")
	} else {
		filterSection = styles.DimStyle.Render("No active filters")
	}

	// Help text
	help := styles.DimStyle.Render("Press: / search • s state • c category • t tracker • g tag • x clear")

	// Calculate space usage
	filterLen := lipgloss.Width(filterSection)
	helpLen := lipgloss.Width(help)
	spacing := 3 // minimum spacing between sections

	// If both fit comfortably on one line, use horizontal layout
	if filterLen+helpLen+spacing <= availableWidth {
		paddingNeeded := availableWidth - filterLen - helpLen
		if paddingNeeded < spacing {
			paddingNeeded = spacing
		}

		padding := strings.Repeat(" ", paddingNeeded)
		return filterSection + padding + help
	}

	// If we have enough width, try to balance the space
	if availableWidth > 100 {
		// Calculate how much space to give each section
		totalContent := filterLen + helpLen
		extraSpace := availableWidth - totalContent - spacing

		if extraSpace > 20 { // If we have significant extra space
			// Add padding in the middle
			midPadding := spacing + (extraSpace / 2)
			padding := strings.Repeat(" ", midPadding)
			return filterSection + padding + help
		}
	}

	// Fallback to simple layout for narrow screens or when content is too long
	return filterSection + "  " + help
}

// renderSimpleNormalMode provides a fallback layout when width is not available
func (f *FilterPanel) renderSimpleNormalMode() string {
	var parts []string

	// Active filters
	if !f.filter.IsEmpty() {
		parts = append(parts, styles.TitleStyle.Render("Active Filters:"))

		if f.filter.Search != "" {
			parts = append(parts, fmt.Sprintf("  Search: %s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(f.filter.Search)))
		}
		if len(f.filter.States) > 0 {
			parts = append(parts, fmt.Sprintf("  States: %s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(strings.Join(f.filter.States, ", "))))
		}
		if f.filter.Category != "" {
			parts = append(parts, fmt.Sprintf("  Category: %s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(f.filter.Category)))
		}
		if len(f.filter.Trackers) > 0 {
			parts = append(parts, fmt.Sprintf("  Trackers: %s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(strings.Join(f.filter.Trackers, ", "))))
		}
		if len(f.filter.Tags) > 0 {
			parts = append(parts, fmt.Sprintf("  Tags: %s", lipgloss.NewStyle().Foreground(styles.AccentColor).Render(strings.Join(f.filter.Tags, ", "))))
		}
	} else {
		parts = append(parts, styles.DimStyle.Render("No active filters"))
	}

	// Help text
	help := styles.DimStyle.Render("Press: / search • s state • c category • t tracker • g tag • x clear")
	parts = append(parts, help)

	return strings.Join(parts, " ")
}

// renderSearchMode renders the search input
func (f *FilterPanel) renderSearchMode() string {
	title := styles.TitleStyle.Render("Search:")
	input := f.searchInput.View()
	help := styles.DimStyle.Render("Enter save • Esc cancel")

	return fmt.Sprintf("%s %s  %s", title, input, help)
}

// renderListMode renders a selection list
func (f *FilterPanel) renderListMode(title string, options []string, selected []string) string {
	if len(options) == 0 {
		return styles.DimStyle.Render(fmt.Sprintf("No %s options available", strings.ToLower(title)))
	}

	var parts []string
	parts = append(parts, styles.TitleStyle.Render(fmt.Sprintf("Select %s:", title)))

	// Calculate how many options we can show based on terminal width
	maxVisible := f.calculateMaxVisibleOptions()

	// Show options with selection state
	visibleOptions := options
	if len(options) > maxVisible {
		// Show cursor and surrounding options
		start := f.cursor - (maxVisible / 2)
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > len(options) {
			end = len(options)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
		visibleOptions = options[start:end]

		for i, opt := range visibleOptions {
			actualIndex := start + i
			isSelected := contains(selected, opt)
			isCursor := actualIndex == f.cursor

			line := f.renderOption(opt, isSelected, isCursor)
			parts = append(parts, line)
		}
	} else {
		for i, opt := range visibleOptions {
			isSelected := contains(selected, opt)
			isCursor := i == f.cursor

			line := f.renderOption(opt, isSelected, isCursor)
			parts = append(parts, line)
		}
	}

	help := styles.DimStyle.Render("↑↓ navigate • Space toggle • a all • n none • Enter save • Esc cancel")
	parts = append(parts, help)

	return strings.Join(parts, " ")
}

// calculateMaxVisibleOptions determines how many filter options to show based on terminal width
func (f *FilterPanel) calculateMaxVisibleOptions() int {
	// Default minimum and maximum
	minVisible := 3
	maxVisible := 6

	// If width is not set, use minimum
	if f.width == 0 {
		return minVisible
	}

	// Use a simpler, more generous calculation based on width ranges
	// This gives users more visible options on wider terminals

	if f.width < 120 {
		return minVisible // 3 options on narrow screens
	} else if f.width < 160 {
		return 4 // 4 options on medium screens
	} else if f.width < 200 {
		return 5 // 5 options on wide screens
	} else {
		return maxVisible // 6 options on very wide screens
	}
}

// renderOption renders a single option in list mode
func (f *FilterPanel) renderOption(option string, selected, cursor bool) string {
	checkbox := "[ ]"
	if selected {
		checkbox = "[✓]"
	}

	text := fmt.Sprintf("%s %s", checkbox, option)

	if cursor {
		return styles.SelectedRowStyle.Render(text)
	}
	return text
}

// Movement and selection methods
func (f *FilterPanel) moveCursorUp() {
	if f.cursor > 0 {
		f.cursor--
	}
}

func (f *FilterPanel) moveCursorDown() {
	maxCursor := 0
	switch f.mode {
	case FilterModeState:
		maxCursor = len(f.availableStates) - 1
	case FilterModeCategory:
		maxCursor = len(f.availableCategories) - 1
	case FilterModeTracker:
		maxCursor = len(f.availableTrackers) - 1
	case FilterModeTag:
		maxCursor = len(f.availableTags) - 1
	}

	if f.cursor < maxCursor {
		f.cursor++
	}
}

func (f *FilterPanel) toggleSelection() {
	switch f.mode {
	case FilterModeState:
		if f.cursor < len(f.availableStates) {
			state := f.availableStates[f.cursor]
			if contains(f.filter.States, state) {
				f.filter.States = remove(f.filter.States, state)
			} else {
				f.filter.States = append(f.filter.States, state)
			}
		}
	case FilterModeCategory:
		if f.cursor < len(f.availableCategories) {
			category := f.availableCategories[f.cursor]
			// Toggle category: if already selected, clear it; otherwise set it
			if f.filter.Category == category {
				f.filter.Category = ""
			} else {
				f.filter.Category = category
			}
		}
	case FilterModeTracker:
		if f.cursor < len(f.availableTrackers) {
			tracker := f.availableTrackers[f.cursor]
			if contains(f.filter.Trackers, tracker) {
				f.filter.Trackers = remove(f.filter.Trackers, tracker)
			} else {
				f.filter.Trackers = append(f.filter.Trackers, tracker)
			}
		}
	case FilterModeTag:
		if f.cursor < len(f.availableTags) {
			tag := f.availableTags[f.cursor]
			if contains(f.filter.Tags, tag) {
				f.filter.Tags = remove(f.filter.Tags, tag)
			} else {
				f.filter.Tags = append(f.filter.Tags, tag)
			}
		}
	}
}

func (f *FilterPanel) selectAll() {
	switch f.mode {
	case FilterModeState:
		f.filter.States = append([]string{}, f.availableStates...)
	case FilterModeCategory:
		// Categories are single-select, so "select all" doesn't apply
		// Do nothing
	case FilterModeTracker:
		f.filter.Trackers = append([]string{}, f.availableTrackers...)
	case FilterModeTag:
		f.filter.Tags = append([]string{}, f.availableTags...)
	}
}

func (f *FilterPanel) selectNone() {
	switch f.mode {
	case FilterModeState:
		f.filter.States = []string{}
	case FilterModeCategory:
		f.filter.Category = ""
	case FilterModeTracker:
		f.filter.Trackers = []string{}
	case FilterModeTag:
		f.filter.Tags = []string{}
	}
}

func (f *FilterPanel) clearFilters() {
	f.filter = filter.Filter{}
	f.searchInput.SetValue("")
}

// SetDimensions updates the component dimensions
func (f *FilterPanel) SetDimensions(width, height int) {
	f.width = width
	f.height = height
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func remove(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

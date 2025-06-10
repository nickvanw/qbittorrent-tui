package components

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/styles"
)

// SortConfig represents the current sort configuration
type SortConfig struct {
	Column    string    // Column key to sort by
	Direction SortDir   // Sort direction
	Secondary string    // Secondary sort column (for ties)
}

// SortDir represents sort direction
type SortDir int

const (
	SortAsc SortDir = iota
	SortDesc
)

// TorrentList is the torrent list component
type TorrentList struct {
	torrents       []api.Torrent
	cursor         int
	offset         int
	height         int
	width          int
	selectedHash   string
	showProgress   bool
	columns        []Column    // Calculated columns for current width
	sortConfig     SortConfig  // Current sort configuration
	visibleColumns []string    // List of visible column keys
	showConfig     bool        // Whether to show column config overlay
}

// ColumnConfig represents a responsive table column
type ColumnConfig struct {
	Key      string  // Unique identifier for the column
	Title    string  // Display title
	MinWidth int     // Minimum column width
	MaxWidth int     // Maximum column width (0 = unlimited)
	FlexGrow float64 // How much this column should grow (0-1)
	Priority int     // Hide order on narrow screens (1=highest priority)
}

// All available column definitions
var allColumns = []ColumnConfig{
	{Key: "name", Title: "Name", MinWidth: 20, MaxWidth: 0, FlexGrow: 0.6, Priority: 1},
	{Key: "size", Title: "Size", MinWidth: 8, MaxWidth: 12, FlexGrow: 0.0, Priority: 3},
	{Key: "progress", Title: "Progress", MinWidth: 10, MaxWidth: 13, FlexGrow: 0.0, Priority: 2}, // "Progress ↑" = 10 chars
	{Key: "status", Title: "Status", MinWidth: 8, MaxWidth: 15, FlexGrow: 0.1, Priority: 2},   // "Status ↑" = 8 chars
	{Key: "down", Title: "Down", MinWidth: 6, MaxWidth: 15, FlexGrow: 0.1, Priority: 3},       // "Down ↑" = 6 chars
	{Key: "up", Title: "Up", MinWidth: 4, MaxWidth: 15, FlexGrow: 0.1, Priority: 4},           // "Up ↑" = 4 chars
	{Key: "seeds", Title: "Seeds", MinWidth: 7, MaxWidth: 12, FlexGrow: 0.05, Priority: 4},    // "Seeds ↑" = 7 chars
	{Key: "peers", Title: "Peers", MinWidth: 7, MaxWidth: 12, FlexGrow: 0.05, Priority: 5},    // "Peers ↑" = 7 chars
	{Key: "ratio", Title: "Ratio", MinWidth: 7, MaxWidth: 10, FlexGrow: 0.0, Priority: 5},     // "Ratio ↑" = 7 chars
	{Key: "eta", Title: "ETA", MinWidth: 5, MaxWidth: 15, FlexGrow: 0.05, Priority: 6},        // "ETA ↑" = 5 chars
	{Key: "added_on", Title: "Added", MinWidth: 7, MaxWidth: 20, FlexGrow: 0.05, Priority: 7}, // "Added ↑" = 7 chars
	{Key: "category", Title: "Category", MinWidth: 10, MaxWidth: 20, FlexGrow: 0.1, Priority: 8}, // "Category ↑" = 10 chars
	{Key: "tags", Title: "Tags", MinWidth: 6, MaxWidth: 25, FlexGrow: 0.1, Priority: 9},       // "Tags ↑" = 6 chars
	{Key: "tracker", Title: "Tracker", MinWidth: 9, MaxWidth: 25, FlexGrow: 0.1, Priority: 10}, // "Tracker ↑" = 9 chars
}

// Default visible columns
var defaultVisibleColumns = []string{
	"name", "size", "progress", "status", "down", "up", "seeds", "peers", "ratio",
}

// Column represents a rendered column with calculated width
type Column struct {
	Config ColumnConfig
	Width  int
}

// NewTorrentList creates a new torrent list component
func NewTorrentList() *TorrentList {
	return &TorrentList{
		torrents:       []api.Torrent{},
		showProgress:   true,
		visibleColumns: append([]string{}, defaultVisibleColumns...), // Copy default columns
		sortConfig: SortConfig{
			Column:    "name",    // Default sort by name
			Direction: SortAsc,   // Ascending
			Secondary: "size",    // Secondary sort by size
		},
	}
}

// SetTorrents updates the torrent list and applies sorting
func (t *TorrentList) SetTorrents(torrents []api.Torrent) {
	t.torrents = torrents
	t.applySorting()
	// Keep cursor in bounds
	if t.cursor >= len(t.torrents) {
		t.cursor = len(t.torrents) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	// Update selected hash
	if t.cursor < len(t.torrents) {
		t.selectedHash = t.torrents[t.cursor].Hash
	}
}

// Update handles messages
func (t *TorrentList) Update(msg tea.Msg) (*TorrentList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle column configuration mode
		if t.showConfig {
			switch msg.String() {
			case "C", "c", "esc":
				t.showConfig = false
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Toggle column visibility
				num := int(msg.String()[0] - '1')
				t.ToggleColumn(num)
			case "0":
				// Handle 10th column
				t.ToggleColumn(9)
			case "q", "Q":
				// Handle 11th column
				t.ToggleColumn(10)
			case "w", "W":
				// Handle 12th column
				t.ToggleColumn(11)
			case "e", "E":
				// Handle 13th column
				t.ToggleColumn(12)
			case "r", "R":
				// Handle 14th column
				t.ToggleColumn(13)
			}
			return t, nil
		}
		
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			t.moveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			t.moveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			t.moveToTop()
		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			t.moveToBottom()
		
		// Column configuration
		case key.Matches(msg, key.NewBinding(key.WithKeys("C"))):
			t.showConfig = !t.showConfig
		
		// Dynamic sorting shortcuts based on visible columns
		default:
			// Check if it's a number key for sorting
			if len(msg.String()) == 1 {
				char := msg.String()[0]
				
				// Number keys 1-9 for sorting visible columns
				if char >= '1' && char <= '9' {
					index := int(char - '1')
					if index < len(t.columns) {
						t.setSortColumn(t.columns[index].Config.Key, false)
					}
				}
				
				// Shift+number (!@#$%^&*() for reverse sorting
				shiftMap := map[byte]int{
					'!': 0, '@': 1, '#': 2, '$': 3, '%': 4,
					'^': 5, '&': 6, '*': 7, '(': 8, ')': 9,
				}
				if idx, ok := shiftMap[char]; ok && idx < len(t.columns) {
					t.setSortColumn(t.columns[idx].Config.Key, true)
				}
			}
		}
	}
	return t, nil
}

// View renders the torrent list
func (t *TorrentList) View() string {
	if len(t.torrents) == 0 {
		return styles.DimStyle.Render("No torrents")
	}

	var s strings.Builder

	// Render header
	s.WriteString(t.renderHeader())
	s.WriteString("\n")

	// Calculate visible torrents
	visibleHeight := t.height - 1 // Account for header (header already includes border)
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Adjust offset to keep cursor visible
	if t.cursor < t.offset {
		t.offset = t.cursor
	} else if t.cursor >= t.offset+visibleHeight {
		t.offset = t.cursor - visibleHeight + 1
	}

	// Render visible torrents
	end := t.offset + visibleHeight
	if end > len(t.torrents) {
		end = len(t.torrents)
	}

	for i := t.offset; i < end; i++ {
		s.WriteString(t.renderTorrent(i))
		if i < end-1 {
			s.WriteString("\n")
		}
	}

	listView := s.String()

	// Show column config overlay if enabled
	if t.showConfig {
		return t.renderWithColumnConfig(listView)
	}

	return listView
}

// renderHeader renders the table header with sort indicators
func (t *TorrentList) renderHeader() string {
	var headers []string
	for _, col := range t.columns {
		title := col.Config.Title
		
		// Check if this column is currently sorted
		if col.Config.Key == t.sortConfig.Column {
			var indicator string
			if t.sortConfig.Direction == SortAsc {
				indicator = " ↑"
			} else {
				indicator = " ↓"
			}
			
			// Calculate full title with indicator
			fullTitle := title + indicator
			
			// If it fits, use it as-is (use rune count for proper Unicode support)
			if utf8.RuneCountInString(fullTitle) <= col.Width {
				header := lipgloss.NewStyle().Width(col.Width).Render(fullTitle)
				headers = append(headers, header)
			} else {
				// Need to truncate - reserve space for indicator
				indicatorWidth := utf8.RuneCountInString(indicator)
				maxTitleWidth := col.Width - indicatorWidth
				if maxTitleWidth < 1 {
					// Column too narrow, just show indicator
					header := lipgloss.NewStyle().Width(col.Width).Render(indicator)
					headers = append(headers, header)
				} else {
					// Truncate title and add indicator
					truncatedTitle := styles.TruncateString(title, maxTitleWidth)
					fullTitle := truncatedTitle + indicator
					header := lipgloss.NewStyle().Width(col.Width).Render(fullTitle)
					headers = append(headers, header)
				}
			}
		} else {
			// No sort indicator needed
			truncatedTitle := styles.TruncateString(title, col.Width)
			header := lipgloss.NewStyle().Width(col.Width).Render(truncatedTitle)
			headers = append(headers, header)
		}
	}
	return styles.HeaderStyle.Render(strings.Join(headers, " "))
}

// renderTorrent renders a single torrent row
func (t *TorrentList) renderTorrent(index int) string {
	torrent := t.torrents[index]
	isSelected := index == t.cursor

	var cells []string

	for _, col := range t.columns {
		var content string
		var style lipgloss.Style

		switch col.Config.Key {
		case "name":
			content = styles.TruncateString(torrent.Name, col.Width)
			style = lipgloss.NewStyle()
		case "size":
			content = styles.FormatBytes(torrent.Size)
			style = lipgloss.NewStyle()
		case "progress":
			content = fmt.Sprintf("%.1f%%", torrent.Progress*100)
			style = lipgloss.NewStyle()
		case "status":
			content = t.getStatusDisplay(torrent.State)
			style = styles.GetStateStyle(torrent.State)
		case "seeds":
			content = fmt.Sprintf("%d/%d", torrent.NumSeeds, torrent.NumComplete)
			style = lipgloss.NewStyle()
		case "peers":
			content = fmt.Sprintf("%d/%d", torrent.NumLeeches, torrent.NumIncomplete)
			style = lipgloss.NewStyle()
		case "down":
			content = styles.FormatSpeed(torrent.DlSpeed)
			style = lipgloss.NewStyle()
		case "up":
			content = styles.FormatSpeed(torrent.UpSpeed)
			style = lipgloss.NewStyle()
		case "ratio":
			content = fmt.Sprintf("%.2f", torrent.Ratio)
			style = lipgloss.NewStyle()
		case "eta":
			content = styles.FormatDuration(torrent.ETA)
			style = lipgloss.NewStyle()
		case "added_on":
			content = styles.FormatTime(torrent.AddedOn)
			style = lipgloss.NewStyle()
		case "category":
			content = styles.TruncateString(torrent.Category, col.Width)
			style = lipgloss.NewStyle()
		case "tags":
			content = styles.TruncateString(torrent.Tags, col.Width)
			style = lipgloss.NewStyle()
		case "tracker":
			content = styles.TruncateString(torrent.Tracker, col.Width)
			style = lipgloss.NewStyle()
		default:
			content = ""
			style = lipgloss.NewStyle()
		}

		cell := style.Width(col.Width).Render(content)
		cells = append(cells, cell)
	}

	row := strings.Join(cells, " ")

	if isSelected {
		return styles.SelectedRowStyle.Render(row)
	}
	return row
}

// getStatusDisplay returns a human-readable status
func (t *TorrentList) getStatusDisplay(state string) string {
	switch state {
	case "downloading":
		return "Downloading"
	case "metaDL":
		return "Metadata"
	case "forcedDL":
		return "Force DL"
	case "allocating":
		return "Allocating"
	case "uploading":
		return "Seeding"
	case "forcedUP":
		return "Force Seed"
	case "stalledUP":
		return "Stalled"
	case "pausedDL":
		return "Paused DL"
	case "pausedUP":
		return "Paused UP"
	case "queuedDL":
		return "Queued DL"
	case "queuedUP":
		return "Queued UP"
	case "error":
		return "Error"
	case "missingFiles":
		return "Missing"
	default:
		return state
	}
}

// Movement methods
func (t *TorrentList) moveUp() {
	if t.cursor > 0 {
		t.cursor--
		if t.cursor < len(t.torrents) {
			t.selectedHash = t.torrents[t.cursor].Hash
		}
	}
}

func (t *TorrentList) moveDown() {
	if t.cursor < len(t.torrents)-1 {
		t.cursor++
		t.selectedHash = t.torrents[t.cursor].Hash
	}
}

func (t *TorrentList) moveToTop() {
	t.cursor = 0
	t.offset = 0
	if len(t.torrents) > 0 {
		t.selectedHash = t.torrents[0].Hash
	}
}

func (t *TorrentList) moveToBottom() {
	if len(t.torrents) > 0 {
		t.cursor = len(t.torrents) - 1
		t.selectedHash = t.torrents[t.cursor].Hash
	}
}

// SetDimensions updates the component dimensions and recalculates columns
func (t *TorrentList) SetDimensions(width, height int) {
	t.width = width
	t.height = height
	t.columns = t.calculateColumnWidths(width)
}

// calculateColumnWidths determines column widths based on available space
func (t *TorrentList) calculateColumnWidths(availableWidth int) []Column {
	// Build list of configs for visible columns
	var selectedConfigs []ColumnConfig
	for _, key := range t.visibleColumns {
		for _, config := range allColumns {
			if config.Key == key {
				selectedConfigs = append(selectedConfigs, config)
				break
			}
		}
	}
	
	// Account for spacing between columns (1 space each)
	spacing := len(selectedConfigs) - 1
	if spacing < 0 {
		spacing = 0
	}
	usableWidth := availableWidth - spacing
	
	// First pass: allocate minimum widths and check what fits
	var visibleConfigs []ColumnConfig
	totalMinWidth := 0
	
	// Sort by priority (1 = highest priority, shown first)
	sortedConfigs := make([]ColumnConfig, len(selectedConfigs))
	copy(sortedConfigs, selectedConfigs)
	
	// Simple sort by priority
	for i := 0; i < len(sortedConfigs); i++ {
		for j := i + 1; j < len(sortedConfigs); j++ {
			if sortedConfigs[j].Priority < sortedConfigs[i].Priority {
				sortedConfigs[i], sortedConfigs[j] = sortedConfigs[j], sortedConfigs[i]
			}
		}
	}
	
	// Determine which columns fit
	for _, config := range sortedConfigs {
		if totalMinWidth+config.MinWidth <= usableWidth {
			visibleConfigs = append(visibleConfigs, config)
			totalMinWidth += config.MinWidth
		}
	}
	
	// Second pass: distribute remaining space
	remainingWidth := usableWidth - totalMinWidth
	var columns []Column
	
	// Calculate total flex grow weight
	totalFlexGrow := 0.0
	for _, config := range visibleConfigs {
		totalFlexGrow += config.FlexGrow
	}
	
	// Multi-pass allocation to handle max width constraints
	allocatedWidths := make([]int, len(visibleConfigs))
	remainingWidthToDistribute := remainingWidth
	
	// First pass: allocate based on flex grow, respecting max widths
	for i, config := range visibleConfigs {
		allocatedWidths[i] = config.MinWidth
		
		if totalFlexGrow > 0 && remainingWidth > 0 {
			flexWidth := int(float64(remainingWidth) * (config.FlexGrow / totalFlexGrow))
			targetWidth := config.MinWidth + flexWidth
			
			// Respect max width constraints
			if config.MaxWidth > 0 && targetWidth > config.MaxWidth {
				allocatedWidths[i] = config.MaxWidth
				// Track how much space we couldn't use
				remainingWidthToDistribute -= (config.MaxWidth - config.MinWidth)
			} else {
				allocatedWidths[i] = targetWidth
				remainingWidthToDistribute -= flexWidth
			}
		}
	}
	
	// Second pass: distribute any leftover space to the name column (unlimited growth)
	if remainingWidthToDistribute > 0 {
		for i, config := range visibleConfigs {
			if config.Key == "name" && config.MaxWidth == 0 {
				allocatedWidths[i] += remainingWidthToDistribute
				break
			}
		}
	}
	
	// Create final columns
	for i, config := range visibleConfigs {
		columns = append(columns, Column{
			Config: config,
			Width:  allocatedWidths[i],
		})
	}
	
	return columns
}

// GetSelectedHash returns the currently selected torrent hash
func (t *TorrentList) GetSelectedHash() string {
	return t.selectedHash
}

// GetColumns returns the current column configuration (for testing)
func (t *TorrentList) GetColumns() []Column {
	return t.columns
}

// setSortColumn sets the sort column and direction
func (t *TorrentList) setSortColumn(column string, reverse bool) {
	if t.sortConfig.Column == column {
		// Toggle direction if same column
		if t.sortConfig.Direction == SortAsc {
			t.sortConfig.Direction = SortDesc
		} else {
			t.sortConfig.Direction = SortAsc
		}
	} else {
		// New column, start with ascending unless reverse requested
		t.sortConfig.Column = column
		if reverse {
			t.sortConfig.Direction = SortDesc
		} else {
			t.sortConfig.Direction = SortAsc
		}
	}
	
	t.applySorting()
}

// applySorting sorts the torrents based on current sort configuration
func (t *TorrentList) applySorting() {
	if len(t.torrents) <= 1 {
		return
	}
	
	sort.Slice(t.torrents, func(i, j int) bool {
		return t.compareTorrents(t.torrents[i], t.torrents[j])
	})
}

// compareTorrents compares two torrents based on current sort configuration
func (t *TorrentList) compareTorrents(a, b api.Torrent) bool {
	result := t.compareByColumn(a, b, t.sortConfig.Column)
	
	// If primary comparison is equal, use secondary sort
	if result == 0 && t.sortConfig.Secondary != "" && t.sortConfig.Secondary != t.sortConfig.Column {
		result = t.compareByColumn(a, b, t.sortConfig.Secondary)
	}
	
	// If still equal, fall back to name for consistency
	if result == 0 {
		result = strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	}
	
	// Apply sort direction
	if t.sortConfig.Direction == SortDesc {
		return result > 0
	}
	return result < 0
}

// compareByColumn compares two torrents by a specific column
func (t *TorrentList) compareByColumn(a, b api.Torrent, column string) int {
	switch column {
	case "name":
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	case "size":
		if a.Size < b.Size {
			return -1
		} else if a.Size > b.Size {
			return 1
		}
		return 0
	case "progress":
		if a.Progress < b.Progress {
			return -1
		} else if a.Progress > b.Progress {
			return 1
		}
		return 0
	case "status":
		return strings.Compare(a.State, b.State)
	case "down":
		if a.DlSpeed < b.DlSpeed {
			return -1
		} else if a.DlSpeed > b.DlSpeed {
			return 1
		}
		return 0
	case "up":
		if a.UpSpeed < b.UpSpeed {
			return -1
		} else if a.UpSpeed > b.UpSpeed {
			return 1
		}
		return 0
	case "ratio":
		if a.Ratio < b.Ratio {
			return -1
		} else if a.Ratio > b.Ratio {
			return 1
		}
		return 0
	case "seeds":
		if a.NumSeeds < b.NumSeeds {
			return -1
		} else if a.NumSeeds > b.NumSeeds {
			return 1
		}
		return 0
	case "peers":
		if a.NumLeeches < b.NumLeeches {
			return -1
		} else if a.NumLeeches > b.NumLeeches {
			return 1
		}
		return 0
	case "added_on":
		if a.AddedOn < b.AddedOn {
			return -1
		} else if a.AddedOn > b.AddedOn {
			return 1
		}
		return 0
	case "eta":
		if a.ETA < b.ETA {
			return -1
		} else if a.ETA > b.ETA {
			return 1
		}
		return 0
	case "category":
		return strings.Compare(strings.ToLower(a.Category), strings.ToLower(b.Category))
	case "tags":
		return strings.Compare(strings.ToLower(a.Tags), strings.ToLower(b.Tags))
	case "tracker":
		return strings.Compare(strings.ToLower(a.Tracker), strings.ToLower(b.Tracker))
	default:
		// Unknown column, fall back to name
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	}
}

// GetSortConfig returns the current sort configuration (for persistence)
func (t *TorrentList) GetSortConfig() SortConfig {
	return t.sortConfig
}

// SetSortConfig sets the sort configuration (for loading saved preferences)
func (t *TorrentList) SetSortConfig(config SortConfig) {
	t.sortConfig = config
	t.applySorting()
}

// renderWithColumnConfig renders the torrent list with column configuration overlay
func (t *TorrentList) renderWithColumnConfig(listView string) string {
	// Calculate layout - use most of the terminal space
	contentWidth := t.width - 4  // Leave some margin
	if contentWidth > 100 {
		contentWidth = 100  // Cap max width for readability
	}
	contentHeight := t.height - 4
	
	// Build the overlay content
	var content strings.Builder
	
	// Title
	title := "Column Configuration"
	titleStyle := styles.TitleStyle.Bold(true).Underline(true)
	content.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, titleStyle.Render(title)))
	content.WriteString("\n\n")
	
	// Instructions
	content.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, 
		styles.DimStyle.Render("Toggle column visibility: 1-9, 0, q, w, e, r")))
	content.WriteString("\n\n")
	
	// Two-column layout for the column list
	halfWidth := (contentWidth - 4) / 2
	var leftCol, rightCol strings.Builder
	
	halfPoint := (len(allColumns) + 1) / 2
	
	// Left column
	for i := 0; i < halfPoint && i < len(allColumns); i++ {
		config := allColumns[i]
		isVisible := false
		for _, key := range t.visibleColumns {
			if key == config.Key {
				isVisible = true
				break
			}
		}
		
		checkbox := "[ ]"
		checkStyle := styles.DimStyle
		if isVisible {
			checkbox = "[✓]"
			checkStyle = styles.AccentStyle
		}
		
		// Format with proper spacing
		var num string
		if i < 9 {
			num = fmt.Sprintf("%d.", i+1)
		} else if i == 9 {
			num = "0."  // Use 0 for 10th item
		} else {
			// Use q, w, e, r for columns 11-14 to avoid conflicts
			extraKeys := []string{"q", "w", "e", "r"}
			if i-10 < len(extraKeys) {
				num = fmt.Sprintf("%s.", extraKeys[i-10])
			} else {
				num = fmt.Sprintf("%d.", i+1)
			}
		}
		line := fmt.Sprintf("%-3s %s %-15s", num, checkStyle.Render(checkbox), config.Title)
		
		if isVisible {
			line = styles.TextStyle.Render(line)
		} else {
			line = styles.DimStyle.Render(line)
		}
		leftCol.WriteString(line)
		leftCol.WriteString("\n")
	}
	
	// Right column
	for i := halfPoint; i < len(allColumns); i++ {
		config := allColumns[i]
		isVisible := false
		for _, key := range t.visibleColumns {
			if key == config.Key {
				isVisible = true
				break
			}
		}
		
		checkbox := "[ ]"
		checkStyle := styles.DimStyle
		if isVisible {
			checkbox = "[✓]"
			checkStyle = styles.AccentStyle
		}
		
		// Format with proper spacing - handle numbers > 9
		var num string
		if i < 9 {
			num = fmt.Sprintf("%d.", i+1)
		} else if i == 9 {
			num = "0."
		} else {
			// Use q, w, e, r for columns 11-14 to avoid conflicts
			extraKeys := []string{"q", "w", "e", "r"}
			if i-10 < len(extraKeys) {
				num = fmt.Sprintf("%s.", extraKeys[i-10])
			} else {
				num = fmt.Sprintf("%d.", i+1)
			}
		}
		line := fmt.Sprintf("%-3s %s %-15s", num, checkStyle.Render(checkbox), config.Title)
		
		if isVisible {
			line = styles.TextStyle.Render(line)
		} else {
			line = styles.DimStyle.Render(line)
		}
		rightCol.WriteString(line)
		rightCol.WriteString("\n")
	}
	
	// Combine columns
	leftColStr := leftCol.String()
	rightColStr := rightCol.String()
	leftLines := strings.Split(leftColStr, "\n")
	rightLines := strings.Split(rightColStr, "\n")
	
	for i := 0; i < len(leftLines) || i < len(rightLines); i++ {
		left := ""
		right := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(rightLines) {
			right = rightLines[i]
		}
		
		if left != "" || right != "" {
			content.WriteString(fmt.Sprintf("  %-*s    %s\n", halfWidth, left, right))
		}
	}
	
	// Footer
	content.WriteString("\n")
	content.WriteString(lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center,
		styles.DimStyle.Render("Press 'C' or 'Esc' to close")))
	
	// Create the box
	box := styles.FocusedPanelStyle.
		Width(contentWidth).
		Height(contentHeight).
		Render(content.String())
	
	// Place the box in the center
	finalView := lipgloss.Place(t.width, t.height, 
		lipgloss.Center, lipgloss.Top,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a1a")))
	
	return finalView
}

// ToggleColumn toggles visibility of a column by index (0-based)
func (t *TorrentList) ToggleColumn(index int) {
	if index < 0 || index >= len(allColumns) {
		return
	}
	
	columnKey := allColumns[index].Key
	
	// Check if column is currently visible
	found := -1
	for i, key := range t.visibleColumns {
		if key == columnKey {
			found = i
			break
		}
	}
	
	if found >= 0 {
		// Remove column
		t.visibleColumns = append(t.visibleColumns[:found], t.visibleColumns[found+1:]...)
	} else {
		// Add column
		t.visibleColumns = append(t.visibleColumns, columnKey)
	}
	
	// Recalculate column widths
	t.columns = t.calculateColumnWidths(t.width)
}

// GetVisibleColumns returns the list of visible column keys
func (t *TorrentList) GetVisibleColumns() []string {
	return append([]string{}, t.visibleColumns...)
}

// SetVisibleColumns sets the list of visible columns
func (t *TorrentList) SetVisibleColumns(columns []string) {
	t.visibleColumns = append([]string{}, columns...)
	t.columns = t.calculateColumnWidths(t.width)
}

// IsInConfigMode returns whether the column configuration overlay is shown
func (t *TorrentList) IsInConfigMode() bool {
	return t.showConfig
}

// GetSortableColumns returns a map of column positions to their titles for help display
func (t *TorrentList) GetSortableColumns() map[int]string {
	result := make(map[int]string)
	for i, col := range t.columns {
		if i < 9 { // Only first 9 columns have keyboard shortcuts
			result[i+1] = col.Config.Title
		}
	}
	return result
}

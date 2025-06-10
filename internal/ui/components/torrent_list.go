package components

import (
	"fmt"
	"sort"
	"strings"

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
	torrents     []api.Torrent
	cursor       int
	offset       int
	height       int
	width        int
	selectedHash string
	showProgress bool
	columns      []Column    // Calculated columns for current width
	sortConfig   SortConfig  // Current sort configuration
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
	{Key: "name", Title: "Name", MinWidth: 20, MaxWidth: 0, FlexGrow: 0.4, Priority: 1},
	{Key: "size", Title: "Size", MinWidth: 8, MaxWidth: 12, FlexGrow: 0.0, Priority: 3},
	{Key: "progress", Title: "Progress", MinWidth: 8, MaxWidth: 10, FlexGrow: 0.0, Priority: 2},
	{Key: "status", Title: "Status", MinWidth: 8, MaxWidth: 12, FlexGrow: 0.1, Priority: 2},
	{Key: "down", Title: "Down", MinWidth: 8, MaxWidth: 12, FlexGrow: 0.1, Priority: 3},
	{Key: "up", Title: "Up", MinWidth: 8, MaxWidth: 12, FlexGrow: 0.1, Priority: 4},
	{Key: "seeds", Title: "Seeds", MinWidth: 6, MaxWidth: 10, FlexGrow: 0.0, Priority: 4},
	{Key: "peers", Title: "Peers", MinWidth: 6, MaxWidth: 10, FlexGrow: 0.0, Priority: 5},
	{Key: "ratio", Title: "Ratio", MinWidth: 6, MaxWidth: 8, FlexGrow: 0.0, Priority: 5},
	{Key: "eta", Title: "ETA", MinWidth: 8, MaxWidth: 12, FlexGrow: 0.0, Priority: 6},
	{Key: "added_on", Title: "Added", MinWidth: 10, MaxWidth: 16, FlexGrow: 0.0, Priority: 7},
	{Key: "category", Title: "Category", MinWidth: 8, MaxWidth: 15, FlexGrow: 0.0, Priority: 8},
	{Key: "tags", Title: "Tags", MinWidth: 8, MaxWidth: 20, FlexGrow: 0.0, Priority: 9},
	{Key: "tracker", Title: "Tracker", MinWidth: 10, MaxWidth: 20, FlexGrow: 0.0, Priority: 10},
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
		torrents:     []api.Torrent{},
		showProgress: true,
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
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			t.moveUp()
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			t.moveDown()
		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			t.moveToTop()
		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			t.moveToBottom()
		
		// Sorting shortcuts
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			t.setSortColumn("name", false)
		case key.Matches(msg, key.NewBinding(key.WithKeys("!"))):
			t.setSortColumn("name", true)
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			t.setSortColumn("size", false)
		case key.Matches(msg, key.NewBinding(key.WithKeys("@"))):
			t.setSortColumn("size", true)
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			t.setSortColumn("progress", false)
		case key.Matches(msg, key.NewBinding(key.WithKeys("#"))):
			t.setSortColumn("progress", true)
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			t.setSortColumn("status", false)
		case key.Matches(msg, key.NewBinding(key.WithKeys("$"))):
			t.setSortColumn("status", true)
		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			t.setSortColumn("down", false)
		case key.Matches(msg, key.NewBinding(key.WithKeys("%"))):
			t.setSortColumn("down", true)
		case key.Matches(msg, key.NewBinding(key.WithKeys("6"))):
			t.setSortColumn("up", false)
		case key.Matches(msg, key.NewBinding(key.WithKeys("^"))):
			t.setSortColumn("up", true)
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

	return s.String()
}

// renderHeader renders the table header with sort indicators
func (t *TorrentList) renderHeader() string {
	var headers []string
	for _, col := range t.columns {
		title := col.Config.Title
		
		// Add sort indicator if this column is currently sorted
		if col.Config.Key == t.sortConfig.Column {
			if t.sortConfig.Direction == SortAsc {
				title += " ↑"
			} else {
				title += " ↓"
			}
		}
		
		header := styles.TruncateString(title, col.Width)
		header = lipgloss.NewStyle().Width(col.Width).Render(header)
		headers = append(headers, header)
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
	// Account for spacing between columns (1 space each)
	spacing := len(defaultColumns) - 1
	usableWidth := availableWidth - spacing
	
	// First pass: allocate minimum widths and check what fits
	var visibleConfigs []ColumnConfig
	totalMinWidth := 0
	
	// Sort by priority (1 = highest priority, shown first)
	sortedConfigs := make([]ColumnConfig, len(defaultColumns))
	copy(sortedConfigs, defaultColumns)
	
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
	
	for _, config := range visibleConfigs {
		width := config.MinWidth
		
		// Distribute remaining width based on flex grow
		if totalFlexGrow > 0 && remainingWidth > 0 {
			flexWidth := int(float64(remainingWidth) * (config.FlexGrow / totalFlexGrow))
			width += flexWidth
			
			// Respect max width constraints
			if config.MaxWidth > 0 && width > config.MaxWidth {
				width = config.MaxWidth
			}
		}
		
		columns = append(columns, Column{
			Config: config,
			Width:  width,
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

package views

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/config"
	"github.com/nickvanw/qbittorrent-tui/internal/filter"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/components"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/styles"
)

// Removed FocusPane - using single-focus design with torrent list always focused

// AddMode represents the add torrent dialog mode
type AddMode int

const (
	ModeFile AddMode = iota
	ModeURL
)

// FileEntry represents a file or directory entry
type FileEntry struct {
	name     string
	isDir    bool
	size     int64
	fullPath string
}

// FileNavigator handles file browser functionality
type FileNavigator struct {
	currentPath   string
	allEntries    []FileEntry
	filtered      []FileEntry
	selectedIdx   int
	searchPattern string
	searchMode    bool
}

// URLInput handles URL input functionality
type URLInput struct {
	url    string
	cursor int
}

// AddTorrentDialog represents the add torrent dialog state
type AddTorrentDialog struct {
	mode     AddMode
	fileNav  *FileNavigator
	urlInput *URLInput
}

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeMain ViewMode = iota
	ViewModeDetails
)

// Message types
type (
	torrentDataMsg    []api.Torrent
	statsDataMsg      *api.GlobalStats
	categoriesDataMsg map[string]interface{}
	tagsDataMsg       []string
	errorMsg          error
	successMsg        string
	tickMsg           time.Time
	clearErrorMsg     struct{}
	clearSuccessMsg   struct{}
	delayedRefreshMsg struct{}
)

// MainView is the main application view
type MainView struct {
	config    *config.Config
	apiClient api.ClientInterface

	// UI components
	torrentList    *components.TorrentList
	statsPanel     *components.StatsPanel
	filterPanel    *components.FilterPanel
	torrentDetails *components.TorrentDetails
	help           help.Model

	// State
	torrents        []api.Torrent
	allTorrents     []api.Torrent // unfiltered torrents
	stats           *api.GlobalStats
	categories      map[string]interface{}
	tags            []string
	currentFilter   filter.Filter
	viewMode        ViewMode
	detailsViewHash string // Hash of torrent currently being viewed in details
	lastError       error
	lastSuccess     string
	isLoading       bool

	// Delete confirmation dialog state
	showDeleteDialog bool
	deleteTarget     *api.Torrent
	deleteWithFiles  bool

	// Add torrent dialog state
	showAddDialog bool
	addDialog     *AddTorrentDialog

	// Dimensions
	width  int
	height int

	// Keys
	keys KeyMap
}

// KeyMap defines all keyboard shortcuts
type KeyMap struct {
	Up   key.Binding
	Down key.Binding
	// Removed Left, Right, Tab - no longer cycling between panes
	Enter   key.Binding
	Escape  key.Binding
	Refresh key.Binding
	Filter  key.Binding
	Help    key.Binding
	Quit    key.Binding

	// Torrent control
	Pause   key.Binding
	Resume  key.Binding
	Delete  key.Binding
	Add     key.Binding
	Columns key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Escape},    // Navigation and Actions
		{k.Pause, k.Resume, k.Delete, k.Add}, // Torrent Control
		{k.Refresh, k.Filter, k.Columns},     // Features
		{k.Help, k.Quit},                     // General
	}
}

// DefaultKeyMap returns the default keyboard shortcuts
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		// Removed Left, Right, Tab - using single-focus design
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f", "/"),
			key.WithHelp("f", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),

		// Torrent control
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause"),
		),
		Resume: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "unpause/resume"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add torrent"),
		),
		Columns: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "configure columns"),
		),
	}
}

// NewMainView creates a new main view
func NewMainView(cfg *config.Config, client api.ClientInterface) *MainView {
	cwd, _ := os.Getwd() // Get current working directory, ignore error
	return &MainView{
		config:         cfg,
		apiClient:      client,
		torrentList:    components.NewTorrentList(),
		statsPanel:     components.NewStatsPanel(),
		filterPanel:    components.NewFilterPanel(),
		torrentDetails: components.NewTorrentDetails(client),
		help:           help.New(),
		keys:           DefaultKeyMap(),
		viewMode:       ViewModeMain,
		addDialog:      NewAddTorrentDialog(cwd),
	}
}

// Init initializes the view
func (m *MainView) Init() tea.Cmd {
	return tea.Batch(
		m.fetchAllData(),
		m.tickCmd(),
	)
}

// fetchAllData fetches all data from the API
func (m *MainView) fetchAllData() tea.Cmd {
	if m.apiClient == nil {
		return func() tea.Msg {
			return errorMsg(api.NewAuthError("not connected to qBittorrent", nil))
		}
	}

	return tea.Batch(
		m.fetchTorrents(),
		m.fetchStats(),
		m.fetchCategories(),
		m.fetchTags(),
	)
}

// fetchTorrents fetches torrent data
func (m *MainView) fetchTorrents() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()
		torrents, err := m.apiClient.GetTorrents(ctx)
		if err != nil {
			return errorMsg(err)
		}
		return torrentDataMsg(torrents)
	})
}

// fetchStats fetches global statistics
func (m *MainView) fetchStats() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()
		stats, err := m.apiClient.GetGlobalStats(ctx)
		if err != nil {
			return errorMsg(err)
		}
		return statsDataMsg(stats)
	})
}

// fetchCategories fetches categories
func (m *MainView) fetchCategories() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()
		categories, err := m.apiClient.GetCategories(ctx)
		if err != nil {
			return errorMsg(err)
		}
		return categoriesDataMsg(categories)
	})
}

// fetchTags fetches tags
func (m *MainView) fetchTags() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx := context.Background()
		tags, err := m.apiClient.GetTags(ctx)
		if err != nil {
			return errorMsg(err)
		}
		return tagsDataMsg(tags)
	})
}

// tickCmd creates a periodic tick for refreshing data
func (m *MainView) tickCmd() tea.Cmd {
	return tea.Tick(time.Duration(m.config.UI.RefreshInterval)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// clearErrorTimer creates a timer to clear errors after 5 seconds
func (m *MainView) clearErrorTimer() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

// clearSuccessTimer creates a timer to clear success messages after 3 seconds
func (m *MainView) clearSuccessTimer() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearSuccessMsg{}
	})
}

// delayedRefresh creates a delayed refresh to allow qBittorrent server to process changes
func (m *MainView) delayedRefresh() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return delayedRefreshMsg{}
	})
}

// Update handles messages
func (m *MainView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateDimensions()

	case torrentDataMsg:
		m.allTorrents = []api.Torrent(msg)
		m.applyFilter()
		m.isLoading = false
		// Don't auto-clear errors here - let timer handle it

		// Update torrent details if currently viewing a torrent
		if m.viewMode == ViewModeDetails && m.detailsViewHash != "" {
			// Search in ALL torrents (not just filtered ones)
			torrentFound := false
			for _, torrent := range m.allTorrents {
				if torrent.Hash == m.detailsViewHash {
					m.torrentDetails.UpdateTorrent(&torrent)
					torrentFound = true
					break
				}
			}

			// If torrent no longer exists (deleted), exit details mode
			if !torrentFound {
				m.viewMode = ViewModeMain
				m.detailsViewHash = ""
			}
		}

	case statsDataMsg:
		m.stats = (*api.GlobalStats)(msg)
		m.statsPanel.SetStats(m.stats)
		m.isLoading = false

	case categoriesDataMsg:
		m.categories = map[string]interface{}(msg)
		// Extract category names for filter panel
		var categoryNames []string
		for name := range m.categories {
			categoryNames = append(categoryNames, name)
		}
		// Sort categories alphabetically for stable display order
		sort.Strings(categoryNames)
		m.filterPanel.SetAvailableOptions(categoryNames, m.extractTrackerNames(), m.tags)

	case tagsDataMsg:
		m.tags = []string(msg)
		// Sort tags alphabetically for stable display order
		sort.Strings(m.tags)
		m.filterPanel.SetAvailableOptions(m.extractCategoryNames(), m.extractTrackerNames(), m.tags)

	case errorMsg:
		m.lastError = error(msg)
		m.isLoading = false
		// Start timer to clear error after 5 seconds
		cmds = append(cmds, m.clearErrorTimer())

	case clearErrorMsg:
		m.lastError = nil

	case successMsg:
		m.lastSuccess = string(msg)
		// Clear any existing errors when showing success
		m.lastError = nil
		// Start timer to clear success after 3 seconds
		cmds = append(cmds, m.clearSuccessTimer())
		// Trigger refresh to get updated state from server after successful mutation
		// Add small delay to allow qBittorrent server to process the change
		cmds = append(cmds, m.delayedRefresh())

	case clearSuccessMsg:
		m.lastSuccess = ""

	case delayedRefreshMsg:
		// Perform delayed refresh after mutation operations
		cmds = append(cmds, m.fetchAllData())

	case components.DetailsDataMsg:
		// Pass details data to torrent details component
		m.torrentDetails, cmd = m.torrentDetails.Update(msg)
		cmds = append(cmds, cmd)

	case tickMsg:
		// Refresh data periodically
		cmds = append(cmds, m.fetchAllData(), m.tickCmd())

		// Also refresh torrent details if in details view
		if m.viewMode == ViewModeDetails {
			m.torrentDetails, cmd = m.torrentDetails.Update(time.Time(msg))
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// Handle add torrent dialog first (highest priority)
		if m.showAddDialog {
			switch msg.String() {
			case "esc":
				// Close dialog
				m.showAddDialog = false
				return m, tea.Batch(cmds...)
			case "tab":
				// Switch between file and URL modes
				if m.addDialog.mode == ModeFile {
					m.addDialog.mode = ModeURL
				} else {
					m.addDialog.mode = ModeFile
				}
				return m, tea.Batch(cmds...)
			case "ctrl+c":
				// Still allow quit even in dialog
				return m, tea.Quit
			default:
				// Handle mode-specific keys - this catches ALL other keys
				if m.addDialog.mode == ModeFile {
					cmd = m.handleFileNavigatorKeys(msg.String())
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				} else {
					cmd = m.handleURLInputKeys(msg) // Pass the full KeyMsg for URL input
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
				// IMPORTANT: Return here to prevent any other key handling when dialog is open
				return m, tea.Batch(cmds...)
			}
		}

		// Handle delete confirmation dialog second priority
		if m.showDeleteDialog {
			switch msg.String() {
			case "y", "Y", "enter":
				// Confirm deletion
				cmd = m.confirmDeleteTorrent()
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case "n", "N", "esc":
				// Cancel deletion
				m.cancelDeleteTorrent()
				return m, tea.Batch(cmds...)
			case "f", "F":
				// Toggle delete files option
				m.deleteWithFiles = !m.deleteWithFiles
				return m, tea.Batch(cmds...)
			case "ctrl+c":
				// Still allow quit even in dialog
				return m, tea.Quit
			default:
				// Ignore other keys when dialog is open
				return m, tea.Batch(cmds...)
			}
		}

		// Don't clear errors immediately on keypress - let them persist until next action

		// If filter panel is in input mode, let it handle all keys except quit
		if m.filterPanel.IsInInputMode() {
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			default:
				// Pass all other keys to filter panel
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
				return m, tea.Batch(cmds...)
			}
		}

		// If filter panel is in interactive mode, let it handle keys first (except quit)
		if m.filterPanel.IsInInteractiveMode() {
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			default:
				// Pass to filter panel first, then fall through to global keys if not handled
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
				return m, tea.Batch(cmds...)
			}
		}

		// If torrent list is in column config mode, let it handle keys first (except quit)
		if m.torrentList.IsInConfigMode() {
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			default:
				// Pass to torrent list first for column configuration
				m.torrentList, cmd = m.torrentList.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Escape):
			if m.viewMode == ViewModeDetails {
				m.viewMode = ViewModeMain
				m.detailsViewHash = "" // Clear details view tracking
			} else if m.viewMode == ViewModeMain {
				// Let filter panel handle escape to exit search mode
				// Note: column config mode escape is handled earlier in the key hierarchy
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			}

		case key.Matches(msg, m.keys.Enter):
			if m.viewMode == ViewModeMain {
				// Show details for selected torrent
				// Note: filter panel interactive mode enter is handled earlier in the key hierarchy
				selectedHash := m.torrentList.GetSelectedHash()
				if selectedHash != "" {
					for _, torrent := range m.torrents {
						if torrent.Hash == selectedHash {
							cmd = m.torrentDetails.SetTorrent(&torrent)
							cmds = append(cmds, cmd)
							m.viewMode = ViewModeDetails
							m.detailsViewHash = selectedHash // Track which torrent we're viewing
							break
						}
					}
				}
			}

		// Removed tab/left/right focus cycling - using single-focus design

		case key.Matches(msg, m.keys.Refresh):
			m.isLoading = true
			cmds = append(cmds, m.fetchAllData())

		case key.Matches(msg, m.keys.Filter):
			if m.viewMode == ViewModeMain {
				// Start filtering
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
			}

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		// Torrent control actions - handle these BEFORE other key delegation
		case key.Matches(msg, m.keys.Pause):
			cmd = m.handlePauseTorrent()
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.Resume):
			cmd = m.handleResumeTorrent()
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.Delete):
			cmd = m.handleDeleteTorrent()
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.keys.Add):
			m.showAddDialog = true

		// Handle global filter keys BEFORE passing to components (to avoid conflicts)
		case msg.String() == "s": // State filter
			if m.viewMode == ViewModeMain && !m.filterPanel.IsInInteractiveMode() {
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			}

		case key.Matches(msg, m.keys.Columns):
			if m.viewMode == ViewModeMain {
				m.torrentList, cmd = m.torrentList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case msg.String() == "c": // Category filter
			if m.viewMode == ViewModeMain && !m.filterPanel.IsInInteractiveMode() {
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			}

		case msg.String() == "t": // Tracker filter
			if m.viewMode == ViewModeMain && !m.filterPanel.IsInInteractiveMode() {
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			}

		case msg.String() == "T": // Tag filter (Tags)
			if m.viewMode == ViewModeMain && !m.filterPanel.IsInInteractiveMode() {
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			}

		case msg.String() == "x": // Clear filters
			if m.viewMode == ViewModeMain {
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			}

		default:
			// Pass key events to components based on view mode
			if m.viewMode == ViewModeDetails {
				m.torrentDetails, cmd = m.torrentDetails.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				// Normal mode - pass navigation keys to torrent list (main focus)
				// Note: filter panel interactive mode and column config mode are handled earlier in the key hierarchy
				m.torrentList, cmd = m.torrentList.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the view
func (m *MainView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Check if we're in details mode
	if m.viewMode == ViewModeDetails {
		return m.renderDetailsView()
	}

	// Calculate layout dimensions
	helpHeight := strings.Count(m.help.View(m.keys), "\n") + 1
	contentHeight := m.height - helpHeight - 1

	// Stats panel height (fixed)
	statsHeight := 5

	// Filter panel height (fixed)
	filterHeight := 3

	// Torrent list gets remaining space
	torrentListHeight := contentHeight - statsHeight - filterHeight - 2 // 2 for borders

	// Create the layout
	var sections []string

	// Stats panel at the top
	statsView := m.renderStatsPanel(m.width, statsHeight)
	sections = append(sections, statsView)

	// Torrent list in the middle
	torrentView := m.renderTorrentList(m.width, torrentListHeight)
	sections = append(sections, torrentView)

	// Filter panel at the bottom
	filterView := m.renderFilterPanel(m.width, filterHeight)
	sections = append(sections, filterView)

	// Status line at the very bottom - priority: error (red) > success (green) > help
	var statusView string
	if m.lastError != nil {
		statusView = styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.lastError))
	} else if m.lastSuccess != "" {
		statusView = styles.SuccessStyle.Render(fmt.Sprintf("✓ %s", m.lastSuccess))
	} else {
		statusView = m.help.View(m.keys)
	}
	sections = append(sections, statusView)

	mainContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlay dialogs if active (add torrent has priority)
	if m.showAddDialog {
		dialog := m.renderAddDialog()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
	}

	if m.showDeleteDialog {
		dialog := m.renderDeleteDialog()
		// Simple overlay - place dialog in center
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
	}

	return mainContent
}

// renderStatsPanel renders the stats panel
func (m *MainView) renderStatsPanel(width, height int) string {
	style := styles.PanelStyle

	// Fixed dimensions: stats panel is always 5 lines tall
	m.statsPanel.SetDimensions(width-6, 2) // 5 total - 3 for borders/padding = 2 content

	content := m.statsPanel.View()
	return style.Width(width).Height(height).Render(content)
}

// renderTorrentList renders the torrent list (always focused)
func (m *MainView) renderTorrentList(width, height int) string {
	style := styles.FocusedPanelStyle // Always focused

	// Set dimensions for the torrent list component
	// Account for panel borders (2) and padding (4 horizontal, 1 vertical)
	m.torrentList.SetDimensions(width-6, height-3)

	content := m.torrentList.View()
	return style.Width(width).Height(height).Render(content)
}

// renderFilterPanel renders the filter panel
func (m *MainView) renderFilterPanel(width, height int) string {
	style := styles.PanelStyle

	// Fixed dimensions: filter panel is always 3 lines tall
	m.filterPanel.SetDimensions(width-6, 1) // 3 total - 2 for borders = 1 content line

	content := m.filterPanel.View()
	return style.Width(width).Height(height).Render(content)
}

// updateDimensions updates component dimensions based on window size
func (m *MainView) updateDimensions() {
	// Components will be updated with proper dimensions in the View method
	// Set dimensions for torrent details if in details mode
	if m.viewMode == ViewModeDetails {
		helpHeight := strings.Count(m.help.View(m.keys), "\n") + 1
		contentHeight := m.height - helpHeight - 1
		m.torrentDetails.SetSize(m.width-4, contentHeight-3)
	}
}

// Removed focus cycling methods - using single-focus design

// applyFilter applies the current filter to torrents
func (m *MainView) applyFilter() {
	m.torrents = m.currentFilter.Apply(m.allTorrents)
	m.torrentList.SetTorrents(m.torrents)
}

// handlePauseTorrent pauses the currently selected torrent
func (m *MainView) handlePauseTorrent() tea.Cmd {
	selectedHash := m.getSelectedTorrentHash()
	if selectedHash == "" {
		// Return an error message to help debug
		return func() tea.Msg {
			return errorMsg(fmt.Errorf("no torrent selected"))
		}
	}

	// Find the selected torrent to get its name for the success message
	var torrentName string
	for _, torrent := range m.torrents {
		if torrent.Hash == selectedHash {
			torrentName = torrent.Name
			break
		}
	}

	return func() tea.Msg {
		ctx := context.Background()
		err := m.apiClient.PauseTorrents(ctx, []string{selectedHash})
		if err != nil {
			return errorMsg(fmt.Errorf("failed to pause torrent: %w", err))
		}
		// Return success message
		return successMsg(fmt.Sprintf("paused: %s", styles.TruncateString(torrentName, 40)))
	}
}

// handleResumeTorrent resumes the currently selected torrent
func (m *MainView) handleResumeTorrent() tea.Cmd {
	selectedHash := m.getSelectedTorrentHash()
	if selectedHash == "" {
		return func() tea.Msg {
			return errorMsg(fmt.Errorf("no torrent selected"))
		}
	}

	// Find the selected torrent to get its name for the success message
	var torrentName string
	for _, torrent := range m.torrents {
		if torrent.Hash == selectedHash {
			torrentName = torrent.Name
			break
		}
	}

	return func() tea.Msg {
		ctx := context.Background()
		err := m.apiClient.ResumeTorrents(ctx, []string{selectedHash})
		if err != nil {
			return errorMsg(fmt.Errorf("failed to resume torrent: %w", err))
		}
		// Return success message
		return successMsg(fmt.Sprintf("resumed: %s", styles.TruncateString(torrentName, 40)))
	}
}

// handleDeleteTorrent shows confirmation dialog for deleting the currently selected torrent
func (m *MainView) handleDeleteTorrent() tea.Cmd {
	selectedHash := m.getSelectedTorrentHash()
	if selectedHash == "" {
		return func() tea.Msg {
			return errorMsg(fmt.Errorf("no torrent selected"))
		}
	}

	// Find the selected torrent object
	var selectedTorrent *api.Torrent
	for _, torrent := range m.torrents {
		if torrent.Hash == selectedHash {
			selectedTorrent = &torrent
			break
		}
	}

	if selectedTorrent == nil {
		return func() tea.Msg {
			return errorMsg(fmt.Errorf("selected torrent not found"))
		}
	}

	// Show confirmation dialog instead of immediate deletion
	m.showDeleteDialog = true
	m.deleteTarget = selectedTorrent
	m.deleteWithFiles = false // Default to not deleting files

	return nil // No command needed, just update UI state
}

// confirmDeleteTorrent performs the actual deletion after user confirmation
func (m *MainView) confirmDeleteTorrent() tea.Cmd {
	if m.deleteTarget == nil {
		return nil
	}

	hash := m.deleteTarget.Hash
	torrentName := m.deleteTarget.Name
	deleteFiles := m.deleteWithFiles

	// Close dialog
	m.showDeleteDialog = false
	m.deleteTarget = nil

	return func() tea.Msg {
		ctx := context.Background()
		err := m.apiClient.DeleteTorrents(ctx, []string{hash}, deleteFiles)
		if err != nil {
			return errorMsg(fmt.Errorf("failed to delete torrent: %w", err))
		}
		// Return success message
		successText := fmt.Sprintf("deleted: %s", styles.TruncateString(torrentName, 40))
		if deleteFiles {
			successText += " (with files)"
		}
		return successMsg(successText)
	}
}

// cancelDeleteTorrent cancels the delete operation
func (m *MainView) cancelDeleteTorrent() {
	m.showDeleteDialog = false
	m.deleteTarget = nil
	m.deleteWithFiles = false
}

// getSelectedTorrentHash returns the hash of the currently selected torrent
func (m *MainView) getSelectedTorrentHash() string {
	if m.viewMode == ViewModeDetails {
		// In details view, use the torrent list's selected hash
		return m.torrentList.GetSelectedHash()
	} else if m.viewMode == ViewModeMain {
		// In main view, use the torrent list's selected hash
		return m.torrentList.GetSelectedHash()
	}
	return ""
}

// extractCategoryNames extracts category names from the categories map
func (m *MainView) extractCategoryNames() []string {
	var names []string
	for name := range m.categories {
		names = append(names, name)
	}
	// Sort categories alphabetically for stable display order
	sort.Strings(names)
	return names
}

// extractTrackerNames extracts unique tracker names from torrents
func (m *MainView) extractTrackerNames() []string {
	if len(m.allTorrents) == 0 {
		return nil
	}
	return filter.ExtractUniqueTrackers(m.allTorrents)
}

// filterEqual compares two filters for equality
func filterEqual(a, b filter.Filter) bool {
	if a.Search != b.Search || a.Category != b.Category {
		return false
	}
	if len(a.States) != len(b.States) || len(a.Trackers) != len(b.Trackers) || len(a.Tags) != len(b.Tags) {
		return false
	}

	// Compare slices (order doesn't matter for our use case)
	for _, state := range a.States {
		found := false
		for _, bState := range b.States {
			if state == bState {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, tracker := range a.Trackers {
		found := false
		for _, bTracker := range b.Trackers {
			if tracker == bTracker {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, tag := range a.Tags {
		found := false
		for _, bTag := range b.Tags {
			if tag == bTag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// renderDetailsView renders the torrent details view
func (m *MainView) renderDetailsView() string {
	// Calculate available space for details
	helpHeight := strings.Count(m.help.View(m.keys), "\n") + 1
	contentHeight := m.height - helpHeight - 1

	// Set dimensions for the details component
	m.torrentDetails.SetSize(m.width-4, contentHeight-3) // Account for panel borders

	// Render the details in a panel
	content := m.torrentDetails.View()
	detailsPanel := styles.FocusedPanelStyle.Width(m.width).Height(contentHeight).Render(content)

	// Status line at the bottom - priority: error (red) > success (green) > help
	var statusView string
	if m.lastError != nil {
		statusView = styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.lastError))
	} else if m.lastSuccess != "" {
		statusView = styles.SuccessStyle.Render(fmt.Sprintf("✓ %s", m.lastSuccess))
	} else {
		statusView = m.help.View(m.keys)
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Left, detailsPanel, statusView)

	// Overlay dialogs if active (add torrent has priority)
	if m.showAddDialog {
		dialog := m.renderAddDialog()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
	}

	if m.showDeleteDialog {
		dialog := m.renderDeleteDialog()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
	}

	return mainContent
}

// renderDeleteDialog renders the delete confirmation dialog
func (m *MainView) renderDeleteDialog() string {
	if m.deleteTarget == nil {
		return ""
	}

	// Dialog box styling
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.AccentStyle.GetForeground()).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	// Title
	title := styles.AccentStyle.Render("Delete Torrent")

	// Torrent name (truncated if too long)
	torrentName := m.deleteTarget.Name
	if len(torrentName) > 50 {
		torrentName = torrentName[:47] + "..."
	}
	nameText := fmt.Sprintf("Torrent: %s", styles.TextStyle.Render(torrentName))

	// File deletion option
	var fileText string
	if m.deleteWithFiles {
		fileText = fmt.Sprintf("[%s] Delete files from disk", styles.ErrorStyle.Render("✓"))
	} else {
		fileText = fmt.Sprintf("[%s] Delete files from disk", styles.DimStyle.Render(" "))
	}

	// Instructions
	instructions := styles.DimStyle.Render("Y/Enter: Confirm  N/Esc: Cancel  F: Toggle files")

	// Combine all parts
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		nameText,
		"",
		fileText,
		"",
		instructions,
	)

	return dialogStyle.Render(content)
}

// renderAddDialog renders the add torrent dialog
func (m *MainView) renderAddDialog() string {
	// Dialog box styling
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.AccentStyle.GetForeground()).
		Padding(1, 2).
		Width(70).
		Height(25).
		Align(lipgloss.Center)

	// Title with mode tabs
	title := styles.AccentStyle.Render("Add Torrent")

	// Mode tabs
	var fileTab, urlTab string
	if m.addDialog.mode == ModeFile {
		fileTab = styles.SelectedRowStyle.Render("[File Browser]")
		urlTab = styles.DimStyle.Render(" URL ")
	} else {
		fileTab = styles.DimStyle.Render(" File Browser ")
		urlTab = styles.SelectedRowStyle.Render("[URL]")
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Left, fileTab, "  ", urlTab)

	// Content based on mode
	var content string
	if m.addDialog.mode == ModeFile {
		content = m.renderFileNavigator()
	} else {
		content = m.renderURLInput()
	}

	// Instructions
	instructions := styles.DimStyle.Render("Tab: Switch mode  Enter: Add  Esc: Cancel")

	// Combine all parts
	dialogContent := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		tabs,
		"",
		content,
		"",
		instructions,
	)

	return dialogStyle.Render(dialogContent)
}

// renderFileNavigator renders the file navigator component
func (m *MainView) renderFileNavigator() string {
	nav := m.addDialog.fileNav

	// Current path display
	pathDisplay := fmt.Sprintf("📁 %s", styles.DimStyle.Render(nav.currentPath))

	// Search pattern and match count
	matchCount := len(nav.filtered) - countDirectories(nav.filtered) // Subtract directories
	searchDisplay := fmt.Sprintf("🔍 %s", styles.AccentStyle.Render(nav.searchPattern))
	if nav.searchMode {
		searchDisplay += styles.AccentStyle.Render("▊") // Show cursor when in search mode
	}
	if matchCount > 0 {
		searchDisplay += styles.DimStyle.Render(fmt.Sprintf("  [%d files]", matchCount))
	}

	// File listing
	var fileLines []string
	maxLines := 12 // Limit displayed files
	startIdx := 0

	// Calculate visible range (simple scrolling)
	if nav.selectedIdx >= maxLines {
		startIdx = nav.selectedIdx - maxLines + 1
	}

	for i := startIdx; i < len(nav.filtered) && i < startIdx+maxLines; i++ {
		entry := nav.filtered[i]
		line := m.formatFileEntry(entry, i == nav.selectedIdx)
		fileLines = append(fileLines, line)
	}

	// Ensure we have some content
	if len(fileLines) == 0 {
		fileLines = append(fileLines, styles.DimStyle.Render("No files match pattern"))
	}

	// Instructions
	navInstructions := styles.DimStyle.Render("↑↓: Navigate  Enter: Select  /: Filter  h: Up dir")

	// Combine all parts
	return lipgloss.JoinVertical(lipgloss.Left,
		pathDisplay,
		searchDisplay,
		strings.Repeat("─", 60),
		strings.Join(fileLines, "\n"),
		strings.Repeat("─", 60),
		navInstructions,
	)
}

// renderURLInput renders the URL input component
func (m *MainView) renderURLInput() string {
	urlInput := m.addDialog.urlInput

	// URL input field
	inputStyle := styles.InputStyle.Width(50)

	// Show actual URL or placeholder
	urlDisplay := urlInput.url
	if len(urlDisplay) == 0 {
		urlDisplay = styles.DimStyle.Render("https://example.com/file.torrent")
	} else {
		urlDisplay = styles.TextStyle.Render(urlDisplay + "▊") // Show cursor
	}

	inputField := inputStyle.Render(urlDisplay)

	// Instructions
	urlInstructions := styles.DimStyle.Render("Enter URL to .torrent file")

	return lipgloss.JoinVertical(lipgloss.Center,
		"",
		"URL:",
		inputField,
		"",
		urlInstructions,
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	)
}

// formatFileEntry formats a file entry for display
func (m *MainView) formatFileEntry(entry FileEntry, selected bool) string {
	var icon, sizeStr string

	if entry.isDir {
		icon = "📁"
		if entry.name == ".." {
			sizeStr = ""
		} else {
			sizeStr = styles.DimStyle.Render("DIR")
		}
	} else {
		icon = "📄"
		sizeStr = styles.FormatBytes(entry.size)
	}

	// Truncate filename if too long
	name := entry.name
	maxNameLen := 40
	if len(name) > maxNameLen {
		name = name[:maxNameLen-3] + "..."
	}

	line := fmt.Sprintf("%s %s", icon, name)

	// Add size info (right-aligned)
	if sizeStr != "" {
		padding := 50 - len(line)
		if padding > 0 {
			line += strings.Repeat(" ", padding) + sizeStr
		}
	}

	// Apply selection styling
	if selected {
		return styles.SelectedRowStyle.Render(line)
	}
	return line
}

// countDirectories counts directory entries in the filtered list
func countDirectories(entries []FileEntry) int {
	count := 0
	for _, entry := range entries {
		if entry.isDir {
			count++
		}
	}
	return count
}

// NewAddTorrentDialog creates a new add torrent dialog
func NewAddTorrentDialog(startPath string) *AddTorrentDialog {
	fileNav := &FileNavigator{
		currentPath:   startPath,
		selectedIdx:   0,
		searchPattern: "*.torrent", // Default to torrent files
		searchMode:    false,
	}
	fileNav.readDirectory() // Initialize directory contents

	return &AddTorrentDialog{
		mode:    ModeFile, // Start with file browser by default
		fileNav: fileNav,
		urlInput: &URLInput{
			url:    "",
			cursor: 0,
		},
	}
}

// NewFileNavigator creates a new file navigator
func NewFileNavigator(startPath string) *FileNavigator {
	fn := &FileNavigator{
		currentPath:   startPath,
		selectedIdx:   0,
		searchPattern: "*.torrent",
		searchMode:    false,
	}
	fn.readDirectory()
	return fn
}

// readDirectory reads the current directory and applies filter
func (f *FileNavigator) readDirectory() {
	f.allEntries = []FileEntry{}
	f.selectedIdx = 0

	entries, err := os.ReadDir(f.currentPath)
	if err != nil {
		return // Handle error gracefully by showing empty directory
	}

	// Add parent directory entry if not at root
	if f.currentPath != "/" && f.currentPath != "" {
		f.allEntries = append(f.allEntries, FileEntry{
			name:     "..",
			isDir:    true,
			size:     0,
			fullPath: filepath.Dir(f.currentPath),
		})
	}

	// Add all entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(f.currentPath, entry.Name())
		f.allEntries = append(f.allEntries, FileEntry{
			name:     entry.Name(),
			isDir:    entry.IsDir(),
			size:     info.Size(),
			fullPath: fullPath,
		})
	}

	f.applyFilter()
}

// applyFilter applies the current search pattern to files
func (f *FileNavigator) applyFilter() {
	f.filtered = []FileEntry{}

	// Always include directories for navigation
	for _, entry := range f.allEntries {
		if entry.isDir {
			f.filtered = append(f.filtered, entry)
		}
	}

	// Add matching files
	for _, entry := range f.allEntries {
		if !entry.isDir {
			matched, _ := filepath.Match(f.searchPattern, entry.name)
			if matched {
				f.filtered = append(f.filtered, entry)
			}
		}
	}

	// Ensure selectedIdx is in bounds
	if f.selectedIdx >= len(f.filtered) {
		f.selectedIdx = len(f.filtered) - 1
	}
	if f.selectedIdx < 0 {
		f.selectedIdx = 0
	}
}

// handleFileNavigatorKeys handles keyboard input for file navigator
func (m *MainView) handleFileNavigatorKeys(key string) tea.Cmd {
	nav := m.addDialog.fileNav

	switch key {
	case "up", "k":
		if nav.selectedIdx > 0 {
			nav.selectedIdx--
		}
	case "down", "j":
		if nav.selectedIdx < len(nav.filtered)-1 {
			nav.selectedIdx++
		}
	case "enter":
		if nav.selectedIdx < len(nav.filtered) {
			selected := nav.filtered[nav.selectedIdx]
			if selected.isDir {
				// Navigate into directory
				nav.currentPath = selected.fullPath
				nav.readDirectory()
			} else {
				// Select file for upload
				return m.addTorrentFile(selected.fullPath)
			}
		}
	case "h", "backspace":
		// Go up one directory
		if nav.currentPath != "/" && nav.currentPath != "" {
			nav.currentPath = filepath.Dir(nav.currentPath)
			nav.readDirectory()
		}
	case "/":
		// Toggle search mode
		nav.searchMode = !nav.searchMode
		if nav.searchMode {
			nav.searchPattern = ""
		} else {
			nav.searchPattern = "*.torrent"
		}
		nav.applyFilter()
	default:
		// Handle search input if in search mode
		if nav.searchMode {
			if key == "backspace" {
				if len(nav.searchPattern) > 0 {
					nav.searchPattern = nav.searchPattern[:len(nav.searchPattern)-1]
					nav.applyFilter()
				}
			} else if len(key) == 1 {
				nav.searchPattern += key
				nav.applyFilter()
			}
		}
	}

	return nil
}

// handleURLInputKeys handles keyboard input for URL input
func (m *MainView) handleURLInputKeys(keyMsg tea.KeyMsg) tea.Cmd {
	urlInput := m.addDialog.urlInput

	switch keyMsg.String() {
	case "enter":
		if len(urlInput.url) > 0 {
			return m.addTorrentURL(urlInput.url)
		}
	case "backspace":
		if len(urlInput.url) > 0 {
			urlInput.url = urlInput.url[:len(urlInput.url)-1]
		}
	case "ctrl+a":
		// Select all (clear field)
		urlInput.url = ""
	default:
		// Handle character input - this includes paste operations
		// Bubble Tea provides the input as runes which handles multi-character paste
		if len(keyMsg.Runes) > 0 {
			for _, r := range keyMsg.Runes {
				// Only add printable characters (includes all URL chars: :, /, ?, =, etc.)
				if r >= 32 && r <= 126 {
					urlInput.url += string(r)
				}
			}
		}
	}

	return nil
}

// addTorrentFile adds a torrent from a local file
func (m *MainView) addTorrentFile(filePath string) tea.Cmd {
	m.showAddDialog = false

	return func() tea.Msg {
		ctx := context.Background()
		err := m.apiClient.AddTorrentFile(ctx, filePath)
		if err != nil {
			return errorMsg(fmt.Errorf("failed to add torrent file: %w", err))
		}
		filename := filepath.Base(filePath)
		return successMsg(fmt.Sprintf("added torrent: %s", styles.TruncateString(filename, 40)))
	}
}

// addTorrentURL adds a torrent from a URL
func (m *MainView) addTorrentURL(url string) tea.Cmd {
	m.showAddDialog = false

	return func() tea.Msg {
		ctx := context.Background()
		err := m.apiClient.AddTorrentURL(ctx, url)
		if err != nil {
			return errorMsg(fmt.Errorf("failed to add torrent from URL: %w", err))
		}
		return successMsg(fmt.Sprintf("added torrent from URL"))
	}
}

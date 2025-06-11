package views

import (
	"context"
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
	tickMsg           time.Time
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
	torrents      []api.Torrent
	allTorrents   []api.Torrent // unfiltered torrents
	stats         *api.GlobalStats
	categories    map[string]interface{}
	tags          []string
	currentFilter filter.Filter
	viewMode      ViewMode
	lastError     error
	isLoading     bool

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
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Escape}, // Navigation and Actions
		{k.Refresh, k.Filter},             // Features
		{k.Help, k.Quit},                  // General
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
	}
}

// NewMainView creates a new main view
func NewMainView(cfg *config.Config, client api.ClientInterface) *MainView {
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
		m.filterPanel.SetAvailableOptions(categoryNames, m.extractTrackerNames(), m.tags)

	case tagsDataMsg:
		m.tags = []string(msg)
		m.filterPanel.SetAvailableOptions(m.extractCategoryNames(), m.extractTrackerNames(), m.tags)

	case errorMsg:
		m.lastError = error(msg)
		m.isLoading = false

	case components.DetailsDataMsg:
		// Pass details data to torrent details component
		m.torrentDetails, cmd = m.torrentDetails.Update(msg)
		cmds = append(cmds, cmd)

	case tickMsg:
		// Refresh data periodically
		cmds = append(cmds, m.fetchAllData(), m.tickCmd())

	case tea.KeyMsg:
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

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Escape):
			if m.viewMode == ViewModeDetails {
				m.viewMode = ViewModeMain
			} else if m.viewMode == ViewModeMain {
				// Let filter panel handle escape to exit search mode
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
				// If filter panel is in any interactive mode, let it handle enter
				if m.filterPanel.IsInInteractiveMode() {
					oldFilter := m.filterPanel.GetFilter()
					m.filterPanel, cmd = m.filterPanel.Update(msg)
					cmds = append(cmds, cmd)
					if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
						m.currentFilter = m.filterPanel.GetFilter()
						m.applyFilter()
					}
				} else {
					// Show details for selected torrent
					selectedHash := m.torrentList.GetSelectedHash()
					if selectedHash != "" {
						for _, torrent := range m.torrents {
							if torrent.Hash == selectedHash {
								cmd = m.torrentDetails.SetTorrent(&torrent)
								cmds = append(cmds, cmd)
								m.viewMode = ViewModeDetails
								break
							}
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

		case msg.String() == "C": // Column configuration
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

		case msg.String() == "a": // Tag filter (tAgs)
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
			} else if m.torrentList.IsInConfigMode() {
				// Torrent list is in column config mode - let it handle all keys
				m.torrentList, cmd = m.torrentList.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.filterPanel.IsInInteractiveMode() {
				// Filter panel is in interactive mode (state/category/tracker/tag selection)
				// Let it handle navigation keys
				oldFilter := m.filterPanel.GetFilter()
				m.filterPanel, cmd = m.filterPanel.Update(msg)
				cmds = append(cmds, cmd)
				if !filterEqual(oldFilter, m.filterPanel.GetFilter()) {
					m.currentFilter = m.filterPanel.GetFilter()
					m.applyFilter()
				}
			} else {
				// Normal mode - pass navigation keys to torrent list (main focus)
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

	// Help at the very bottom
	helpView := m.help.View(m.keys)
	sections = append(sections, helpView)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderStatsPanel renders the stats panel
func (m *MainView) renderStatsPanel(width, height int) string {
	style := styles.PanelStyle

	// Set dimensions for the stats panel component
	m.statsPanel.SetDimensions(width-6, height-4)

	content := m.statsPanel.View()
	return style.Width(width).Height(height).Render(content)
}

// renderTorrentList renders the torrent list (always focused)
func (m *MainView) renderTorrentList(width, height int) string {
	style := styles.FocusedPanelStyle // Always focused

	// Set dimensions for the torrent list component
	// Account for panel borders (2) and padding (4 horizontal, 2 vertical)
	m.torrentList.SetDimensions(width-6, height-4)

	content := m.torrentList.View()
	return style.Width(width).Height(height).Render(content)
}

// renderFilterPanel renders the filter panel
func (m *MainView) renderFilterPanel(width, height int) string {
	style := styles.PanelStyle

	// Set dimensions for the filter panel component
	m.filterPanel.SetDimensions(width-6, height-4)

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
		m.torrentDetails.SetSize(m.width-4, contentHeight-4)
	}
}

// Removed focus cycling methods - using single-focus design

// applyFilter applies the current filter to torrents
func (m *MainView) applyFilter() {
	m.torrents = m.currentFilter.Apply(m.allTorrents)
	m.torrentList.SetTorrents(m.torrents)
}

// extractCategoryNames extracts category names from the categories map
func (m *MainView) extractCategoryNames() []string {
	var names []string
	for name := range m.categories {
		names = append(names, name)
	}
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
	m.torrentDetails.SetSize(m.width-4, contentHeight-4) // Account for panel borders

	// Render the details in a panel
	content := m.torrentDetails.View()
	detailsPanel := styles.FocusedPanelStyle.Width(m.width).Height(contentHeight).Render(content)

	// Add help at the bottom
	helpView := m.help.View(m.keys)

	return lipgloss.JoinVertical(lipgloss.Left, detailsPanel, helpView)
}

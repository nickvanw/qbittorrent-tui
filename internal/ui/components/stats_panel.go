package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/styles"
)

// StatsPanel displays global statistics
type StatsPanel struct {
	stats  *api.GlobalStats
	width  int
	height int
}

// NewStatsPanel creates a new stats panel
func NewStatsPanel() *StatsPanel {
	return &StatsPanel{}
}

// SetStats updates the statistics
func (s *StatsPanel) SetStats(stats *api.GlobalStats) {
	s.stats = stats
}

// Update handles messages
func (s *StatsPanel) Update(msg tea.Msg) (*StatsPanel, tea.Cmd) {
	// Stats panel doesn't handle any specific messages
	return s, nil
}

// View renders the stats panel
func (s *StatsPanel) View() string {
	if s.stats == nil {
		return styles.DimStyle.Render("Loading statistics...")
	}

	// Add title
	title := styles.TitleStyle.Render("Global Statistics")

	// For narrow terminals, use vertical layout
	if s.width < 80 {
		var lines []string
		lines = append(lines, title)
		lines = append(lines, s.renderCompactStats())
		return strings.Join(lines, "\n")
	}

	// Calculate section widths based on available width
	// Divide available width among 3 sections
	sectionWidth := s.width / 3
	if sectionWidth < 20 {
		sectionWidth = 20 // minimum section width
	}

	// Create stats grid
	var sections []string

	// Connection status
	connStatus := s.renderConnectionStatusWithWidth(sectionWidth)
	sections = append(sections, connStatus)

	// Transfer stats
	transferStats := s.renderTransferStatsWithWidth(sectionWidth)
	sections = append(sections, transferStats)

	// Session stats
	sessionStats := s.renderSessionStatsWithWidth(sectionWidth)
	sections = append(sections, sessionStats)

	// Join sections horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, sections...)

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

// SetDimensions updates the component dimensions
func (s *StatsPanel) SetDimensions(width, height int) {
	s.width = width
	s.height = height
}

// renderCompactStats renders stats in a compact vertical format for narrow terminals
func (s *StatsPanel) renderCompactStats() string {
	var lines []string

	// Connection status
	connState := "Disconnected"
	stateStyle := styles.ErrorStyle
	if s.stats.ConnectionStatus == "connected" {
		connState = "Connected"
		stateStyle = styles.DownloadingStyle
	} else if s.stats.ConnectionStatus == "firewalled" {
		connState = "Firewalled"
		stateStyle = styles.WarningStyle
	}

	// Speed info
	dlSpeed := styles.FormatSpeed(s.stats.DlInfoSpeed)
	upSpeed := styles.FormatSpeed(s.stats.UpInfoSpeed)

	// Primary status line
	lines = append(lines, fmt.Sprintf("%s | ↓ %s ↑ %s | %d torrents",
		stateStyle.Render(connState),
		styles.DownloadingStyle.Render(dlSpeed),
		styles.SeedingStyle.Render(upSpeed),
		s.stats.TorrentsCount))

	// Session data line
	totalDl := styles.FormatBytes(s.stats.DlInfoData)
	totalUp := styles.FormatBytes(s.stats.UpInfoData)
	lines = append(lines, fmt.Sprintf("Session: ↓%s ↑%s", totalDl, totalUp))

	// Additional info line
	freeSpace := styles.FormatBytes(s.stats.FreeSpaceOnDisk)
	dhtNodes := fmt.Sprintf("%d nodes", s.stats.DHTNodes)
	lines = append(lines, fmt.Sprintf("Free: %s | DHT: %s", freeSpace, dhtNodes))

	return strings.Join(lines, "\n")
}

// renderConnectionStatusWithWidth renders the connection status section with specified width
func (s *StatsPanel) renderConnectionStatusWithWidth(width int) string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Connection"))

	// Connection state
	connState := "Disconnected"
	stateStyle := styles.ErrorStyle
	if s.stats.ConnectionStatus == "connected" {
		connState = "Connected"
		stateStyle = styles.DownloadingStyle
	} else if s.stats.ConnectionStatus == "firewalled" {
		connState = "Firewalled"
		stateStyle = styles.WarningStyle
	}
	lines = append(lines, fmt.Sprintf("Status: %s", stateStyle.Render(connState)))

	// DHT nodes
	dhtNodes := fmt.Sprintf("DHT: %d nodes", s.stats.DHTNodes)
	lines = append(lines, styles.DimStyle.Render(dhtNodes))

	return lipgloss.NewStyle().
		Width(width-2). // Account for padding
		Padding(0, 1, 0, 0).
		Render(strings.Join(lines, "\n"))
}

// renderTransferStatsWithWidth renders the transfer statistics section with specified width
func (s *StatsPanel) renderTransferStatsWithWidth(width int) string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Transfer"))

	// Download and Upload speeds on same line for clarity
	dlSpeed := styles.FormatSpeed(s.stats.DlInfoSpeed)
	upSpeed := styles.FormatSpeed(s.stats.UpInfoSpeed)
	speedLine := fmt.Sprintf("↓ %s  ↑ %s",
		styles.DownloadingStyle.Render(dlSpeed),
		styles.SeedingStyle.Render(upSpeed))
	lines = append(lines, speedLine)

	// Total downloaded/uploaded
	totalDl := styles.FormatBytes(s.stats.DlInfoData)
	totalUp := styles.FormatBytes(s.stats.UpInfoData)
	ratio := float64(s.stats.UpInfoData) / float64(s.stats.DlInfoData)
	if s.stats.DlInfoData == 0 {
		ratio = 0
	}

	lines = append(lines, styles.DimStyle.Render(fmt.Sprintf("Session: ↓%s ↑%s (%.2f)", totalDl, totalUp, ratio)))

	return lipgloss.NewStyle().
		Width(width-2). // Account for padding
		Padding(0, 1, 0, 0).
		Render(strings.Join(lines, "\n"))
}

// renderSessionStatsWithWidth renders the session statistics section with specified width
func (s *StatsPanel) renderSessionStatsWithWidth(width int) string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Session"))

	// Active torrents
	activeTorrents := fmt.Sprintf("Active: %d torrents", s.stats.TorrentsCount)
	lines = append(lines, activeTorrents)

	// Free space
	freeSpace := styles.FormatBytes(s.stats.FreeSpaceOnDisk)
	lines = append(lines, fmt.Sprintf("Free space: %s", freeSpace))

	return lipgloss.NewStyle().
		Width(width - 2). // Account for padding
		Render(strings.Join(lines, "\n"))
}

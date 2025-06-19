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

	// Create stats grid
	var sections []string

	// Connection status
	connStatus := s.renderConnectionStatus()
	sections = append(sections, connStatus)

	// Transfer stats
	transferStats := s.renderTransferStats()
	sections = append(sections, transferStats)

	// Session stats
	sessionStats := s.renderSessionStats()
	sections = append(sections, sessionStats)

	// Join sections horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, sections...)

	// Add title
	title := styles.TitleStyle.Render("Global Statistics")

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

// renderConnectionStatus renders the connection status section
func (s *StatsPanel) renderConnectionStatus() string {
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
		Width(30).
		Padding(0, 2, 0, 0).
		Render(strings.Join(lines, "\n"))
}

// renderTransferStats renders the transfer statistics section
func (s *StatsPanel) renderTransferStats() string {
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
		Width(50).
		Padding(0, 2, 0, 0).
		Render(strings.Join(lines, "\n"))
}

// renderSessionStats renders the session statistics section
func (s *StatsPanel) renderSessionStats() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Session"))

	// Note: Torrent counts removed - not available from qBittorrent API
	// Would need to be calculated from torrent list if needed

	// Free space
	freeSpace := styles.FormatBytes(s.stats.FreeSpaceOnDisk)
	lines = append(lines, fmt.Sprintf("Free space: %s", freeSpace))

	return lipgloss.NewStyle().
		Width(30).
		Render(strings.Join(lines, "\n"))
}

// SetDimensions updates the component dimensions
func (s *StatsPanel) SetDimensions(width, height int) {
	s.width = width
	s.height = height
}

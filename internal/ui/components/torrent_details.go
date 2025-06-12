package components

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/styles"
)

// DetailsTab represents the different detail tabs
type DetailsTab int

const (
	TabGeneral DetailsTab = iota
	TabTrackers
	TabPeers
	TabFiles
)

// TorrentDetails displays detailed information about a single torrent with tabs
type TorrentDetails struct {
	torrent    *api.Torrent
	properties *api.TorrentProperties
	trackers   []api.Tracker
	peers      map[string]api.Peer
	files      []api.TorrentFile
	client     api.ClientInterface
	width      int
	height     int
	scroll     int
	activeTab  DetailsTab
	isLoading  bool
	lastError  error
}

// NewTorrentDetails creates a new torrent details component
func NewTorrentDetails(client api.ClientInterface) *TorrentDetails {
	return &TorrentDetails{
		client:    client,
		activeTab: TabGeneral,
	}
}

// SetTorrent updates the torrent being displayed and fetches detailed data
func (t *TorrentDetails) SetTorrent(torrent *api.Torrent) tea.Cmd {
	t.torrent = torrent
	t.scroll = 0
	t.activeTab = TabGeneral
	t.isLoading = true
	t.lastError = nil

	// Fetch detailed data for all tabs
	return t.fetchDetailedData()
}

// UpdateTorrent updates the torrent data without resetting UI state (scroll, activeTab)
func (t *TorrentDetails) UpdateTorrent(torrent *api.Torrent) {
	t.torrent = torrent
}

// SetSize updates the component dimensions
func (t *TorrentDetails) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// DetailsDataMsg represents detailed torrent data
type DetailsDataMsg struct {
	Properties *api.TorrentProperties
	Trackers   []api.Tracker
	Peers      map[string]api.Peer
	Files      []api.TorrentFile
	Err        error
}

// fetchDetailedData fetches all detailed data for the torrent
func (t *TorrentDetails) fetchDetailedData() tea.Cmd {
	if t.torrent == nil {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		hash := t.torrent.Hash

		var properties *api.TorrentProperties
		var trackers []api.Tracker
		var peers map[string]api.Peer
		var files []api.TorrentFile
		var err error

		// Fetch properties
		properties, err = t.client.GetTorrentProperties(ctx, hash)
		if err != nil {
			return DetailsDataMsg{Err: fmt.Errorf("failed to get properties: %w", err)}
		}

		// Fetch trackers
		trackers, err = t.client.GetTorrentTrackers(ctx, hash)
		if err != nil {
			return DetailsDataMsg{Err: fmt.Errorf("failed to get trackers: %w", err)}
		}

		// Fetch peers
		peers, err = t.client.GetTorrentPeers(ctx, hash)
		if err != nil {
			return DetailsDataMsg{Err: fmt.Errorf("failed to get peers: %w", err)}
		}

		// Fetch files
		files, err = t.client.GetTorrentFiles(ctx, hash)
		if err != nil {
			return DetailsDataMsg{Err: fmt.Errorf("failed to get files: %w", err)}
		}

		return DetailsDataMsg{
			Properties: properties,
			Trackers:   trackers,
			Peers:      peers,
			Files:      files,
		}
	}
}

// Update handles messages
func (t *TorrentDetails) Update(msg tea.Msg) (*TorrentDetails, tea.Cmd) {
	switch msg := msg.(type) {
	case time.Time:
		// Handle tick messages for periodic refresh
		if t.torrent != nil {
			return t, t.fetchDetailedData()
		}

	case DetailsDataMsg:
		t.isLoading = false
		if msg.Err != nil {
			t.lastError = msg.Err
		} else {
			t.properties = msg.Properties
			t.trackers = msg.Trackers
			t.peers = msg.Peers
			t.files = msg.Files
			t.lastError = nil

			// Sort trackers for stable display order
			t.sortTrackers()
		}

	case tea.KeyMsg:
		switch {
		// Tab navigation
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			t.activeTab = TabGeneral
			t.scroll = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			t.activeTab = TabTrackers
			t.scroll = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			t.activeTab = TabPeers
			t.scroll = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			t.activeTab = TabFiles
			t.scroll = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if t.activeTab > TabGeneral {
				t.activeTab--
				t.scroll = 0
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if t.activeTab < TabFiles {
				t.activeTab++
				t.scroll = 0
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			// Cycle through tabs
			t.activeTab = (t.activeTab + 1) % 4
			t.scroll = 0

		// Scrolling
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if t.scroll > 0 {
				t.scroll--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			t.scroll++
		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			t.scroll = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			maxScroll := t.getMaxScroll()
			if maxScroll > 0 {
				t.scroll = maxScroll
			}
		}
	}
	return t, nil
}

// View renders the torrent details with tabs
func (t *TorrentDetails) View() string {
	if t.torrent == nil {
		return styles.DimStyle.Render("No torrent selected")
	}

	var sections []string

	// Title
	title := styles.TitleStyle.Render(fmt.Sprintf("Torrent Details: %s", t.torrent.Name))
	sections = append(sections, title)

	// Tab bar
	tabBar := t.renderTabBar()
	sections = append(sections, tabBar)

	// Content based on active tab
	var content string
	if t.isLoading {
		content = styles.DimStyle.Render("Loading detailed information...")
	} else if t.lastError != nil {
		content = styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", t.lastError))
	} else {
		switch t.activeTab {
		case TabGeneral:
			content = t.renderGeneralTab()
		case TabTrackers:
			content = t.renderTrackersTab()
		case TabPeers:
			content = t.renderPeersTab()
		case TabFiles:
			content = t.renderFilesTab()
		}
	}

	sections = append(sections, content)

	// Help text
	help := styles.DimStyle.Render("↑↓ scroll • ←→ or 1-4 tabs • Tab cycle • Esc back")
	sections = append(sections, help)

	fullContent := strings.Join(sections, "\n\n")

	// Apply scrolling
	lines := strings.Split(fullContent, "\n")
	if t.height > 0 {
		visibleLines := t.height - 2 // Account for borders
		if visibleLines < 1 {
			visibleLines = 1
		}

		maxScroll := len(lines) - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}
		if t.scroll > maxScroll {
			t.scroll = maxScroll
		}

		if len(lines) > visibleLines {
			endLine := t.scroll + visibleLines
			if endLine > len(lines) {
				endLine = len(lines)
			}
			lines = lines[t.scroll:endLine]
		}
	}

	return strings.Join(lines, "\n")
}

// renderTabBar renders the tab navigation bar
func (t *TorrentDetails) renderTabBar() string {
	tabs := []string{"General", "Trackers", "Peers", "Files"}
	var renderedTabs []string

	for i, tab := range tabs {
		if DetailsTab(i) == t.activeTab {
			renderedTabs = append(renderedTabs, styles.SelectedRowStyle.Render(fmt.Sprintf("[%s]", tab)))
		} else {
			renderedTabs = append(renderedTabs, styles.DimStyle.Render(fmt.Sprintf(" %s ", tab)))
		}
	}

	return strings.Join(renderedTabs, " ")
}

// renderGeneralTab renders the general information tab (existing functionality)
func (t *TorrentDetails) renderGeneralTab() string {
	var sections []string

	// Basic information
	basicInfo := t.renderBasicInfo()
	sections = append(sections, basicInfo)

	// Transfer information
	transferInfo := t.renderTransferInfo()
	sections = append(sections, transferInfo)

	// Files and paths
	pathInfo := t.renderPathInfo()
	sections = append(sections, pathInfo)

	// Technical details
	techInfo := t.renderTechnicalInfo()
	sections = append(sections, techInfo)

	return strings.Join(sections, "\n\n")
}

// renderTrackersTab renders the trackers information
func (t *TorrentDetails) renderTrackersTab() string {
	if len(t.trackers) == 0 {
		return styles.DimStyle.Render("No trackers found")
	}

	var lines []string
	lines = append(lines, styles.SubtitleStyle.Render("Trackers"))
	lines = append(lines, "")

	for i, tracker := range t.trackers {
		status := t.getTrackerStatus(tracker.Status)
		statusStyle := styles.DimStyle
		if tracker.Status == 2 { // Working
			statusStyle = styles.AccentStyle
		} else if tracker.Status == 4 { // Error
			statusStyle = styles.ErrorStyle
		}

		line := fmt.Sprintf("%d. %s", i+1, tracker.URL)
		lines = append(lines, styles.TextStyle.Render(line))

		details := fmt.Sprintf("   Status: %s | Tier: %d | Peers: %d | Seeds: %d | Leeches: %d",
			statusStyle.Render(status), tracker.Tier, tracker.NumPeers, tracker.NumSeeds, tracker.NumLeeches)
		lines = append(lines, details)

		if tracker.Msg != "" && tracker.Msg != status {
			msg := fmt.Sprintf("   Message: %s", tracker.Msg)
			lines = append(lines, styles.DimStyle.Render(msg))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderPeersTab renders the peers information
func (t *TorrentDetails) renderPeersTab() string {
	if len(t.peers) == 0 {
		return styles.DimStyle.Render("No peers connected")
	}

	var lines []string
	lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("Peers (%d connected)", len(t.peers))))
	lines = append(lines, "")

	// Calculate available width (subtract padding)
	availableWidth := t.width - 4
	if availableWidth < 80 {
		availableWidth = 80 // Minimum width
	}

	// Define column configurations with responsive widths
	columns := []struct {
		title    string
		minWidth int
		flex     float64
	}{
		{"IP:Port", 18, 0.2},
		{"Client", 15, 0.25},
		{"Country", 8, 0.1},
		{"Connection", 10, 0.15},
		{"Progress", 8, 0.1},
		{"DL Speed", 9, 0.1},
		{"UL Speed", 9, 0.1},
	}

	// Calculate actual column widths
	totalMinWidth := 0
	for _, col := range columns {
		totalMinWidth += col.minWidth
	}

	// Calculate flex space
	flexSpace := availableWidth - totalMinWidth
	if flexSpace < 0 {
		flexSpace = 0
	}

	var widths []int
	for _, col := range columns {
		width := col.minWidth + int(float64(flexSpace)*col.flex)
		widths = append(widths, width)
	}

	// Build header
	var headerParts []string
	for i, col := range columns {
		headerParts = append(headerParts, t.padString(col.title, widths[i]))
	}
	lines = append(lines, styles.DimStyle.Render(strings.Join(headerParts, "")))
	lines = append(lines, styles.DimStyle.Render(strings.Repeat("─", availableWidth)))

	// Build data rows
	for _, peer := range t.peers {
		progress := fmt.Sprintf("%.1f%%", peer.Progress*100)
		dlSpeed := formatBytes(peer.DlSpeed) + "/s"
		ulSpeed := formatBytes(peer.UpSpeed) + "/s"
		address := fmt.Sprintf("%s:%d", peer.IP, peer.Port)

		values := []string{
			address,
			peer.Client,
			peer.Country,
			peer.Connection,
			progress,
			dlSpeed,
			ulSpeed,
		}

		var rowParts []string
		for i, value := range values {
			rowParts = append(rowParts, t.padString(value, widths[i]))
		}
		lines = append(lines, strings.Join(rowParts, ""))
	}

	return strings.Join(lines, "\n")
}

// padString pads or truncates a string to the specified width
func (t *TorrentDetails) padString(s string, width int) string {
	// Use rune count for proper Unicode handling
	runes := []rune(s)
	runeLen := len(runes)

	if runeLen > width {
		// Truncate if too long
		if width <= 3 {
			return string(runes[:width])
		}
		return string(runes[:width-3]) + "..."
	}
	// Pad with spaces if too short
	return s + strings.Repeat(" ", width-runeLen)
}

// renderFilesTab renders the files information
func (t *TorrentDetails) renderFilesTab() string {
	if len(t.files) == 0 {
		return styles.DimStyle.Render("No files found")
	}

	var lines []string
	lines = append(lines, styles.SubtitleStyle.Render(fmt.Sprintf("Files (%d total)", len(t.files))))
	lines = append(lines, "")

	// Calculate available width (subtract padding)
	availableWidth := t.width - 4
	if availableWidth < 80 {
		availableWidth = 80 // Minimum width
	}

	// Define column configurations with responsive widths
	columns := []struct {
		title    string
		minWidth int
		flex     float64
	}{
		{"Name", 35, 0.6}, // Give most space to filename
		{"Size", 10, 0.15},
		{"Progress", 10, 0.15},
		{"Priority", 8, 0.1},
	}

	// Calculate actual column widths
	totalMinWidth := 0
	for _, col := range columns {
		totalMinWidth += col.minWidth
	}

	// Calculate flex space
	flexSpace := availableWidth - totalMinWidth
	if flexSpace < 0 {
		flexSpace = 0
	}

	var widths []int
	for _, col := range columns {
		width := col.minWidth + int(float64(flexSpace)*col.flex)
		widths = append(widths, width)
	}

	// Build header
	var headerParts []string
	for i, col := range columns {
		headerParts = append(headerParts, t.padString(col.title, widths[i]))
	}
	lines = append(lines, styles.DimStyle.Render(strings.Join(headerParts, "")))
	lines = append(lines, styles.DimStyle.Render(strings.Repeat("─", availableWidth)))

	// Build data rows
	for _, file := range t.files {
		size := formatBytes(file.Size)
		progress := fmt.Sprintf("%.1f%%", file.Progress*100)
		priority := t.getFilePriority(file.Priority)

		values := []string{
			file.Name,
			size,
			progress,
			priority,
		}

		var rowParts []string
		for i, value := range values {
			rowParts = append(rowParts, t.padString(value, widths[i]))
		}

		line := strings.Join(rowParts, "")

		// Style based on progress
		if file.Progress >= 1.0 {
			line = styles.AccentStyle.Render(line)
		} else if file.Progress > 0 {
			line = styles.TextStyle.Render(line)
		} else {
			line = styles.DimStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// Helper methods (keeping existing ones and adding new ones)

func (t *TorrentDetails) getMaxScroll() int {
	// This would need to calculate based on content and height
	// For now, return a reasonable default
	return 50
}

func (t *TorrentDetails) getTrackerStatus(status int) string {
	switch status {
	case 0:
		return "Disabled"
	case 1:
		return "Not contacted"
	case 2:
		return "Working"
	case 3:
		return "Updating"
	case 4:
		return "Error"
	default:
		return "Unknown"
	}
}

func (t *TorrentDetails) getFilePriority(priority int) string {
	switch priority {
	case 0:
		return "Skip"
	case 1:
		return "Normal"
	case 6:
		return "High"
	case 7:
		return "Max"
	default:
		return fmt.Sprintf("%d", priority)
	}
}

// Keep existing helper methods (renderBasicInfo, renderTransferInfo, etc.)
// I'll add these back from the original file...

func (t *TorrentDetails) renderBasicInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Basic Information"))
	lines = append(lines, fmt.Sprintf("Hash: %s", t.torrent.Hash))
	lines = append(lines, fmt.Sprintf("State: %s", t.getStatusDisplay(t.torrent.State)))
	lines = append(lines, fmt.Sprintf("Size: %s", formatBytes(t.torrent.Size)))
	lines = append(lines, fmt.Sprintf("Progress: %.2f%%", t.torrent.Progress*100))

	if t.torrent.Category != "" {
		lines = append(lines, fmt.Sprintf("Category: %s", t.torrent.Category))
	}

	if t.torrent.Tags != "" {
		lines = append(lines, fmt.Sprintf("Tags: %s", t.torrent.Tags))
	}

	// Format dates
	if t.torrent.AddedOn > 0 {
		addedTime := time.Unix(t.torrent.AddedOn, 0)
		lines = append(lines, fmt.Sprintf("Added: %s", addedTime.Format("2006-01-02 15:04:05")))
	}

	if t.torrent.CompletedOn > 0 {
		completedTime := time.Unix(t.torrent.CompletedOn, 0)
		lines = append(lines, fmt.Sprintf("Completed: %s", completedTime.Format("2006-01-02 15:04:05")))
	}

	return strings.Join(lines, "\n")
}

func (t *TorrentDetails) renderTransferInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Transfer Information"))
	lines = append(lines, fmt.Sprintf("Download Speed: %s/s", formatBytes(t.torrent.DlSpeed)))
	lines = append(lines, fmt.Sprintf("Upload Speed: %s/s", formatBytes(t.torrent.UpSpeed)))
	lines = append(lines, fmt.Sprintf("Downloaded: %s", formatBytes(t.torrent.Downloaded)))
	lines = append(lines, fmt.Sprintf("Uploaded: %s", formatBytes(t.torrent.Uploaded)))
	lines = append(lines, fmt.Sprintf("Ratio: %.3f", t.torrent.Ratio))

	if t.torrent.ETA > 0 && t.torrent.ETA < 8640000 { // Less than 100 days
		etaDuration := time.Duration(t.torrent.ETA) * time.Second
		lines = append(lines, fmt.Sprintf("ETA: %s", etaDuration.String()))
	} else if t.torrent.ETA == 8640000 {
		lines = append(lines, "ETA: ∞")
	}

	lines = append(lines, fmt.Sprintf("Seeds: %d (%d total)", t.torrent.NumSeeds, t.torrent.NumComplete))
	lines = append(lines, fmt.Sprintf("Peers: %d (%d total)", t.torrent.NumLeeches, t.torrent.NumIncomplete))

	return strings.Join(lines, "\n")
}

func (t *TorrentDetails) renderPathInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Files and Paths"))
	lines = append(lines, fmt.Sprintf("Save Path: %s", t.torrent.SavePath))

	if t.torrent.Tracker != "" {
		lines = append(lines, fmt.Sprintf("Tracker: %s", t.torrent.Tracker))
	}

	return strings.Join(lines, "\n")
}

func (t *TorrentDetails) renderTechnicalInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Technical Information"))
	lines = append(lines, fmt.Sprintf("Priority: %d", t.torrent.Priority))
	lines = append(lines, fmt.Sprintf("Auto TMM: %t", t.torrent.AutoTMM))

	if t.torrent.TimeActive > 0 {
		activeDuration := time.Duration(t.torrent.TimeActive) * time.Second
		lines = append(lines, fmt.Sprintf("Time Active: %s", activeDuration.String()))
	}

	if t.torrent.MaxRatio >= 0 {
		if t.torrent.MaxRatio == -1 {
			lines = append(lines, "Max Ratio: ∞")
		} else {
			lines = append(lines, fmt.Sprintf("Max Ratio: %.3f", t.torrent.MaxRatio))
		}
	}

	return strings.Join(lines, "\n")
}

func (t *TorrentDetails) getStatusDisplay(state string) string {
	switch state {
	case "error":
		return styles.ErrorStyle.Render("Error")
	case "missingFiles":
		return styles.ErrorStyle.Render("Missing files")
	case "uploading", "stalledUP":
		return styles.AccentStyle.Render("Seeding")
	case "downloading", "metaDL", "stalledDL":
		return styles.AccentStyle.Render("Downloading")
	case "pausedDL":
		return styles.DimStyle.Render("Paused")
	case "pausedUP":
		return styles.DimStyle.Render("Paused (seeding)")
	case "stoppedUP":
		return styles.DimStyle.Render("Paused (seeding)")
	case "queuedDL":
		return styles.DimStyle.Render("Queued (download)")
	case "queuedUP":
		return styles.DimStyle.Render("Queued (seeding)")
	case "checkingDL", "checkingUP":
		return styles.DimStyle.Render("Checking")
	case "checkingResumeData":
		return styles.DimStyle.Render("Checking resume data")
	case "moving":
		return styles.DimStyle.Render("Moving")
	default:
		return state
	}
}

// sortTrackers sorts trackers by tier (ascending) then by URL (alphabetically)
func (t *TorrentDetails) sortTrackers() {
	sort.Slice(t.trackers, func(i, j int) bool {
		// First sort by tier (lower tier = higher priority)
		if t.trackers[i].Tier != t.trackers[j].Tier {
			return t.trackers[i].Tier < t.trackers[j].Tier
		}
		// If tiers are equal, sort by URL alphabetically
		return strings.ToLower(t.trackers[i].URL) < strings.ToLower(t.trackers[j].URL)
	})
}

// formatBytes formats a byte count as human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

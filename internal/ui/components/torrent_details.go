package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/styles"
)

// TorrentDetails displays detailed information about a single torrent
type TorrentDetails struct {
	torrent *api.Torrent
	width   int
	height  int
	scroll  int
}

// NewTorrentDetails creates a new torrent details component
func NewTorrentDetails() *TorrentDetails {
	return &TorrentDetails{}
}

// SetTorrent updates the torrent being displayed
func (t *TorrentDetails) SetTorrent(torrent *api.Torrent) {
	t.torrent = torrent
	t.scroll = 0 // Reset scroll when setting new torrent
}

// Update handles messages
func (t *TorrentDetails) Update(msg tea.Msg) (*TorrentDetails, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if t.scroll > 0 {
				t.scroll--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			t.scroll++
		case key.Matches(msg, key.NewBinding(key.WithKeys("g"))):
			t.scroll = 0
		case key.Matches(msg, key.NewBinding(key.WithKeys("G"))):
			// Scroll to bottom
			maxScroll := t.getMaxScroll()
			if maxScroll > 0 {
				t.scroll = maxScroll
			}
		}
	}
	return t, nil
}

// View renders the torrent details
func (t *TorrentDetails) View() string {
	if t.torrent == nil {
		return styles.DimStyle.Render("No torrent selected")
	}

	var sections []string

	// Title
	title := styles.TitleStyle.Render(fmt.Sprintf("Torrent Details: %s", t.torrent.Name))
	sections = append(sections, title)

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

	content := strings.Join(sections, "\n\n")

	// Apply scrolling
	lines := strings.Split(content, "\n")
	if t.height > 0 {
		visibleLines := t.height - 2 // Account for borders
		if visibleLines < 1 {
			visibleLines = 1
		}

		// Ensure scroll doesn't exceed content
		maxScroll := len(lines) - visibleLines
		if maxScroll < 0 {
			maxScroll = 0
		}
		if t.scroll > maxScroll {
			t.scroll = maxScroll
		}

		// Get visible lines
		start := t.scroll
		end := start + visibleLines
		if end > len(lines) {
			end = len(lines)
		}

		if start < len(lines) {
			content = strings.Join(lines[start:end], "\n")
		}
	}

	return content
}

// renderBasicInfo renders basic torrent information
func (t *TorrentDetails) renderBasicInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Basic Information"))

	// Status with color
	status := t.getStatusDisplay(t.torrent.State)
	statusStyle := styles.GetStateStyle(t.torrent.State)
	lines = append(lines, fmt.Sprintf("Status: %s", statusStyle.Render(status)))

	// Progress
	progress := fmt.Sprintf("Progress: %.2f%%", t.torrent.Progress*100)
	lines = append(lines, progress)

	// Size information
	lines = append(lines, fmt.Sprintf("Size: %s", styles.FormatBytes(t.torrent.Size)))
	lines = append(lines, fmt.Sprintf("Downloaded: %s", styles.FormatBytes(t.torrent.Downloaded)))
	lines = append(lines, fmt.Sprintf("Uploaded: %s", styles.FormatBytes(t.torrent.Uploaded)))

	// Ratio
	lines = append(lines, fmt.Sprintf("Ratio: %.3f", t.torrent.Ratio))

	// Priority
	priorityStr := "Normal"
	if t.torrent.Priority > 0 {
		priorityStr = "High"
	} else if t.torrent.Priority < 0 {
		priorityStr = "Low"
	}
	lines = append(lines, fmt.Sprintf("Priority: %s", priorityStr))

	return strings.Join(lines, "\n")
}

// renderTransferInfo renders transfer-related information
func (t *TorrentDetails) renderTransferInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Transfer Information"))

	// Current speeds
	dlSpeed := styles.FormatSpeed(t.torrent.DlSpeed)
	upSpeed := styles.FormatSpeed(t.torrent.UpSpeed)
	lines = append(lines, fmt.Sprintf("Download Speed: %s", styles.DownloadingStyle.Render(dlSpeed)))
	lines = append(lines, fmt.Sprintf("Upload Speed: %s", styles.SeedingStyle.Render(upSpeed)))

	// ETA
	if t.torrent.ETA > 0 {
		eta := time.Duration(t.torrent.ETA) * time.Second
		lines = append(lines, fmt.Sprintf("ETA: %s", eta.String()))
	} else {
		lines = append(lines, "ETA: ∞")
	}

	// Peers and seeds
	lines = append(lines, fmt.Sprintf("Seeds: %d connected (%d total)", t.torrent.NumSeeds, t.torrent.NumComplete))
	lines = append(lines, fmt.Sprintf("Peers: %d connected (%d total)", t.torrent.NumLeeches, t.torrent.NumIncomplete))

	// Time active
	if t.torrent.TimeActive > 0 {
		activeTime := time.Duration(t.torrent.TimeActive) * time.Second
		lines = append(lines, fmt.Sprintf("Active Time: %s", activeTime.String()))
	}

	return strings.Join(lines, "\n")
}

// renderPathInfo renders path and file information
func (t *TorrentDetails) renderPathInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("File Information"))

	// Save path
	lines = append(lines, fmt.Sprintf("Save Path: %s", t.torrent.SavePath))

	// Tracker
	if t.torrent.Tracker != "" {
		lines = append(lines, fmt.Sprintf("Tracker: %s", t.torrent.Tracker))
	}

	// Category and tags
	if t.torrent.Category != "" {
		lines = append(lines, fmt.Sprintf("Category: %s", t.torrent.Category))
	}

	if t.torrent.Tags != "" {
		lines = append(lines, fmt.Sprintf("Tags: %s", t.torrent.Tags))
	}

	return strings.Join(lines, "\n")
}

// renderTechnicalInfo renders technical details
func (t *TorrentDetails) renderTechnicalInfo() string {
	var lines []string

	lines = append(lines, styles.SubtitleStyle.Render("Technical Details"))

	// Hash
	lines = append(lines, fmt.Sprintf("Hash: %s", t.torrent.Hash))

	// Added date
	if t.torrent.AddedOn > 0 {
		addedTime := time.Unix(t.torrent.AddedOn, 0)
		lines = append(lines, fmt.Sprintf("Added: %s", addedTime.Format("2006-01-02 15:04:05")))
	}

	// Completion date
	if t.torrent.CompletedOn > 0 {
		completionTime := time.Unix(t.torrent.CompletedOn, 0)
		lines = append(lines, fmt.Sprintf("Completed: %s", completionTime.Format("2006-01-02 15:04:05")))
	}

	// Remaining size
	if t.torrent.RemainingSize > 0 {
		remaining := styles.FormatBytes(t.torrent.RemainingSize)
		lines = append(lines, fmt.Sprintf("Remaining: %s", remaining))
	}

	// Max ratio and seeding time limits
	if t.torrent.MaxRatio > 0 {
		lines = append(lines, fmt.Sprintf("Max Ratio: %.2f", t.torrent.MaxRatio))
	}

	if t.torrent.MaxSeedingTime > 0 {
		maxSeedTime := time.Duration(t.torrent.MaxSeedingTime) * time.Second
		lines = append(lines, fmt.Sprintf("Max Seeding Time: %s", maxSeedTime.String()))
	}

	// Auto TMM
	if t.torrent.AutoTMM {
		lines = append(lines, styles.WarningStyle.Render("Auto Torrent Management: Enabled"))
	}

	return strings.Join(lines, "\n")
}

// getStatusDisplay returns a human-readable status
func (t *TorrentDetails) getStatusDisplay(state string) string {
	switch state {
	case "downloading":
		return "Downloading"
	case "metaDL":
		return "Downloading Metadata"
	case "forcedDL":
		return "Forced Download"
	case "allocating":
		return "Allocating Space"
	case "uploading":
		return "Seeding"
	case "forcedUP":
		return "Forced Seeding"
	case "stalledUP":
		return "Stalled (Seeding)"
	case "stalledDL":
		return "Stalled (Downloading)"
	case "pausedDL":
		return "Paused (Download)"
	case "pausedUP":
		return "Paused (Seeding)"
	case "queuedDL":
		return "Queued (Download)"
	case "queuedUP":
		return "Queued (Seeding)"
	case "error":
		return "Error"
	case "missingFiles":
		return "Missing Files"
	case "checkingUP":
		return "Checking (Seeding)"
	case "checkingDL":
		return "Checking (Download)"
	case "checkingResumeData":
		return "Checking Resume Data"
	default:
		return state
	}
}

// getMaxScroll calculates the maximum scroll value
func (t *TorrentDetails) getMaxScroll() int {
	if t.torrent == nil || t.height <= 0 {
		return 0
	}

	// This is a rough estimate - in practice you'd want to calculate
	// the actual number of lines in the rendered content
	totalSections := 5 // Basic, Transfer, Path, Technical + title
	avgLinesPerSection := 8
	totalLines := totalSections * avgLinesPerSection

	visibleLines := t.height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}

	return maxScroll
}

// SetDimensions updates the component dimensions
func (t *TorrentDetails) SetDimensions(width, height int) {
	t.width = width
	t.height = height
}

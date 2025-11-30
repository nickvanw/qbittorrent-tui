package terminal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
)

// TitleData contains all data available for terminal title templates
type TitleData struct {
	DlSpeed           int64  // Download speed in bytes/sec
	UpSpeed           int64  // Upload speed in bytes/sec
	SessionDownloaded int64  // Total downloaded this session
	SessionUploaded   int64  // Total uploaded this session
	ServerURL         string // Connected qBittorrent server URL
	ActiveTorrents    int    // Number of active torrents
	TotalTorrents     int    // Total number of torrents
	DlTorrents        int    // Number of downloading torrents
	UpTorrents        int    // Number of uploading torrents
	PausedTorrents    int    // Number of paused torrents
}

// SetTerminalTitle returns the ANSI escape sequence to set the terminal window title
// The sequence uses OSC 0 (Operating System Command 0) which sets both icon and window title
// Format: ESC ] 0 ; title BEL
func SetTerminalTitle(title string) string {
	return fmt.Sprintf("\033]0;%s\007", title)
}

// formatSpeed formats bytes per second into plain text human-readable format
// This is specifically for terminal titles and does not include any ANSI codes
func formatSpeed(bytesPerSec int64) string {
	return formatBytes(bytesPerSec) + "/s"
}

// formatBytes formats bytes into plain text human-readable format
// This is specifically for terminal titles and does not include any ANSI codes
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

// RenderTitle renders a title template with the provided data
// Template uses {variable} syntax for placeholders
func RenderTitle(template string, data TitleData) (string, error) {
	if template == "" {
		return "", fmt.Errorf("template is empty")
	}

	result := template

	// Define variable mappings using plain text formatting (no ANSI codes)
	variables := map[string]string{
		"{dl_speed}":           formatSpeed(data.DlSpeed),
		"{up_speed}":           formatSpeed(data.UpSpeed),
		"{session_downloaded}": formatBytes(data.SessionDownloaded),
		"{session_uploaded}":   formatBytes(data.SessionUploaded),
		"{server_url}":         data.ServerURL,
		"{active_torrents}":    fmt.Sprintf("%d", data.ActiveTorrents),
		"{total_torrents}":     fmt.Sprintf("%d", data.TotalTorrents),
		"{dl_torrents}":        fmt.Sprintf("%d", data.DlTorrents),
		"{up_torrents}":        fmt.Sprintf("%d", data.UpTorrents),
		"{paused_torrents}":    fmt.Sprintf("%d", data.PausedTorrents),
	}

	// Replace all variables
	for placeholder, value := range variables {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// ValidateTemplate validates a title template
// Returns an error if the template contains unknown variables
func ValidateTemplate(template string) error {
	if template == "" {
		return nil // Empty template is valid (will be disabled)
	}

	// Check for valid variables
	validVars := []string{
		"{dl_speed}", "{up_speed}", "{server_url}",
		"{active_torrents}", "{total_torrents}",
		"{dl_torrents}", "{up_torrents}", "{paused_torrents}",
		"{session_downloaded}", "{session_uploaded}",
	}

	// Find all {variable} patterns
	re := regexp.MustCompile(`\{[^}]+\}`)
	matches := re.FindAllString(template, -1)

	for _, match := range matches {
		valid := false
		for _, validVar := range validVars {
			if match == validVar {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unknown variable: %s (valid variables: %s)",
				match, strings.Join(validVars, ", "))
		}
	}

	return nil
}

// CalculateTorrentCounts calculates torrent counts from a torrent list
// Returns: active, downloading, uploading, paused counts
func CalculateTorrentCounts(torrents []api.Torrent) (active, downloading, uploading, paused int) {
	for _, t := range torrents {
		state := api.TorrentState(t.State)
		if state.IsActive() {
			active++
		}
		if state.IsDownloading() {
			downloading++
		}
		if state.IsUploading() {
			uploading++
		}
		if state.IsPaused() {
			paused++
		}
	}
	return
}

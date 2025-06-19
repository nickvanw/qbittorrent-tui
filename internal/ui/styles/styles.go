package styles

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Base colors
	PrimaryColor   = lipgloss.Color("#7B61FF")
	SecondaryColor = lipgloss.Color("#6366F1")
	AccentColor    = lipgloss.Color("#10B981")
	ErrorColor     = lipgloss.Color("#EF4444")
	WarningColor   = lipgloss.Color("#F59E0B")

	// Text colors
	TextColor       = lipgloss.Color("#E5E7EB")
	DimTextColor    = lipgloss.Color("#9CA3AF")
	BrightTextColor = lipgloss.Color("#F9FAFB")

	// Background colors
	BgColor      = lipgloss.Color("#111827")
	BgDarkColor  = lipgloss.Color("#0F172A")
	BgLightColor = lipgloss.Color("#1F2937")

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(DimTextColor).
			Padding(1, 2, 0, 2)

	FocusedPanelStyle = PanelStyle.
				BorderForeground(PrimaryColor)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(BrightTextColor).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	DimStyle = lipgloss.NewStyle().
			Foreground(DimTextColor)

	TextStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	// Status styles
	DownloadingStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true)

	SeedingStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	PausedStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	AccentStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)

	// Progress bar styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor)

	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(BgLightColor)

	// Table styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(BrightTextColor).
			Bold(true).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(DimTextColor)

	SelectedRowStyle = lipgloss.NewStyle().
				Background(BgLightColor).
				Foreground(BrightTextColor)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(DimTextColor).
			Padding(0, 1)

	FocusedInputStyle = InputStyle.
				BorderForeground(PrimaryColor)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(DimTextColor)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(TextColor)
)

// GetStateStyle returns the appropriate style for a torrent state
func GetStateStyle(state string) lipgloss.Style {
	switch state {
	case "downloading", "metaDL", "forcedDL", "allocating":
		return DownloadingStyle
	case "uploading", "forcedUP", "stalledUP":
		return SeedingStyle
	case "pausedDL", "pausedUP", "queuedDL", "queuedUP":
		return PausedStyle
	case "error", "missingFiles":
		return ErrorStyle
	default:
		return lipgloss.NewStyle()
	}
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return lipgloss.NewStyle().Render(fmt.Sprintf("%d B", bytes))
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return lipgloss.NewStyle().Render(fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp]))
}

// FormatSpeed formats bytes per second into human-readable format
func FormatSpeed(bytesPerSec int64) string {
	return FormatBytes(bytesPerSec) + "/s"
}

// TruncateString truncates a string to a maximum length with ellipsis
func TruncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		// For very short maxLen, just truncate to the length
		runes := []rune(s)
		if len(runes) <= maxLen {
			return s
		}
		return string(runes[:maxLen])
	}

	// Truncate and add ellipsis
	runes := []rune(s)
	if len(runes) <= maxLen-3 {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// FormatDuration formats seconds into human-readable duration
func FormatDuration(seconds int64) string {
	if seconds <= 0 {
		return "∞"
	}
	d := time.Duration(seconds) * time.Second
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}

// FormatTime formats Unix timestamp to human-readable time
func FormatTime(timestamp int64) string {
	if timestamp <= 0 {
		return "-"
	}
	t := time.Unix(timestamp, 0)
	now := time.Now()

	// If today, show time only
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}

	// If this year, show month and day
	if t.Year() == now.Year() {
		return t.Format("Jan 02")
	}

	// Otherwise show full date
	return t.Format("2006-01-02")
}

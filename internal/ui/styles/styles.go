package styles

import (
	"fmt"

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
			Padding(1, 2)

	FocusedPanelStyle = PanelStyle.Copy().
				BorderForeground(PrimaryColor)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(BrightTextColor).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	DimStyle = lipgloss.NewStyle().
			Foreground(DimTextColor)

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

	FocusedInputStyle = InputStyle.Copy().
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
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

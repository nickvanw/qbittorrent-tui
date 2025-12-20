package logger

import (
	"fmt"
	"strings"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
)

// LogSyncUpdate logs a sync API update with details
func LogSyncUpdate(syncData *api.SyncMainDataResponse) {
	if !isEnabled {
		return
	}

	// Summary line for the sync update
	Info("Sync update",
		"rid", syncData.RID,
		"full_update", syncData.FullUpdate,
		"torrents_updated", len(syncData.Torrents),
		"torrents_removed", len(syncData.TorrentsRemoved),
		"categories_updated", len(syncData.Categories),
		"categories_removed", len(syncData.CategoriesRemoved),
		"tags_added", len(syncData.Tags),
		"tags_removed", len(syncData.TagsRemoved),
	)

	// Detailed logging - one line per torrent with updated fields
	logTorrentDetails(syncData)
	logCategoryDetails(syncData)
	logTagDetails(syncData)
}

// logTorrentDetails logs one line per changed torrent with all updated field values
func logTorrentDetails(syncData *api.SyncMainDataResponse) {
	// One line per updated torrent with all field values
	for hash, partial := range syncData.Torrents {
		// Build attributes dynamically based on which fields are present
		attrs := []any{
			"hash", truncateHash(hash),
		}

		// Add all non-nil fields as structured attributes
		if partial.Name != nil {
			attrs = append(attrs, "name", *partial.Name)
		}
		if partial.State != nil {
			attrs = append(attrs, "state", *partial.State)
		}
		if partial.Progress != nil {
			attrs = append(attrs, "progress", fmt.Sprintf("%.2f%%", *partial.Progress*100))
		}
		if partial.DlSpeed != nil {
			attrs = append(attrs, "dlspeed", formatSpeed(*partial.DlSpeed))
		}
		if partial.UpSpeed != nil {
			attrs = append(attrs, "upspeed", formatSpeed(*partial.UpSpeed))
		}
		if partial.Size != nil {
			attrs = append(attrs, "size", formatBytes(*partial.Size))
		}
		if partial.Downloaded != nil {
			attrs = append(attrs, "downloaded", formatBytes(*partial.Downloaded))
		}
		if partial.Uploaded != nil {
			attrs = append(attrs, "uploaded", formatBytes(*partial.Uploaded))
		}
		if partial.Ratio != nil {
			attrs = append(attrs, "ratio", fmt.Sprintf("%.2f", *partial.Ratio))
		}
		if partial.ETA != nil {
			attrs = append(attrs, "eta", formatETA(*partial.ETA))
		}
		if partial.Category != nil {
			attrs = append(attrs, "category", *partial.Category)
		}
		if partial.Tags != nil {
			attrs = append(attrs, "tags", *partial.Tags)
		}
		if partial.NumSeeds != nil {
			attrs = append(attrs, "seeds", *partial.NumSeeds)
		}
		if partial.NumLeeches != nil {
			attrs = append(attrs, "leeches", *partial.NumLeeches)
		}

		Debug("Torrent updated", attrs...)
	}

	// One line for removed torrents
	if len(syncData.TorrentsRemoved) > 0 {
		Debug("Torrents removed",
			"count", len(syncData.TorrentsRemoved),
			"hashes", formatHashList(syncData.TorrentsRemoved),
		)
	}
}

// logCategoryDetails logs category changes (one line per category)
func logCategoryDetails(syncData *api.SyncMainDataResponse) {
	// One line per category change
	for name, cat := range syncData.Categories {
		Debug("Category updated",
			"name", name,
			"save_path", cat.SavePath,
			"download_path", cat.DownloadPath,
		)
	}

	// One line per removed category
	for _, name := range syncData.CategoriesRemoved {
		Debug("Category removed", "name", name)
	}
}

// logTagDetails logs tag changes (one line per tag)
func logTagDetails(syncData *api.SyncMainDataResponse) {
	// One line per added tag
	for _, tag := range syncData.Tags {
		Debug("Tag added", "tag", tag)
	}

	// One line per removed tag
	for _, tag := range syncData.TagsRemoved {
		Debug("Tag removed", "tag", tag)
	}
}

// truncateHash truncates a hash to first 8 characters for readability
func truncateHash(hash string) string {
	if len(hash) > 8 {
		return hash[:8] + "..."
	}
	return hash
}

// formatBytes formats bytes to a human-readable string
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatETA formats ETA in seconds to a human-readable string
func formatETA(seconds int64) string {
	if seconds < 0 {
		return "∞"
	}
	if seconds == 0 {
		return "0s"
	}

	const (
		minute = 60
		hour   = 60 * minute
		day    = 24 * hour
	)

	switch {
	case seconds >= day:
		return fmt.Sprintf("%dd %dh", seconds/day, (seconds%day)/hour)
	case seconds >= hour:
		return fmt.Sprintf("%dh %dm", seconds/hour, (seconds%hour)/minute)
	case seconds >= minute:
		return fmt.Sprintf("%dm %ds", seconds/minute, seconds%minute)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// formatSpeed formats a speed in bytes/sec to a human-readable string
func formatSpeed(bytesPerSec int64) string {
	if bytesPerSec == 0 {
		return "0 B/s"
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytesPerSec >= GB:
		return fmt.Sprintf("%.2f GB/s", float64(bytesPerSec)/GB)
	case bytesPerSec >= MB:
		return fmt.Sprintf("%.2f MB/s", float64(bytesPerSec)/MB)
	case bytesPerSec >= KB:
		return fmt.Sprintf("%.2f KB/s", float64(bytesPerSec)/KB)
	default:
		return fmt.Sprintf("%d B/s", bytesPerSec)
	}
}

// formatHashList formats a list of hashes for logging (truncates long lists)
func formatHashList(hashes []string) string {
	if len(hashes) == 0 {
		return "[]"
	}
	if len(hashes) <= 3 {
		// Show all hashes (truncated)
		truncated := make([]string, len(hashes))
		for i, h := range hashes {
			if len(h) > 8 {
				truncated[i] = h[:8] + "..."
			} else {
				truncated[i] = h
			}
		}
		return "[" + strings.Join(truncated, ", ") + "]"
	}
	// Show first 3 and count
	truncated := make([]string, 3)
	for i := 0; i < 3; i++ {
		if len(hashes[i]) > 8 {
			truncated[i] = hashes[i][:8] + "..."
		} else {
			truncated[i] = hashes[i]
		}
	}
	return fmt.Sprintf("[%s, ... (%d more)]", strings.Join(truncated, ", "), len(hashes)-3)
}

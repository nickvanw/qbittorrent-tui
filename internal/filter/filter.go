package filter

import (
	"net/url"
	"sort"
	"strings"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
)

type Filter struct {
	States   []string // downloading, seeding, paused, etc.
	Trackers []string // tracker domains
	Category string
	Tags     []string
	Search   string // text search
}

func (f *Filter) IsEmpty() bool {
	return len(f.States) == 0 &&
		len(f.Trackers) == 0 &&
		f.Category == "" &&
		len(f.Tags) == 0 &&
		f.Search == ""
}

func (f *Filter) Apply(torrents []api.Torrent) []api.Torrent {
	if f.IsEmpty() {
		return torrents
	}

	if torrents == nil {
		return []api.Torrent{}
	}

	filtered := make([]api.Torrent, 0)
	for _, t := range torrents {
		if f.matches(t) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (f *Filter) matches(t api.Torrent) bool {
	// State filter
	if len(f.States) > 0 && !f.matchesAnyState(t) {
		return false
	}

	// Tracker filter (by domain)
	if len(f.Trackers) > 0 {
		domain := extractDomain(t.Tracker)
		if !contains(f.Trackers, domain) {
			return false
		}
	}

	// Category filter
	if f.Category != "" && t.Category != f.Category {
		return false
	}

	// Tags filter
	if len(f.Tags) > 0 {
		torrentTags := splitTags(t.Tags)
		if !hasAnyTag(torrentTags, f.Tags) {
			return false
		}
	}

	// Search filter
	if f.Search != "" {
		searchLower := strings.ToLower(f.Search)
		nameLower := strings.ToLower(t.Name)
		if !strings.Contains(nameLower, searchLower) {
			return false
		}
	}

	return true
}

// matchesAnyState checks if torrent matches any of the selected states
func (f *Filter) matchesAnyState(t api.Torrent) bool {
	for _, state := range f.States {
		if f.matchesState(t, state) {
			return true
		}
	}
	return false
}

// matchesState checks if torrent matches a specific state (including logical states)
func (f *Filter) matchesState(t api.Torrent, state string) bool {
	switch state {
	case "active":
		// Active means actually transferring data - NOT stalled
		return t.State == "downloading" || t.State == "uploading" ||
			t.State == "metaDL" || t.State == "forcedDL" ||
			t.State == "forcedUP" || t.State == "allocating"
	case "paused":
		// Any paused state
		return api.TorrentState(t.State).IsPaused()
	case "completed":
		// Download completed - any state where the download is done and torrent is actively seeding
		// Excludes paused states (those should be filtered separately)
		return t.State == "uploading" || t.State == "stalledUP" ||
			t.State == "forcedUP" || t.State == "queuedUP"
	case "queued":
		// Any queued state
		return t.State == "queuedDL" || t.State == "queuedUP" || t.State == "queuedForChecking"
	case "stalled":
		// Any stalled state
		return t.State == "stalledDL" || t.State == "stalledUP"
	case "checking":
		// Any checking state
		return t.State == "checkingDL" || t.State == "checkingUP" ||
			t.State == "checkingResumeData" || t.State == "queuedForChecking"
	default:
		// Exact state match for specific states like "downloading", "uploading", etc.
		return t.State == state
	}
}

// ExtractUniqueTrackers gets unique tracker domains from torrents
func ExtractUniqueTrackers(torrents []api.Torrent) []string {
	domains := make(map[string]bool)
	for _, t := range torrents {
		domain := extractDomain(t.Tracker)
		if domain != "" {
			domains[domain] = true
		}
	}

	var result []string
	for d := range domains {
		result = append(result, d)
	}
	// Sort tracker domains alphabetically for stable display order
	sort.Strings(result)
	return result
}

// ExtractUniqueCategories gets unique categories from torrents
func ExtractUniqueCategories(torrents []api.Torrent) []string {
	categories := make(map[string]bool)
	for _, t := range torrents {
		if t.Category != "" {
			categories[t.Category] = true
		}
	}

	var result []string
	for c := range categories {
		result = append(result, c)
	}
	// Sort categories alphabetically for stable display order
	sort.Strings(result)
	return result
}

// ExtractUniqueTags gets unique tags from torrents
func ExtractUniqueTags(torrents []api.Torrent) []string {
	tags := make(map[string]bool)
	for _, t := range torrents {
		torrentTags := splitTags(t.Tags)
		for _, tag := range torrentTags {
			if tag != "" {
				tags[tag] = true
			}
		}
	}

	var result []string
	for tag := range tags {
		result = append(result, tag)
	}
	// Sort tags alphabetically for stable display order
	sort.Strings(result)
	return result
}

// ExtractUniqueStates gets unique states from torrents
func ExtractUniqueStates(torrents []api.Torrent) []string {
	states := make(map[string]bool)
	for _, t := range torrents {
		if t.State != "" {
			states[t.State] = true
		}
	}

	var result []string
	for s := range states {
		result = append(result, s)
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractDomain(tracker string) string {
	if tracker == "" {
		return ""
	}

	u, err := url.Parse(tracker)
	if err != nil {
		return ""
	}

	host := u.Host
	if host == "" {
		return ""
	}

	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host
}

func splitTags(tags string) []string {
	if tags == "" {
		return nil
	}

	parts := strings.Split(tags, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func hasAnyTag(torrentTags, filterTags []string) bool {
	for _, ft := range filterTags {
		for _, tt := range torrentTags {
			if tt == ft {
				return true
			}
		}
	}
	return false
}

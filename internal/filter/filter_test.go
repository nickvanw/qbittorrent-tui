package filter

import (
	"testing"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		filter   Filter
		expected bool
	}{
		{
			name:     "empty filter",
			filter:   Filter{},
			expected: true,
		},
		{
			name: "filter with states",
			filter: Filter{
				States: []string{"downloading"},
			},
			expected: false,
		},
		{
			name: "filter with trackers",
			filter: Filter{
				Trackers: []string{"tracker.example.com"},
			},
			expected: false,
		},
		{
			name: "filter with category",
			filter: Filter{
				Category: "movies",
			},
			expected: false,
		},
		{
			name: "filter with tags",
			filter: Filter{
				Tags: []string{"important"},
			},
			expected: false,
		},
		{
			name: "filter with search",
			filter: Filter{
				Search: "ubuntu",
			},
			expected: false,
		},
		{
			name: "filter with all fields",
			filter: Filter{
				States:   []string{"downloading"},
				Trackers: []string{"tracker.example.com"},
				Category: "movies",
				Tags:     []string{"important"},
				Search:   "ubuntu",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.filter.IsEmpty())
		})
	}
}

func TestFilterApply(t *testing.T) {
	torrents := []api.Torrent{
		{
			Hash:     "hash1",
			Name:     "Ubuntu 22.04 LTS",
			State:    "downloading",
			Category: "linux",
			Tags:     "important, os",
			Tracker:  "https://tracker.ubuntu.com:6969/announce",
		},
		{
			Hash:     "hash2",
			Name:     "Debian 11",
			State:    "seeding",
			Category: "linux",
			Tags:     "os",
			Tracker:  "https://tracker.debian.org:8080/announce",
		},
		{
			Hash:     "hash3",
			Name:     "Movie XYZ",
			State:    "pausedDL",
			Category: "movies",
			Tags:     "entertainment, hd",
			Tracker:  "https://tracker.example.com/announce",
		},
		{
			Hash:     "hash4",
			Name:     "Game ABC",
			State:    "downloading",
			Category: "games",
			Tags:     "",
			Tracker:  "https://tracker.gaming.com/announce",
		},
	}

	tests := []struct {
		name     string
		filter   Filter
		expected []string // expected hashes
	}{
		{
			name:     "empty filter returns all",
			filter:   Filter{},
			expected: []string{"hash1", "hash2", "hash3", "hash4"},
		},
		{
			name: "filter by state",
			filter: Filter{
				States: []string{"downloading"},
			},
			expected: []string{"hash1", "hash4"},
		},
		{
			name: "filter by multiple states",
			filter: Filter{
				States: []string{"downloading", "seeding"},
			},
			expected: []string{"hash1", "hash2", "hash4"},
		},
		{
			name: "filter by tracker domain",
			filter: Filter{
				Trackers: []string{"tracker.ubuntu.com"},
			},
			expected: []string{"hash1"},
		},
		{
			name: "filter by category",
			filter: Filter{
				Category: "linux",
			},
			expected: []string{"hash1", "hash2"},
		},
		{
			name: "filter by tag",
			filter: Filter{
				Tags: []string{"os"},
			},
			expected: []string{"hash1", "hash2"},
		},
		{
			name: "filter by multiple tags (OR operation)",
			filter: Filter{
				Tags: []string{"important", "hd"},
			},
			expected: []string{"hash1", "hash3"},
		},
		{
			name: "filter by search (case insensitive)",
			filter: Filter{
				Search: "ubuntu",
			},
			expected: []string{"hash1"},
		},
		{
			name: "filter by search partial match",
			filter: Filter{
				Search: "game",
			},
			expected: []string{"hash4"},
		},
		{
			name: "combined filters (AND operation)",
			filter: Filter{
				States:   []string{"downloading", "seeding"},
				Category: "linux",
			},
			expected: []string{"hash1", "hash2"},
		},
		{
			name: "complex filter",
			filter: Filter{
				States:   []string{"downloading"},
				Category: "linux",
				Tags:     []string{"important"},
			},
			expected: []string{"hash1"},
		},
		{
			name: "no match",
			filter: Filter{
				Search: "nonexistent",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Apply(torrents)

			resultHashes := make([]string, 0)
			for _, t := range result {
				resultHashes = append(resultHashes, t.Hash)
			}

			assert.Equal(t, tt.expected, resultHashes)
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		tracker  string
		expected string
	}{
		{
			name:     "empty tracker",
			tracker:  "",
			expected: "",
		},
		{
			name:     "http tracker with port",
			tracker:  "http://tracker.example.com:8080/announce",
			expected: "tracker.example.com",
		},
		{
			name:     "https tracker without port",
			tracker:  "https://tracker.example.com/announce",
			expected: "tracker.example.com",
		},
		{
			name:     "udp tracker",
			tracker:  "udp://tracker.example.com:6969",
			expected: "tracker.example.com",
		},
		{
			name:     "invalid url",
			tracker:  "not a url",
			expected: "",
		},
		{
			name:     "ip address with port",
			tracker:  "http://192.168.1.1:8080/announce",
			expected: "192.168.1.1",
		},
		{
			name:     "ipv6 address",
			tracker:  "http://[2001:db8::1]:8080/announce",
			expected: "[2001:db8::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.tracker)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     string
		expected []string
	}{
		{
			name:     "empty tags",
			tags:     "",
			expected: nil,
		},
		{
			name:     "single tag",
			tags:     "important",
			expected: []string{"important"},
		},
		{
			name:     "multiple tags",
			tags:     "important, os, linux",
			expected: []string{"important", "os", "linux"},
		},
		{
			name:     "tags with extra spaces",
			tags:     "  important  ,  os  ,  linux  ",
			expected: []string{"important", "os", "linux"},
		},
		{
			name:     "tags with empty parts",
			tags:     "important,,os,,",
			expected: []string{"important", "os"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitTags(tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasAnyTag(t *testing.T) {
	tests := []struct {
		name        string
		torrentTags []string
		filterTags  []string
		expected    bool
	}{
		{
			name:        "empty both",
			torrentTags: []string{},
			filterTags:  []string{},
			expected:    false,
		},
		{
			name:        "empty torrent tags",
			torrentTags: []string{},
			filterTags:  []string{"important"},
			expected:    false,
		},
		{
			name:        "empty filter tags",
			torrentTags: []string{"important"},
			filterTags:  []string{},
			expected:    false,
		},
		{
			name:        "exact match",
			torrentTags: []string{"important", "os"},
			filterTags:  []string{"important"},
			expected:    true,
		},
		{
			name:        "multiple matches",
			torrentTags: []string{"important", "os", "linux"},
			filterTags:  []string{"os", "windows"},
			expected:    true,
		},
		{
			name:        "no match",
			torrentTags: []string{"important", "os"},
			filterTags:  []string{"entertainment", "hd"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasAnyTag(tt.torrentTags, tt.filterTags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractUniqueTrackers(t *testing.T) {
	torrents := []api.Torrent{
		{Tracker: "https://tracker1.example.com:8080/announce"},
		{Tracker: "https://tracker2.example.com/announce"},
		{Tracker: "https://tracker1.example.com:9999/announce"}, // same domain, different port
		{Tracker: "udp://tracker3.example.com:6969"},
		{Tracker: ""}, // empty tracker
		{Tracker: "invalid url"},
	}

	result := ExtractUniqueTrackers(torrents)

	// Sort for consistent test results
	expected := []string{"tracker1.example.com", "tracker2.example.com", "tracker3.example.com"}

	assert.Len(t, result, 3)
	for _, exp := range expected {
		assert.Contains(t, result, exp)
	}
}

func TestExtractUniqueCategories(t *testing.T) {
	torrents := []api.Torrent{
		{Category: "movies"},
		{Category: "linux"},
		{Category: "movies"}, // duplicate
		{Category: "games"},
		{Category: ""}, // empty category
	}

	result := ExtractUniqueCategories(torrents)

	expected := []string{"movies", "linux", "games"}

	assert.Len(t, result, 3)
	for _, exp := range expected {
		assert.Contains(t, result, exp)
	}
}

func TestExtractUniqueTags(t *testing.T) {
	torrents := []api.Torrent{
		{Tags: "important, os"},
		{Tags: "os, linux"}, // os is duplicate
		{Tags: "entertainment"},
		{Tags: ""},          // empty tags
		{Tags: "important"}, // duplicate
	}

	result := ExtractUniqueTags(torrents)

	expected := []string{"important", "os", "linux", "entertainment"}

	assert.Len(t, result, 4)
	for _, exp := range expected {
		assert.Contains(t, result, exp)
	}
}

func TestExtractUniqueStates(t *testing.T) {
	torrents := []api.Torrent{
		{State: "downloading"},
		{State: "seeding"},
		{State: "downloading"}, // duplicate
		{State: "pausedDL"},
		{State: ""}, // empty state
	}

	result := ExtractUniqueStates(torrents)

	expected := []string{"downloading", "seeding", "pausedDL"}

	assert.Len(t, result, 3)
	for _, exp := range expected {
		assert.Contains(t, result, exp)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "test",
			expected: false,
		},
		{
			name:     "item exists",
			slice:    []string{"one", "two", "three"},
			item:     "two",
			expected: true,
		},
		{
			name:     "item not exists",
			slice:    []string{"one", "two", "three"},
			item:     "four",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterEdgeCases(t *testing.T) {
	t.Run("nil torrents", func(t *testing.T) {
		f := Filter{States: []string{"downloading"}}
		result := f.Apply(nil)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("empty torrents", func(t *testing.T) {
		f := Filter{States: []string{"downloading"}}
		result := f.Apply([]api.Torrent{})
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("torrent with no tracker", func(t *testing.T) {
		torrents := []api.Torrent{
			{
				Hash:    "hash1",
				Name:    "Test",
				State:   "downloading",
				Tracker: "",
			},
		}

		f := Filter{Trackers: []string{"example.com"}}
		result := f.Apply(torrents)
		assert.Empty(t, result)
	})

	t.Run("case sensitivity in state filter", func(t *testing.T) {
		torrents := []api.Torrent{
			{
				Hash:  "hash1",
				Name:  "Test",
				State: "downloading",
			},
		}

		// State filter should be case-sensitive
		f := Filter{States: []string{"Downloading"}}
		result := f.Apply(torrents)
		assert.Empty(t, result)
	})
}

func TestFilterComplexScenarios(t *testing.T) {
	torrents := []api.Torrent{
		{
			Hash:     "hash1",
			Name:     "Big Buck Bunny 4K",
			State:    "downloading",
			Category: "movies",
			Tags:     "4k, animation",
			Tracker:  "https://public.tracker.com:8080/announce",
		},
		{
			Hash:     "hash2",
			Name:     "Sintel 1080p",
			State:    "seeding",
			Category: "movies",
			Tags:     "1080p, animation",
			Tracker:  "https://public.tracker.com:8080/announce",
		},
		{
			Hash:     "hash3",
			Name:     "Tears of Steel",
			State:    "downloading",
			Category: "movies",
			Tags:     "1080p, scifi",
			Tracker:  "https://private.tracker.org/announce",
		},
	}

	t.Run("filter animation movies being downloaded", func(t *testing.T) {
		f := Filter{
			States:   []string{"downloading"},
			Category: "movies",
			Tags:     []string{"animation"},
		}
		result := f.Apply(torrents)
		require.Len(t, result, 1)
		assert.Equal(t, "hash1", result[0].Hash)
	})

	t.Run("filter by tracker and resolution", func(t *testing.T) {
		f := Filter{
			Trackers: []string{"public.tracker.com"},
			Tags:     []string{"1080p"},
		}
		result := f.Apply(torrents)
		require.Len(t, result, 1)
		assert.Equal(t, "hash2", result[0].Hash)
	})

	t.Run("search with state filter", func(t *testing.T) {
		f := Filter{
			Search: "steel",
			States: []string{"downloading", "queued"},
		}
		result := f.Apply(torrents)
		require.Len(t, result, 1)
		assert.Equal(t, "hash3", result[0].Hash)
	})
}

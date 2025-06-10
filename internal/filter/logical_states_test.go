package filter

import (
	"testing"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestLogicalStates(t *testing.T) {
	torrents := []api.Torrent{
		{Name: "Active Download", State: "downloading"},
		{Name: "Active Upload", State: "uploading"},
		{Name: "Paused Download", State: "pausedDL"},
		{Name: "Paused Upload", State: "pausedUP"},
		{Name: "Queued Download", State: "queuedDL"},
		{Name: "Queued Upload", State: "queuedUP"},
		{Name: "Stalled Download", State: "stalledDL"},
		{Name: "Stalled Upload", State: "stalledUP"},
		{Name: "Checking Files", State: "checkingDL"},
		{Name: "Error Torrent", State: "error"},
		{Name: "Allocating Space", State: "allocating"},
		{Name: "Downloading Metadata", State: "metaDL"},
		{Name: "Moving Files", State: "moving"},
	}

	tests := []struct {
		name          string
		filterStates  []string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "Active filter should include downloading and uploading",
			filterStates:  []string{"active"},
			expectedCount: 4, // downloading, uploading, allocating, metaDL (NO stalledUP)
			expectedNames: []string{"Active Download", "Active Upload", "Allocating Space", "Downloading Metadata"},
		},
		{
			name:          "Paused filter should include all paused states",
			filterStates:  []string{"paused"},
			expectedCount: 2, // pausedDL, pausedUP
			expectedNames: []string{"Paused Download", "Paused Upload"},
		},
		{
			name:          "Completed filter should include upload states",
			filterStates:  []string{"completed"},
			expectedCount: 3, // uploading, queuedUP, stalledUP (pausedUP is paused, not completed)
			expectedNames: []string{"Active Upload", "Queued Upload", "Stalled Upload"},
		},
		{
			name:          "Queued filter should include all queued states",
			filterStates:  []string{"queued"},
			expectedCount: 2, // queuedDL, queuedUP
			expectedNames: []string{"Queued Download", "Queued Upload"},
		},
		{
			name:          "Stalled filter should include all stalled states",
			filterStates:  []string{"stalled"},
			expectedCount: 2, // stalledDL, stalledUP
			expectedNames: []string{"Stalled Download", "Stalled Upload"},
		},
		{
			name:          "Checking filter should include checking states",
			filterStates:  []string{"checking"},
			expectedCount: 1, // checkingDL
			expectedNames: []string{"Checking Files"},
		},
		{
			name:          "Exact state match should still work",
			filterStates:  []string{"downloading"},
			expectedCount: 1,
			expectedNames: []string{"Active Download"},
		},
		{
			name:          "Multiple logical states should work",
			filterStates:  []string{"active", "paused"},
			expectedCount: 6, // 4 active + 2 paused
		},
		{
			name:          "Mix of logical and exact states",
			filterStates:  []string{"active", "error"},
			expectedCount: 5, // 4 active + 1 error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := Filter{States: tt.filterStates}
			result := filter.Apply(torrents)

			assert.Equal(t, tt.expectedCount, len(result), "Unexpected number of filtered torrents")

			if tt.expectedNames != nil {
				var actualNames []string
				for _, torrent := range result {
					actualNames = append(actualNames, torrent.Name)
				}

				for _, expectedName := range tt.expectedNames {
					assert.Contains(t, actualNames, expectedName, "Expected torrent not found in results")
				}
			}
		})
	}
}

func TestLogicalStateMatching(t *testing.T) {
	filter := Filter{}

	tests := []struct {
		name     string
		state    string
		torrent  api.Torrent
		expected bool
	}{
		{
			name:     "active matches downloading",
			state:    "active",
			torrent:  api.Torrent{State: "downloading"},
			expected: true,
		},
		{
			name:     "active matches uploading",
			state:    "active",
			torrent:  api.Torrent{State: "uploading"},
			expected: true,
		},
		{
			name:     "active does not match paused",
			state:    "active",
			torrent:  api.Torrent{State: "pausedDL"},
			expected: false,
		},
		{
			name:     "active does not match stalled",
			state:    "active",
			torrent:  api.Torrent{State: "stalledUP"},
			expected: false,
		},
		{
			name:     "paused matches pausedDL",
			state:    "paused",
			torrent:  api.Torrent{State: "pausedDL"},
			expected: true,
		},
		{
			name:     "paused matches pausedUP",
			state:    "paused",
			torrent:  api.Torrent{State: "pausedUP"},
			expected: true,
		},
		{
			name:     "completed matches uploading",
			state:    "completed",
			torrent:  api.Torrent{State: "uploading"},
			expected: true,
		},
		{
			name:     "completed matches stalledUP",
			state:    "completed",
			torrent:  api.Torrent{State: "stalledUP"},
			expected: true,
		},
		{
			name:     "completed does not match downloading",
			state:    "completed",
			torrent:  api.Torrent{State: "downloading"},
			expected: false,
		},
		{
			name:     "exact state match still works",
			state:    "downloading",
			torrent:  api.Torrent{State: "downloading"},
			expected: true,
		},
		{
			name:     "exact state mismatch",
			state:    "downloading",
			torrent:  api.Torrent{State: "uploading"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.matchesState(tt.torrent, tt.state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

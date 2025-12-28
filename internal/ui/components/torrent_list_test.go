package components

import (
	"strings"
	"testing"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
)

func TestResponsiveLayout(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		expectedCols  int
		priorityOrder []string
	}{
		{
			name:          "Very narrow (80 columns)",
			width:         80,
			expectedCols:  4, // name, progress, status should fit
			priorityOrder: []string{"name", "progress", "status"},
		},
		{
			name:          "Medium width (120 columns)",
			width:         120,
			expectedCols:  6, // Should fit more columns
			priorityOrder: []string{"name", "progress", "status"},
		},
		{
			name:          "Wide screen (200 columns)",
			width:         200,
			expectedCols:  9, // Should fit all columns
			priorityOrder: []string{"name", "progress", "status"},
		},
	}

	// Debug output showing responsive behavior
	t.Logf("=== Responsive Layout Test Results ===")
	testWidths := []int{50, 80, 120, 160, 200, 300}

	for _, width := range testWidths {
		torrentList := NewTorrentList()
		torrentList.SetDimensions(width, 20)
		columns := torrentList.GetColumns()

		totalWidth := 0
		for i, col := range columns {
			totalWidth += col.Width
			if i < len(columns)-1 {
				totalWidth += 1 // spacing
			}
		}

		t.Logf("Width %d: %d columns, %d/%d chars used (%.1f%%)",
			width, len(columns), totalWidth, width, float64(totalWidth)/float64(width)*100)

		for _, col := range columns {
			t.Logf("  %-8s: %d chars (min: %d, flex: %.1f)",
				col.Config.Key, col.Width, col.Config.MinWidth, col.Config.FlexGrow)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrentList := NewTorrentList()
			torrentList.SetDimensions(tt.width, 20)

			// Check that columns were calculated
			if len(torrentList.columns) == 0 {
				t.Error("No columns calculated")
				return
			}

			// Check that we don't exceed available width
			totalWidth := 0
			for i, col := range torrentList.columns {
				totalWidth += col.Width
				if i < len(torrentList.columns)-1 {
					totalWidth += 1 // spacing between columns
				}
			}

			if totalWidth > tt.width {
				t.Errorf("Total width %d exceeds available width %d", totalWidth, tt.width)
			}

			// Check that highest priority columns are included
			for _, expectedKey := range tt.priorityOrder {
				found := false
				for _, col := range torrentList.columns {
					if col.Config.Key == expectedKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected high-priority column %s not found", expectedKey)
				}
			}

			// Check that each column has a reasonable width
			for _, col := range torrentList.columns {
				if col.Width < col.Config.MinWidth {
					t.Errorf("Column %s width %d is less than minimum %d",
						col.Config.Key, col.Width, col.Config.MinWidth)
				}
				if col.Config.MaxWidth > 0 && col.Width > col.Config.MaxWidth {
					t.Errorf("Column %s width %d exceeds maximum %d",
						col.Config.Key, col.Width, col.Config.MaxWidth)
				}
			}
		})
	}
}

func TestColumnPriority(t *testing.T) {
	torrentList := NewTorrentList()

	// Test very narrow width - should only show highest priority columns
	torrentList.SetDimensions(50, 20)

	if len(torrentList.columns) == 0 {
		t.Error("No columns calculated for narrow width")
		return
	}

	// Should have at least name column (priority 1)
	hasName := false
	for _, col := range torrentList.columns {
		if col.Config.Key == "name" {
			hasName = true
			break
		}
	}

	if !hasName {
		t.Error("Name column should always be visible")
	}
}

func TestFlexGrow(t *testing.T) {
	torrentList := NewTorrentList()

	// Test with wide width to see flex grow in action
	torrentList.SetDimensions(300, 20)

	if len(torrentList.columns) == 0 {
		t.Error("No columns calculated")
		return
	}

	// Find name column - it should be larger than minimum due to flex grow
	for _, col := range torrentList.columns {
		if col.Config.Key == "name" && col.Config.FlexGrow > 0 {
			if col.Width <= col.Config.MinWidth {
				t.Errorf("Name column with flex grow should be larger than minimum width. Got %d, min %d",
					col.Width, col.Config.MinWidth)
			}
			break
		}
	}
}

func TestTorrentSorting(t *testing.T) {
	torrentList := NewTorrentList()

	// Create test torrents with different values
	testTorrents := []api.Torrent{
		{Name: "Zebra", Size: 1000, Progress: 0.5, State: "downloading", DlSpeed: 500, UpSpeed: 100, Ratio: 1.5},
		{Name: "Alpha", Size: 2000, Progress: 0.8, State: "uploading", DlSpeed: 200, UpSpeed: 300, Ratio: 2.0},
		{Name: "Beta", Size: 500, Progress: 1.0, State: "completed", DlSpeed: 0, UpSpeed: 400, Ratio: 0.5},
	}

	torrentList.SetTorrents(testTorrents)

	// Test default sorting (by name ascending)
	if torrentList.torrents[0].Name != "Alpha" {
		t.Errorf("Expected first torrent to be Alpha, got %s", torrentList.torrents[0].Name)
	}

	// Test sorting by size
	torrentList.setSortColumn("size", false)
	if torrentList.torrents[0].Size != 500 {
		t.Errorf("Expected smallest torrent first, got size %d", torrentList.torrents[0].Size)
	}

	// Test reverse size sorting
	torrentList.setSortColumn("size", true)
	if torrentList.torrents[0].Size != 2000 {
		t.Errorf("Expected largest torrent first, got size %d", torrentList.torrents[0].Size)
	}

	// Test sorting by progress
	torrentList.setSortColumn("progress", false)
	if torrentList.torrents[0].Progress != 0.5 {
		t.Errorf("Expected lowest progress first, got %.1f", torrentList.torrents[0].Progress)
	}

	// Test sorting by state
	torrentList.setSortColumn("status", false)
	if torrentList.torrents[0].State != "completed" {
		t.Errorf("Expected completed state first, got %s", torrentList.torrents[0].State)
	}
}

func TestSortIndicators(t *testing.T) {
	torrentList := NewTorrentList()
	torrentList.SetDimensions(200, 20)

	// Set a specific sort
	torrentList.setSortColumn("size", false)

	// Get the header
	header := torrentList.renderHeader()

	// Check that size column has ascending indicator
	if !strings.Contains(header, "Size ↑") {
		t.Error("Expected ascending sort indicator for Size column")
	}

	// Change to descending
	torrentList.setSortColumn("size", true)
	header = torrentList.renderHeader()

	// Check that size column has descending indicator
	if !strings.Contains(header, "Size ↓") {
		t.Error("Expected descending sort indicator for Size column")
	}
}

func TestSortConfigPersistence(t *testing.T) {
	torrentList := NewTorrentList()

	// Set a custom sort
	torrentList.setSortColumn("progress", true)

	// Get config
	config := torrentList.GetSortConfig()
	if config.Column != "progress" || config.Direction != SortDesc {
		t.Error("Sort config not properly stored")
	}

	// Create new torrent list and apply saved config
	newTorrentList := NewTorrentList()
	newTorrentList.SetSortConfig(config)

	newConfig := newTorrentList.GetSortConfig()
	if newConfig.Column != "progress" || newConfig.Direction != SortDesc {
		t.Error("Sort config not properly restored")
	}
}

func TestSelectionPreservedOnRefresh(t *testing.T) {
	torrentList := NewTorrentList()

	torrents := []api.Torrent{
		{Name: "Alpha", Hash: "a", Size: 100},
		{Name: "Beta", Hash: "b", Size: 200},
		{Name: "Gamma", Hash: "c", Size: 300},
	}
	torrentList.SetTorrents(torrents)
	torrentList.moveDown() // select "Beta"

	if torrentList.selectedHash != "b" {
		t.Fatalf("expected selected hash b, got %s", torrentList.selectedHash)
	}

	// Change ordering by size to ensure selection is preserved by hash, not index.
	updated := []api.Torrent{
		{Name: "Gamma", Hash: "c", Size: 300},
		{Name: "Alpha", Hash: "a", Size: 100},
		{Name: "Beta", Hash: "b", Size: 200},
	}
	torrentList.SetTorrents(updated)

	if torrentList.selectedHash != "b" {
		t.Fatalf("expected selected hash to remain b, got %s", torrentList.selectedHash)
	}
	if torrentList.torrents[torrentList.cursor].Hash != "b" {
		t.Fatalf("expected cursor to point at b, got %s", torrentList.torrents[torrentList.cursor].Hash)
	}
}

func TestSelectionClearedOnEmptyList(t *testing.T) {
	torrentList := NewTorrentList()

	torrentList.SetTorrents([]api.Torrent{
		{Name: "Alpha", Hash: "a"},
	})
	if torrentList.selectedHash == "" {
		t.Fatal("expected selected hash to be set")
	}

	torrentList.SetTorrents([]api.Torrent{})
	if torrentList.selectedHash != "" {
		t.Fatalf("expected selected hash to be cleared, got %s", torrentList.selectedHash)
	}
	if torrentList.cursor != 0 {
		t.Fatalf("expected cursor reset to 0, got %d", torrentList.cursor)
	}
}

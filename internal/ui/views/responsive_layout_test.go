package views

import (
	"fmt"
	"strings"
	"testing"

	// Removed unused tea import
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/testutil"
	"github.com/stretchr/testify/assert"
)

// TestResponsiveLayout_TerminalSizes tests UI adaptation to different terminal sizes
func TestResponsiveLayout_TerminalSizes(t *testing.T) {
	// Create test torrents with various data
	createTestTorrentsWithData := func() []api.Torrent {
		return []api.Torrent{
			{
				Name:       "Ubuntu 22.04 Desktop ISO",
				Size:       3825205248, // ~3.6GB
				Progress:   0.75,
				State:      "downloading",
				DlSpeed:    5242880, // 5MB/s
				UpSpeed:    1048576, // 1MB/s
				NumSeeds:   42,
				NumLeeches: 15,
				Ratio:      1.5,
				ETA:        3600, // 1 hour
				Category:   "Software",
				Tags:       "linux,iso",
				Tracker:    "tracker.ubuntu.com",
			},
			{
				Name:       "Very Long Torrent Name That Should Be Truncated In Narrow Views",
				Size:       1073741824, // 1GB
				Progress:   1.0,
				State:      "seeding",
				DlSpeed:    0,
				UpSpeed:    2097152, // 2MB/s
				NumSeeds:   10,
				NumLeeches: 25,
				Ratio:      3.2,
				Category:   "Media",
				Tags:       "video,hd",
			},
		}
	}

	terminalSizes := []struct {
		name   string
		width  int
		height int
	}{
		{"very narrow terminal", 60, 20},
		{"narrow terminal", 80, 24},
		{"standard terminal", 120, 40},
		{"wide terminal", 160, 50},
		{"ultra-wide terminal", 200, 60},
		{"minimal terminal", 40, 10},
	}

	for _, size := range terminalSizes {
		t.Run(size.name, func(t *testing.T) {
			view, mockClient := createTestMainView()
			torrents := createTestTorrentsWithData()
			mockClient.Torrents = torrents
			view.torrents = torrents
			view.torrentList.SetTorrents(torrents)

			// Resize the terminal
			model := testutil.SimulateResize(t, view, size.width, size.height)
			mainView := model.(*MainView)

			// Verify dimensions were updated
			assert.Equal(t, size.width, mainView.width)
			assert.Equal(t, size.height, mainView.height)

			// Get the rendered view
			output := mainView.View()
			lines := testutil.GetViewLines(mainView)

			// All lines should fit within reasonable bounds (allowing for border/padding)
			testutil.AssertLineWidth(t, mainView, size.width+2)

			// Should have appropriate number of lines for height (allowing for some overflow)
			assert.LessOrEqual(t, len(lines), size.height+3)

			// Check specific adaptations based on width
			if size.width < 80 {
				// In narrow view, some columns should be hidden
				headerLine := findHeaderLine(lines)
				if headerLine != "" {
					// Should not show all columns in narrow view
					assert.NotContains(t, headerLine, "Tracker", "Tracker column should be hidden in narrow view")
					assert.NotContains(t, headerLine, "Added", "Added column should be hidden in narrow view")
				}
			} else if size.width >= 160 {
				// In wide view, more columns should be visible
				headerLine := findHeaderLine(lines)
				if headerLine != "" && len(torrents) > 0 {
					// Wide terminals can show more columns
					assert.Contains(t, output, "MB/s", "Speed columns should show in wide view")
				}
			}

			// Minimum content should always be visible
			if len(torrents) > 0 {
				// At least torrent names should be visible (possibly truncated)
				assert.Contains(t, output, "Ubuntu", "Should show at least part of torrent names")
			}
		})
	}
}

// TestResponsiveLayout_ColumnDropping tests that columns are dropped appropriately at narrow widths
func TestResponsiveLayout_ColumnDropping(t *testing.T) {
	view, mockClient := createTestMainView()

	// Create torrents with all data fields populated
	torrents := []api.Torrent{
		{
			Name:       "Test Torrent with All Fields",
			Size:       1000000000,
			Progress:   0.5,
			State:      "downloading",
			DlSpeed:    1000000,
			UpSpeed:    500000,
			NumSeeds:   20,
			NumLeeches: 10,
			Ratio:      2.5,
			ETA:        1800,
			Category:   "Test",
			Tags:       "test,sample",
			Tracker:    "test.tracker.com",
			AddedOn:    1609459200, // 2021-01-01
		},
	}

	mockClient.Torrents = torrents
	view.torrents = torrents
	view.torrentList.SetTorrents(torrents)

	// Test column visibility at different widths
	widthTests := []struct {
		width           int
		expectedColumns []string
		hiddenColumns   []string
	}{
		{
			width:           200,
			expectedColumns: []string{"Name", "Size", "Progress", "Status", "Down", "Up"},
			hiddenColumns:   []string{},
		},
		{
			width:           120,
			expectedColumns: []string{"Name", "Size", "Progress", "Status"},
			hiddenColumns:   []string{"Tracker", "Added"}, // Some columns might be hidden
		},
		{
			width:           80,
			expectedColumns: []string{"Name", "Size", "Progress"},
			hiddenColumns:   []string{"Seeds", "Peers", "Ratio", "Category", "Tags"},
		},
		{
			width:           60,
			expectedColumns: []string{"Name"}, // Only essential columns
			hiddenColumns:   []string{"Down", "Up", "ETA", "Tracker"},
		},
	}

	for _, test := range widthTests {
		t.Run(fmt.Sprintf("width_%d", test.width), func(t *testing.T) {
			// Resize to test width
			model := testutil.SimulateResize(t, view, test.width, 30)
			mainView := model.(*MainView)

			// Get the torrent list view
			torrentListView := mainView.torrentList.View()
			lines := strings.Split(torrentListView, "\n")

			// Find header line
			headerLine := ""
			for _, line := range lines {
				cleaned := testutil.StripANSI(line)
				if strings.Contains(cleaned, "Name") {
					headerLine = cleaned
					break
				}
			}

			// Check expected columns are visible
			for _, col := range test.expectedColumns {
				if headerLine != "" && col == "Progress" {
					// Progress might be shown as %
					assert.True(t,
						strings.Contains(headerLine, col) || strings.Contains(headerLine, "%"),
						"Expected column %s to be visible at width %d", col, test.width)
				} else if headerLine != "" {
					assert.Contains(t, headerLine, col,
						"Expected column %s to be visible at width %d", col, test.width)
				}
			}

			// Check hidden columns are not visible
			for _, col := range test.hiddenColumns {
				if headerLine != "" {
					assert.NotContains(t, headerLine, col,
						"Expected column %s to be hidden at width %d", col, test.width)
				}
			}
		})
	}
}

// TestResponsiveLayout_TextTruncation tests that text is properly truncated in narrow columns
func TestResponsiveLayout_TextTruncation(t *testing.T) {
	view, mockClient := createTestMainView()

	// Create a torrent with very long text fields
	longName := "This is an extremely long torrent name that should definitely be truncated in any reasonable terminal width to prevent line wrapping"
	torrents := []api.Torrent{
		{
			Name:     longName,
			Size:     999999999999, // Very large size
			Category: "VeryLongCategoryNameThatShouldBeTruncated",
			Tags:     "tag1,tag2,tag3,tag4,tag5,tag6,tag7,tag8,tag9,tag10",
			Tracker:  "https://very-long-tracker-domain-name-that-exceeds-column-width.example.com:8080/announce",
		},
	}

	mockClient.Torrents = torrents
	view.torrents = torrents
	view.torrentList.SetTorrents(torrents)

	widths := []int{60, 80, 120}

	for _, width := range widths {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			// Resize terminal
			model := testutil.SimulateResize(t, view, width, 20)
			mainView := model.(*MainView)

			lines := testutil.GetViewLines(mainView)

			// No line should exceed reasonable bounds (allowing for border/padding)
			testutil.AssertLineWidth(t, mainView, width+2)

			// Long torrent name should be truncated with ellipsis
			if width < 120 {
				// Should see truncation indicator
				contentLines := findContentLines(lines)
				for _, line := range contentLines {
					if strings.Contains(line, "This is an extremely long") {
						assert.Contains(t, line, "...",
							"Long text should be truncated with ellipsis at width %d", width)
						assert.NotContains(t, line, longName,
							"Full long name should not appear at width %d", width)
					}
				}
			}
		})
	}
}

// TestResponsiveLayout_PanelSizing tests that panels resize appropriately
func TestResponsiveLayout_PanelSizing(t *testing.T) {
	view, _ := createTestMainView()

	// Test different terminal sizes and verify basic layout properties
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "standard terminal",
			width:  120,
			height: 40,
		},
		{
			name:   "tall terminal",
			width:  100,
			height: 60,
		},
		{
			name:   "short terminal",
			width:  120,
			height: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Resize terminal
			model := testutil.SimulateResize(t, view, tt.width, tt.height)
			mainView := model.(*MainView)

			// The updateDimensions method should have set appropriate sizes
			output := mainView.View()
			lines := strings.Split(output, "\n")

			// Should use reasonable amount of terminal height (not exceed it)
			assert.LessOrEqual(t, len(lines), tt.height+5) // Allow some flexibility

			// Should contain all key sections
			assert.Contains(t, output, "Loading statistics", "Should show stats panel")
			assert.Contains(t, output, "No active filters", "Should show filter panel")
			assert.Contains(t, output, "help", "Should show help")

			// Basic sanity check that we have a reasonable layout
			assert.Greater(t, len(lines), 5, "Should have multiple lines of output")
		})
	}
}

// TestResponsiveLayout_EmptyStates tests layout with no data
func TestResponsiveLayout_EmptyStates(t *testing.T) {
	view, mockClient := createTestMainView()

	// Clear all data
	mockClient.Torrents = []api.Torrent{}
	mockClient.GlobalStats = &api.GlobalStats{}
	view.torrents = []api.Torrent{}
	view.torrentList.SetTorrents([]api.Torrent{})

	sizes := []struct {
		width  int
		height int
	}{
		{60, 20},
		{120, 40},
		{200, 50},
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			model := testutil.SimulateResize(t, view, size.width, size.height)
			mainView := model.(*MainView)

			output := mainView.View()

			// Should show empty state message
			assert.Contains(t, output, "No torrents")

			// Should still show stats panel (even if loading)
			assert.Contains(t, output, "Loading statistics")

			// Should show filter panel
			assert.Contains(t, output, "No active filters")

			// All content should fit within reasonable bounds (allowing for border/padding)
			testutil.AssertLineWidth(t, mainView, size.width+2)
		})
	}
}

// Helper functions

func findHeaderLine(lines []string) string {
	for _, line := range lines {
		cleaned := testutil.StripANSI(line)
		if strings.Contains(cleaned, "Name") &&
			(strings.Contains(cleaned, "Size") || strings.Contains(cleaned, "Progress")) {
			return cleaned
		}
	}
	return ""
}

func findContentLines(lines []string) []string {
	var contentLines []string
	inContent := false

	for _, line := range lines {
		cleaned := testutil.StripANSI(line)

		// Start collecting after header
		if strings.Contains(cleaned, "Name") &&
			(strings.Contains(cleaned, "Size") || strings.Contains(cleaned, "Progress")) {
			inContent = true
			continue
		}

		// Stop at filter panel
		if strings.Contains(cleaned, "No active filters") || strings.Contains(cleaned, "Press:") {
			break
		}

		if inContent && cleaned != "" && !strings.HasPrefix(cleaned, "│") {
			contentLines = append(contentLines, cleaned)
		}
	}

	return contentLines
}

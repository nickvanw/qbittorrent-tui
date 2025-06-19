package components

import (
	"testing"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/stretchr/testify/assert"
)

// TestStatsPanel_Creation tests stats panel initialization
func TestStatsPanel_Creation(t *testing.T) {
	panel := NewStatsPanel()

	assert.NotNil(t, panel)
	assert.Equal(t, 0, panel.width)
	assert.Equal(t, 0, panel.height)
	assert.Nil(t, panel.stats)
}

// TestStatsPanel_SetStats tests setting statistics
func TestStatsPanel_SetStats(t *testing.T) {
	panel := NewStatsPanel()

	stats := &api.GlobalStats{
		DlInfoSpeed:      1048576,     // 1 MB/s
		UpInfoSpeed:      524288,      // 512 KB/s
		DlInfoData:       500000000,   // 500 MB
		UpInfoData:       250000000,   // 250 MB
		FreeSpaceOnDisk:  10737418240, // 10 GB
		DHTNodes:         150,
		ConnectionStatus: "connected",
		TorrentsCount:    10,
		NumTorrents:      10,
		NumActiveItems:   5,
		DHT:              true,
		PeerExchange:     true,
	}

	panel.SetStats(stats)
	assert.Equal(t, stats, panel.stats)
}

// TestStatsPanel_SetDimensions tests dimension setting
func TestStatsPanel_SetDimensions(t *testing.T) {
	panel := NewStatsPanel()

	panel.SetDimensions(100, 5)
	assert.Equal(t, 100, panel.width)
	assert.Equal(t, 5, panel.height)
}

// TestStatsPanel_ViewRendering tests view output
func TestStatsPanel_ViewRendering(t *testing.T) {
	panel := NewStatsPanel()
	panel.SetDimensions(100, 5)

	t.Run("render without stats", func(t *testing.T) {
		view := panel.View()
		assert.Contains(t, view, "Loading statistics...")
	})

	t.Run("render with stats", func(t *testing.T) {
		stats := &api.GlobalStats{
			DlInfoSpeed:      10485760,    // 10 MB/s
			UpInfoSpeed:      5242880,     // 5 MB/s
			DlInfoData:       536870912,   // 512 MB
			UpInfoData:       268435456,   // 256 MB
			FreeSpaceOnDisk:  53687091200, // 50 GB
			ConnectionStatus: "connected",
			TorrentsCount:    15,
			DHTNodes:         200,
		}

		panel.SetStats(stats)
		view := panel.View()

		// Check for download stats
		assert.Contains(t, view, "↓")
		assert.Contains(t, view, "10.0 MB/s")
		assert.Contains(t, view, "512.0 MB")
		assert.Contains(t, view, "15 torrents")

		// Check for upload stats
		assert.Contains(t, view, "↑")
		assert.Contains(t, view, "5.0 MB/s")
		assert.Contains(t, view, "256.0 MB")

		// Check for disk space
		assert.Contains(t, view, "Free space:")
		assert.Contains(t, view, "50.0 GB")

		// Check for connection status - Status is uppercase
		assert.Contains(t, view, "Connected")
	})

	t.Run("render disconnected state", func(t *testing.T) {
		stats := &api.GlobalStats{
			ConnectionStatus: "disconnected",
		}

		panel.SetStats(stats)
		view := panel.View()

		assert.Contains(t, view, "Disconnected")
	})

	t.Run("render firewalled state", func(t *testing.T) {
		stats := &api.GlobalStats{
			ConnectionStatus: "firewalled",
		}

		panel.SetStats(stats)
		view := panel.View()

		assert.Contains(t, view, "Firewalled")
	})
}

// TestStatsPanel_ResponsiveLayout tests layout rendering at different widths
func TestStatsPanel_ResponsiveLayout(t *testing.T) {
	stats := &api.GlobalStats{
		DlInfoSpeed:      10485760,
		UpInfoSpeed:      5242880,
		DlInfoData:       536870912,
		UpInfoData:       268435456,
		FreeSpaceOnDisk:  53687091200,
		ConnectionStatus: "connected",
		TorrentsCount:    20,
		DHTNodes:         300,
	}

	tests := []struct {
		name  string
		width int
	}{
		{"narrow width", 50},
		{"medium width", 80},
		{"wide width", 120},
		{"extra wide", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panel := NewStatsPanel()
			panel.SetDimensions(tt.width, 5)
			panel.SetStats(stats)

			view := panel.View()

			// All information should be shown regardless of width
			// The panel uses horizontal layout with fixed sections
			assert.Contains(t, view, "10.0 MB/s")   // Speed
			assert.Contains(t, view, "512.0 MB")    // Session data
			assert.Contains(t, view, "20 torrents") // Torrents count
			assert.Contains(t, view, "50.0 GB")     // Free space
			assert.Contains(t, view, "Connected")   // Connection status
			assert.Contains(t, view, "300 nodes")   // DHT nodes
		})
	}
}

// TestStatsPanel_FormatSpeed tests speed formatting
func TestStatsPanel_FormatSpeed(t *testing.T) {
	tests := []struct {
		name     string
		speed    int64
		expected string
	}{
		{"zero speed", 0, "0 B/s"},
		{"bytes per second", 512, "512 B/s"},
		{"kilobytes per second", 1024, "1.0 KB/s"},
		{"megabytes per second", 1048576, "1.0 MB/s"},
		{"gigabytes per second", 1073741824, "1.0 GB/s"},
		{"fractional MB", 1572864, "1.5 MB/s"},
	}

	panel := NewStatsPanel()
	panel.SetDimensions(100, 5)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &api.GlobalStats{
				DlInfoSpeed: tt.speed,
			}
			panel.SetStats(stats)
			view := panel.View()

			assert.Contains(t, view, tt.expected)
		})
	}
}

// TestStatsPanel_ConnectionStatusColors tests connection status rendering
func TestStatsPanel_ConnectionStatusColors(t *testing.T) {
	panel := NewStatsPanel()
	panel.SetDimensions(100, 5)

	tests := []struct {
		status   string
		contains string
	}{
		{"connected", "Connected"},
		{"disconnected", "Disconnected"},
		{"firewalled", "Firewalled"},
		{"", "Disconnected"}, // Unknown status defaults to disconnected
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			stats := &api.GlobalStats{
				ConnectionStatus: tt.status,
			}
			panel.SetStats(stats)
			view := panel.View()

			assert.Contains(t, view, tt.contains)
		})
	}
}

// TestStatsPanel_EdgeCases tests edge cases and error conditions
func TestStatsPanel_EdgeCases(t *testing.T) {
	panel := NewStatsPanel()

	t.Run("zero dimensions", func(t *testing.T) {
		// Stats panel should still render even with zero dimensions
		panel.SetDimensions(0, 0)
		panel.SetStats(nil)
		view := panel.View()
		assert.Contains(t, view, "Loading statistics...")
	})

	t.Run("very narrow width", func(t *testing.T) {
		panel.SetDimensions(20, 5)
		stats := &api.GlobalStats{
			DlInfoSpeed: 1048576,
			UpInfoSpeed: 524288,
		}
		panel.SetStats(stats)

		view := panel.View()
		// Should still render something meaningful
		assert.NotEmpty(t, view)
		assert.NotEqual(t, "Loading statistics...", view)
	})

	t.Run("negative values", func(t *testing.T) {
		panel.SetDimensions(100, 5)
		stats := &api.GlobalStats{
			DlInfoSpeed:     -1000,
			UpInfoSpeed:     -2000,
			FreeSpaceOnDisk: -1,
		}
		panel.SetStats(stats)

		view := panel.View()
		// Negative values are shown as-is in the current implementation
		// This could be improved to show 0 instead
		assert.Contains(t, view, "-")
	})
}

// TestStatsPanel_MultilineLayout tests multi-line rendering
func TestStatsPanel_MultilineLayout(t *testing.T) {
	panel := NewStatsPanel()
	panel.SetDimensions(150, 3) // 3 lines of height

	stats := &api.GlobalStats{
		DlInfoSpeed:      52428800,     // 50 MB/s
		UpInfoSpeed:      26214400,     // 25 MB/s
		DlInfoData:       1073741824,   // 1 GB
		UpInfoData:       536870912,    // 512 MB
		FreeSpaceOnDisk:  107374182400, // 100 GB
		ConnectionStatus: "connected",
		DHTNodes:         250,
		TorrentsCount:    50,
	}

	panel.SetStats(stats)
	view := panel.View()

	// The stats panel uses horizontal layout, not vertical lines for dl/up
	// So both download and upload info should be on the same line
	assert.Contains(t, view, "↓")
	assert.Contains(t, view, "50.0 MB/s")
	assert.Contains(t, view, "↑")
	assert.Contains(t, view, "25.0 MB/s")
}

// TestStatsPanel_CompactMode tests very compact rendering
func TestStatsPanel_CompactMode(t *testing.T) {
	panel := NewStatsPanel()
	panel.SetDimensions(40, 2) // Very limited space

	stats := &api.GlobalStats{
		DlInfoSpeed:      1048576,
		UpInfoSpeed:      524288,
		ConnectionStatus: "connected",
	}

	panel.SetStats(stats)
	view := panel.View()

	// Should still show essential information
	assert.Contains(t, view, "1.0 MB/s")
	assert.Contains(t, view, "512.0 KB/s")

	// The stats panel uses fixed width sections that may exceed 40 chars
	// when rendered with styles. The important thing is it shows the data.
}

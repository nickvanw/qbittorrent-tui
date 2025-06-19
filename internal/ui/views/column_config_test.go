package views

import (
	"errors"
	"strings"
	"testing"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/testutil"
	"github.com/stretchr/testify/assert"
)

// TestColumnConfiguration_ComprehensiveToggling tests all column configuration scenarios
func TestColumnConfiguration_ComprehensiveToggling(t *testing.T) {
	t.Run("enter and exit column config mode", func(t *testing.T) {
		view, _ := createTestMainView()

		// Add at least one torrent so the View doesn't return early
		mockTorrents := []api.Torrent{{Name: "Test", Size: 1000}}
		view.torrents = mockTorrents
		view.torrentList.SetTorrents(mockTorrents)

		// Update torrent list to ensure proper initialization
		view.torrentList.SetDimensions(view.width-4, view.height-10)

		// Verify not in config mode initially
		assert.False(t, view.torrentList.IsInConfigMode())

		// Press 'C' to enter config mode
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		// The torrent list component should render the overlay
		torrentListView := mainView.torrentList.View()
		assert.Contains(t, torrentListView, "Column Configuration")

		// ESC exits config mode
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.torrentList.IsInConfigMode())

		// View should not show column configuration
		torrentListView = mainView.torrentList.View()
		assert.NotContains(t, torrentListView, "Column Configuration")
	})

	t.Run("toggle individual columns", func(t *testing.T) {
		view, _ := createTestMainView()
		mockTorrents := []api.Torrent{
			{
				Name:       "Test Torrent",
				Size:       1000000,
				Progress:   0.5,
				State:      "downloading",
				DlSpeed:    100000,
				UpSpeed:    50000,
				NumSeeds:   10,
				NumLeeches: 5,
				Ratio:      1.5,
			},
		}
		view.torrents = mockTorrents
		view.torrentList.SetTorrents(mockTorrents)

		// Enter config mode
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)

		torrentListView := mainView.torrentList.View()
		// Should show all column options with numbers
		assert.Contains(t, torrentListView, "Name")
		assert.Contains(t, torrentListView, "Size")
		assert.Contains(t, torrentListView, "Progress")
		assert.Contains(t, torrentListView, "Status")
		assert.Contains(t, torrentListView, "Down")
		assert.Contains(t, torrentListView, "Up")
		assert.Contains(t, torrentListView, "Seeds")
		assert.Contains(t, torrentListView, "Peers")
		assert.Contains(t, torrentListView, "Ratio")
		assert.Contains(t, torrentListView, "ETA")
		assert.Contains(t, torrentListView, "Category")
		assert.Contains(t, torrentListView, "Tags")
		assert.Contains(t, torrentListView, "Added")
		assert.Contains(t, torrentListView, "Tracker")

		// Toggle off Progress column (3)
		model = testutil.SendKey(t, mainView, "3")
		mainView = model.(*MainView)

		// Exit config mode
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)

		// Check that Progress is not shown in the list
		lines := testutil.GetViewLines(mainView.torrentList)

		// Find header line (should contain column headers)
		headerLine := ""
		for _, line := range lines {
			if strings.Contains(line, "Name") && strings.Contains(line, "Size") {
				headerLine = line
				break
			}
		}

		// Progress column should not be in the header
		assert.NotContains(t, headerLine, "Progress")
		assert.NotContains(t, headerLine, "%") // Progress uses % symbol
	})

	t.Run("toggle multiple columns", func(t *testing.T) {
		view, _ := createTestMainView()

		// Enter config mode
		model := testutil.SendKey(t, view, "C")

		// Toggle off multiple columns
		model = testutil.SendKey(t, model, "5") // Seeds
		model = testutil.SendKey(t, model, "6") // Peers
		model = testutil.SendKey(t, model, "7") // Down Speed
		model = testutil.SendKey(t, model, "8") // Up Speed

		mainView := model.(*MainView)

		// Exit config mode
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)

		// Set some torrents to verify columns
		mockTorrents := []api.Torrent{
			{
				Name:       "Test Torrent",
				Size:       1000000,
				Progress:   0.5,
				State:      "downloading",
				DlSpeed:    100000,
				UpSpeed:    50000,
				NumSeeds:   10,
				NumLeeches: 5,
			},
		}
		mainView.torrents = mockTorrents
		mainView.torrentList.SetTorrents(mockTorrents)

		lines := testutil.GetViewLines(mainView)

		// Find header line
		headerLine := ""
		for _, line := range lines {
			if strings.Contains(line, "Name") {
				headerLine = line
				break
			}
		}

		// These columns should not be shown
		assert.NotContains(t, headerLine, "Seeds")
		assert.NotContains(t, headerLine, "Peers")
		assert.NotContains(t, headerLine, "Down")
		assert.NotContains(t, headerLine, "Up")

		// These should still be shown
		assert.Contains(t, headerLine, "Name")
		assert.Contains(t, headerLine, "Size")
		assert.Contains(t, headerLine, "Progress")
		assert.Contains(t, headerLine, "Status")
	})

	t.Run("column visibility persists across config sessions", func(t *testing.T) {
		view, _ := createTestMainView()

		// Add at least one torrent so the View doesn't return early
		mockTorrents := []api.Torrent{{Name: "Test", Size: 1000}}
		view.torrents = mockTorrents
		view.torrentList.SetTorrents(mockTorrents)

		// First session: turn off some columns
		model := testutil.SendKey(t, view, "C")
		model = testutil.SendKey(t, model, "3") // Progress
		model = testutil.SendKey(t, model, "5") // Seeds
		model = testutil.SendKey(t, model, "esc")

		// Second session: check state
		model = testutil.SendKey(t, model, "C")
		mainView := model.(*MainView)

		// Progress and Seeds should show as unchecked
		// Look for the checkbox pattern - unchecked should be [ ]
		configView := mainView.torrentList.View()

		// In config mode, we should see the column configuration overlay
		assert.Contains(t, configView, "Column Configuration", "Should be in config mode")

		// Look for the checkbox patterns directly in the rendered view
		assert.Contains(t, configView, "3.  [ ] Progress", "Progress column should be unchecked")
		assert.Contains(t, configView, "5.  [ ] Down", "Down column should be unchecked")
	})

	t.Run("invalid keys in config mode", func(t *testing.T) {
		view, _ := createTestMainView()

		// Enter config mode
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		// Try invalid keys - should stay in config mode
		model = testutil.SendKey(t, mainView, "z")
		mainView = model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		model = testutil.SendKey(t, mainView, "!")
		mainView = model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		// Navigation keys should be ignored
		model = testutil.SendKey(t, mainView, "up")
		mainView = model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())
	})

	t.Run("column config with no torrents", func(t *testing.T) {
		view, _ := createTestMainView()
		view.torrents = []api.Torrent{}
		view.torrentList.SetTorrents([]api.Torrent{})

		// Should still be able to configure columns
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		output := mainView.View()
		assert.Contains(t, output, "Column Configuration")

		// Toggle some columns
		model = testutil.SendKey(t, mainView, "3")
		model = testutil.SendKey(t, model, "5")

		// Exit
		model = testutil.SendKey(t, model, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.torrentList.IsInConfigMode())
	})

	t.Run("all columns can be toggled", func(t *testing.T) {
		view, _ := createTestMainView()

		// Enter config mode
		model := testutil.SendKey(t, view, "C")

		// Toggle all numeric columns
		for i := 0; i <= 9; i++ {
			model = testutil.SendKey(t, model, string(rune('0'+i)))
		}

		// Toggle letter columns
		model = testutil.SendKey(t, model, "q") // Category
		model = testutil.SendKey(t, model, "w") // Tags
		model = testutil.SendKey(t, model, "e") // Added On
		model = testutil.SendKey(t, model, "r") // Completed On

		mainView := model.(*MainView)

		// Exit config mode
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)

		// With all columns hidden, should show minimal view
		mockTorrents := []api.Torrent{{Name: "Test"}}
		mainView.torrents = mockTorrents
		mainView.torrentList.SetTorrents(mockTorrents)

		output := mainView.View()
		// Should still show something, even with all columns hidden
		assert.NotEmpty(t, output)
	})
}

// TestColumnConfiguration_ViewRendering tests how column config affects rendering
func TestColumnConfiguration_ViewRendering(t *testing.T) {
	t.Run("column config overlay rendering", func(t *testing.T) {
		view, _ := createTestMainView()
		view.width = 120
		view.height = 40

		// Add some torrents
		mockTorrents := []api.Torrent{
			{Name: "Torrent 1", Size: 1000000},
			{Name: "Torrent 2", Size: 2000000},
		}
		view.torrents = mockTorrents
		view.torrentList.SetTorrents(mockTorrents)

		// Enter config mode
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)

		output := mainView.View()
		lines := testutil.GetViewLines(mainView)

		// Should show config overlay
		assert.Contains(t, output, "Column Configuration")
		assert.Contains(t, output, "Toggle column visibility: 1-9, 0, q, w, e, r")

		// Should show columns in a grid layout - Name and Size are in different lines
		nameFound := false
		sizeFound := false
		for _, line := range lines {
			if strings.Contains(line, "1.") && strings.Contains(line, "Name") {
				nameFound = true
			}
			if strings.Contains(line, "2.") && strings.Contains(line, "Size") {
				sizeFound = true
			}
		}
		assert.True(t, nameFound && sizeFound, "Should show Name and Size column options in grid")

		// Background should still be visible (dimmed)
		assert.True(t, len(lines) > 20, "Should show full height with overlay")
	})

	t.Run("checkbox state rendering", func(t *testing.T) {
		view, _ := createTestMainView()

		// Enter config mode
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)

		output := mainView.View()

		// All columns should be checked by default
		assert.Contains(t, output, "[✓] Name")
		assert.Contains(t, output, "[✓] Size")
		assert.Contains(t, output, "[✓] Progress")

		// Toggle Progress off
		model = testutil.SendKey(t, mainView, "3")
		mainView = model.(*MainView)

		output = mainView.View()
		// Progress should now be unchecked
		assert.Contains(t, output, "[ ] Progress")
		// Others should still be checked
		assert.Contains(t, output, "[✓] Name")
		assert.Contains(t, output, "[✓] Size")
	})
}

// TestColumnConfiguration_KeyBindings tests all key bindings in config mode
func TestColumnConfiguration_KeyBindings(t *testing.T) {
	// Default visible columns (from torrent_list.go defaultVisibleColumns)
	defaultVisible := map[string]bool{
		"Name": true, "Size": true, "Progress": true, "Status": true,
		"Down": true, "Up": true, "Seeds": true, "Peers": true, "Ratio": true,
		"ETA": false, "Category": false, "Tags": false, "Added": false, "Tracker": false,
	}

	columnKeys := map[string]string{
		"1": "Name",
		"2": "Size",
		"3": "Progress",
		"4": "Status",
		"5": "Down",
		"6": "Up",
		"7": "Seeds",
		"8": "Peers",
		"9": "Ratio",
		"0": "ETA",
		"q": "Added",    // 11th column (position 10)
		"w": "Category", // 12th column (position 11)
		"e": "Tags",     // 13th column (position 12)
		"r": "Tracker",  // 14th column (position 13)
	}

	for key, columnName := range columnKeys {
		t.Run("toggle "+columnName+" with key "+key, func(t *testing.T) {
			view, _ := createTestMainView()

			// Enter config mode
			model := testutil.SendKey(t, view, "C")

			// Press the key to toggle
			model = testutil.SendKey(t, model, key)
			mainView := model.(*MainView)

			// Should still be in config mode
			assert.True(t, mainView.torrentList.IsInConfigMode())

			configView := mainView.torrentList.View()

			// Check the state after first toggle
			if defaultVisible[columnName] {
				// Was visible, now should be unchecked
				expectedPattern := key + ".  [ ] " + columnName
				assert.Contains(t, configView, expectedPattern, "Column %s should be unchecked after first toggle", columnName)
			} else {
				// Was not visible, now should be checked
				expectedPattern := key + ".  [✓] " + columnName
				assert.Contains(t, configView, expectedPattern, "Column %s should be checked after first toggle", columnName)
			}

			// Toggle again
			model = testutil.SendKey(t, mainView, key)
			mainView = model.(*MainView)

			configView = mainView.torrentList.View()

			// Check the state after second toggle (should be back to original)
			if defaultVisible[columnName] {
				// Should be back to checked
				expectedPattern := key + ".  [✓] " + columnName
				assert.Contains(t, configView, expectedPattern, "Column %s should be checked after second toggle", columnName)
			} else {
				// Should be back to unchecked
				expectedPattern := key + ".  [ ] " + columnName
				assert.Contains(t, configView, expectedPattern, "Column %s should be unchecked after second toggle", columnName)
			}
		})
	}
}

// TestColumnConfiguration_EdgeCases tests edge cases and error conditions
func TestColumnConfiguration_EdgeCases(t *testing.T) {
	t.Run("rapid key presses", func(t *testing.T) {
		view, _ := createTestMainView()

		// Enter config mode
		model := testutil.SendKey(t, view, "C")

		// Rapid toggles of same column
		for i := 0; i < 10; i++ {
			model = testutil.SendKey(t, model, "3")
		}

		mainView := model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		// Should handle gracefully - even number of toggles means column is on
		output := mainView.View()
		assert.Contains(t, output, "[✓] Progress") // 10 toggles = on
	})

	t.Run("config mode during loading", func(t *testing.T) {
		view, _ := createTestMainView()
		view.isLoading = true

		// Should still allow config mode during loading
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)

		// Config mode should work even when loading
		assert.True(t, mainView.torrentList.IsInConfigMode())

		output := mainView.View()
		assert.Contains(t, output, "Column Configuration")
	})

	t.Run("config mode with error state", func(t *testing.T) {
		view, _ := createTestMainView()
		view.lastError = errors.New("Connection failed")

		// Should still allow config mode with error
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)

		assert.True(t, mainView.torrentList.IsInConfigMode())

		output := mainView.View()
		assert.Contains(t, output, "Column Configuration")
	})
}

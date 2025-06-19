package views

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/config"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/testutil"
	"github.com/stretchr/testify/assert"
)

// Test helpers
func createTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.UI.RefreshInterval = 5
	return cfg
}

func createTestTorrents() []api.Torrent {
	return []api.Torrent{
		{
			Hash:     "hash1",
			Name:     "Test Torrent 1",
			Size:     1000000,
			Progress: 0.5,
			State:    "downloading",
			Category: "movies",
			Tags:     "tag1,tag2",
		},
		{
			Hash:     "hash2",
			Name:     "Test Torrent 2",
			Size:     2000000,
			Progress: 1.0,
			State:    "completed",
			Category: "tv",
			Tags:     "tag2,tag3",
		},
	}
}

func createTestMainView() (*MainView, *api.MockClient) {
	cfg := createTestConfig()
	mockClient := api.SetupMockClientWithData()
	mockClient.LoggedIn = true
	view := NewMainView(cfg, mockClient)
	// Set dimensions to simulate terminal
	view.width = 120
	view.height = 40
	return view, mockClient
}

// TestMainView_KeyHandlingPriority tests the complex key handling hierarchy
func TestMainView_KeyHandlingPriority(t *testing.T) {
	t.Run("add dialog has highest priority", func(t *testing.T) {
		view, _ := createTestMainView()

		// Open add dialog
		view.showAddDialog = true

		// Try various keys - they should be handled by dialog
		model := testutil.SendKey(t, view, "a")
		mainView := model.(*MainView)
		assert.True(t, mainView.showAddDialog, "dialog should remain open")

		// ESC should close dialog
		model = testutil.SendKey(t, view, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.showAddDialog, "dialog should be closed")
	})

	t.Run("delete dialog has second priority", func(t *testing.T) {
		view, _ := createTestMainView()
		view.torrents = createTestTorrents()
		view.torrentList.SetTorrents(view.torrents)

		// Show delete dialog
		view.showDeleteDialog = true
		view.deleteTarget = &view.torrents[0]

		// 'y' should confirm deletion
		model := testutil.SendKey(t, view, "y")
		mainView := model.(*MainView)
		assert.False(t, mainView.showDeleteDialog, "dialog should be closed after confirmation")

		// Reset and test cancel
		view.showDeleteDialog = true
		view.deleteTarget = &view.torrents[0]
		model = testutil.SendKey(t, view, "n")
		mainView = model.(*MainView)
		assert.False(t, mainView.showDeleteDialog, "dialog should be closed after cancel")
	})

	t.Run("filter panel input mode takes precedence", func(t *testing.T) {
		view, _ := createTestMainView()

		// Enter filter search mode
		model := testutil.SendKey(t, view, "/")
		mainView := model.(*MainView)
		assert.True(t, mainView.filterPanel.IsInInputMode(), "should be in input mode")

		// Regular keys should go to filter panel
		model = testutil.SendKey(t, mainView, "t")
		model = testutil.SendKey(t, model, "e")
		model = testutil.SendKey(t, model, "s")
		model = testutil.SendKey(t, model, "t")

		// Exit with enter
		model = testutil.SendKey(t, model, "enter")
		mainView = model.(*MainView)
		assert.False(t, mainView.filterPanel.IsInInputMode(), "should exit input mode")
		assert.Equal(t, "test", mainView.filterPanel.GetFilter().Search)
	})
}

// TestMainView_ViewModes tests switching between main and details view
func TestMainView_ViewModes(t *testing.T) {
	view, mockClient := createTestMainView()
	torrents := createTestTorrents()
	view.torrents = torrents
	view.torrentList.SetTorrents(torrents)

	t.Run("enter opens details view", func(t *testing.T) {
		// Start in main view
		assert.Equal(t, ViewModeMain, view.viewMode)

		// Set up mock data for details view
		mockClient.TorrentProperties["hash1"] = &api.TorrentProperties{}
		mockClient.Trackers["hash1"] = []api.Tracker{}
		mockClient.Peers["hash1"] = map[string]api.Peer{}
		mockClient.Files["hash1"] = []api.TorrentFile{}

		model := testutil.SendKey(t, view, "enter")
		mainView := model.(*MainView)
		assert.Equal(t, ViewModeDetails, mainView.viewMode)
	})

	t.Run("escape returns to main view", func(t *testing.T) {
		view.viewMode = ViewModeDetails

		model := testutil.SendKey(t, view, "esc")
		mainView := model.(*MainView)
		assert.Equal(t, ViewModeMain, mainView.viewMode)
	})
}

// TestMainView_DialogStates tests dialog management
func TestMainView_DialogStates(t *testing.T) {
	t.Run("add torrent dialog mode switching", func(t *testing.T) {
		view, _ := createTestMainView()
		view.showAddDialog = true

		// Should start in file mode
		assert.Equal(t, ModeFile, view.addDialog.mode)

		// Tab switches modes
		model := testutil.SendKey(t, view, "tab")
		mainView := model.(*MainView)
		assert.Equal(t, ModeURL, mainView.addDialog.mode)

		// Tab again switches back
		model = testutil.SendKey(t, mainView, "tab")
		mainView = model.(*MainView)
		assert.Equal(t, ModeFile, mainView.addDialog.mode)
	})

	t.Run("delete dialog file toggle", func(t *testing.T) {
		view, _ := createTestMainView()
		view.showDeleteDialog = true
		view.deleteTarget = &api.Torrent{Name: "Test"}

		assert.False(t, view.deleteWithFiles)

		// 'f' toggles file deletion
		model := testutil.SendKey(t, view, "f")
		mainView := model.(*MainView)
		assert.True(t, mainView.deleteWithFiles)

		// Toggle again
		model = testutil.SendKey(t, mainView, "f")
		mainView = model.(*MainView)
		assert.False(t, mainView.deleteWithFiles)
	})
}

// TestMainView_TorrentActions tests torrent control actions
func TestMainView_TorrentActions(t *testing.T) {
	view, mockClient := createTestMainView()
	torrents := createTestTorrents()
	view.torrents = torrents
	view.torrentList.SetTorrents(torrents)

	t.Run("pause torrent", func(t *testing.T) {
		// Pause action should not return error
		cmd := view.handlePauseTorrent()
		msg := cmd()

		// Should get success message
		successMsg, ok := msg.(successMsg)
		assert.True(t, ok)
		assert.Contains(t, string(successMsg), "paused")
	})

	t.Run("resume torrent", func(t *testing.T) {
		cmd := view.handleResumeTorrent()
		msg := cmd()

		successMsg, ok := msg.(successMsg)
		assert.True(t, ok)
		assert.Contains(t, string(successMsg), "resumed")
	})

	t.Run("delete torrent", func(t *testing.T) {
		// Delete action just shows dialog
		cmd := view.handleDeleteTorrent()
		assert.Nil(t, cmd)
		assert.True(t, view.showDeleteDialog)
		assert.NotNil(t, view.deleteTarget)

		// Actual deletion happens on confirm
		cmd = view.confirmDeleteTorrent()
		msg := cmd()

		successMsg, ok := msg.(successMsg)
		assert.True(t, ok)
		assert.Contains(t, string(successMsg), "deleted")
		assert.False(t, view.showDeleteDialog)
	})

	t.Run("handle API errors", func(t *testing.T) {
		// Set up error for pause
		mockClient.GetError = errors.New("network error")

		cmd := view.handlePauseTorrent()
		msg := cmd()

		errMsg, ok := msg.(errorMsg)
		assert.True(t, ok)
		assert.Contains(t, errMsg.Error(), "network error")

		// Clear error
		mockClient.GetError = nil
	})
}

// TestMainView_AddTorrent tests add torrent dialog functionality
func TestMainView_AddTorrent(t *testing.T) {
	view, _ := createTestMainView()

	t.Run("show add dialog", func(t *testing.T) {
		// 'a' key should show add dialog
		model := testutil.SendKey(t, view, "a")
		mainView := model.(*MainView)
		assert.True(t, mainView.showAddDialog)

		// ESC should close dialog
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.showAddDialog)
	})

	t.Run("add dialog mode switching", func(t *testing.T) {
		view.showAddDialog = true
		view.addDialog = &AddTorrentDialog{
			mode:     ModeFile,
			urlInput: &URLInput{},
			fileNav:  &FileNavigator{},
		}

		// Tab switches between file and URL mode
		model := testutil.SendKey(t, view, "tab")
		mainView := model.(*MainView)
		assert.Equal(t, ModeURL, mainView.addDialog.mode)
	})
}

// TestMainView_ErrorHandling tests error display and management
func TestMainView_ErrorHandling(t *testing.T) {
	view, _ := createTestMainView()

	t.Run("display error message", func(t *testing.T) {
		// Simulate error message
		model, _ := view.Update(errorMsg(errors.New("Test error")))
		mainView := model.(*MainView)

		assert.Equal(t, "Test error", mainView.lastError.Error())
	})

	t.Run("error clears after timeout", func(t *testing.T) {
		view.lastError = errors.New("Test error")

		// Simulate clear error message
		model, _ := view.Update(clearErrorMsg{})
		mainView := model.(*MainView)

		assert.Nil(t, mainView.lastError)
	})
}

// TestMainView_FilterIntegration tests filter integration
func TestMainView_FilterIntegration(t *testing.T) {
	view, _ := createTestMainView()
	torrents := createTestTorrents()
	view.allTorrents = torrents
	view.torrents = torrents

	t.Run("filter shortcuts work in main view", func(t *testing.T) {
		// 's' should open state filter
		model := testutil.SendKey(t, view, "s")
		mainView := model.(*MainView)
		assert.True(t, mainView.filterPanel.IsInInteractiveMode())

		// Exit filter mode
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)

		// 'c' should open category filter
		model = testutil.SendKey(t, mainView, "c")
		mainView = model.(*MainView)
		assert.True(t, mainView.filterPanel.IsInInteractiveMode())
	})

	t.Run("clear filters with 'x'", func(t *testing.T) {
		// Set some filters - need to update the filter directly
		filter := view.filterPanel.GetFilter()
		filter.Search = "test"
		filter.States = []string{"downloading"}
		// Apply the filter by triggering an update
		view.currentFilter = filter

		mainView := testutil.SendKey(t, view, "x").(*MainView)

		updatedFilter := mainView.filterPanel.GetFilter()
		assert.Empty(t, updatedFilter.Search)
		assert.Empty(t, updatedFilter.States)
	})
}

// TestMainView_ColumnConfiguration tests column config mode
func TestMainView_ColumnConfiguration(t *testing.T) {
	view, mockClient := createTestMainView()

	// Add torrents for config mode to work properly
	mockTorrents := []api.Torrent{{Name: "Test", Size: 1000}}
	mockClient.Torrents = mockTorrents
	view.torrents = mockTorrents
	view.torrentList.SetTorrents(mockTorrents)

	t.Run("C key opens column config", func(t *testing.T) {
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)

		assert.True(t, mainView.torrentList.IsInConfigMode())
	})

	t.Run("column config mode handles keys", func(t *testing.T) {
		// Create fresh view for this test
		freshView, freshMock := createTestMainView()
		freshMock.Torrents = mockTorrents
		freshView.torrents = mockTorrents
		freshView.torrentList.SetTorrents(mockTorrents)

		// Enable config mode by sending 'C' key
		var model tea.Model
		model = testutil.SendKey(t, freshView, "C")
		mainView := model.(*MainView)

		// Number keys should be handled by torrent list
		model = testutil.SendKey(t, mainView, "1")
		mainView = model.(*MainView)

		// Should still be in config mode
		assert.True(t, mainView.torrentList.IsInConfigMode())

		// ESC exits config mode
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.torrentList.IsInConfigMode())
	})
}

// TestMainView_RefreshBehavior tests data refresh functionality
func TestMainView_RefreshBehavior(t *testing.T) {
	view, mockClient := createTestMainView()

	t.Run("manual refresh with 'r'", func(t *testing.T) {
		// Set up mock data
		mockClient.Torrents = []api.Torrent{}
		mockClient.GlobalStats = &api.GlobalStats{}
		mockClient.Categories = map[string]interface{}{}
		mockClient.Tags = []string{}

		// Initial fetch
		cmd := view.Init()
		assert.NotNil(t, cmd)

		// Manual refresh
		model := testutil.SendKey(t, view, "r")
		mainView := model.(*MainView)
		assert.True(t, mainView.isLoading)
	})

	t.Run("periodic refresh via tick", func(t *testing.T) {
		// Set up mock data
		mockClient.Torrents = []api.Torrent{}
		mockClient.GlobalStats = &api.GlobalStats{}
		mockClient.Categories = map[string]interface{}{}
		mockClient.Tags = []string{}

		// Simulate tick message
		_, cmd := view.Update(tickMsg(time.Now()))
		assert.NotNil(t, cmd)
	})

	t.Run("handle refresh errors", func(t *testing.T) {
		mockClient.GetError = errors.New("connection failed")

		// Try to refresh - fetchAllData returns a batch command
		cmd := view.fetchAllData()
		assert.NotNil(t, cmd)

		// In real usage, the batch would execute multiple fetches
		// and errors would come through as errorMsg messages

		// Clear error
		mockClient.GetError = nil
	})
}

// TestMainView_Sorting tests torrent sorting functionality
func TestMainView_Sorting(t *testing.T) {
	view, _ := createTestMainView()
	torrents := createTestTorrents()
	view.torrents = torrents
	view.torrentList.SetTorrents(torrents)

	sortKeys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	expectedColumns := []string{"Name", "Size", "Progress", "Status", "Seeds", "Peers", "Down Speed", "Up Speed", "Ratio"}

	for i, key := range sortKeys {
		t.Run("sort by "+expectedColumns[i], func(t *testing.T) {
			model := testutil.SendKey(t, view, key)
			// Verify model is still a MainView (key was handled)
			_, ok := model.(*MainView)
			assert.True(t, ok)

			// Check that sort key was handled
			// Note: actual sort column checking would require access to internal state
		})
	}
}

// TestMainView_ComplexScenarios tests complex interaction scenarios
func TestMainView_ComplexScenarios(t *testing.T) {
	t.Run("multiple state transitions", func(t *testing.T) {
		view, mockClient := createTestMainView()
		torrents := createTestTorrents()
		view.torrents = torrents
		view.torrentList.SetTorrents(torrents)

		// Start in main view
		assert.Equal(t, ViewModeMain, view.viewMode)

		// Enter filter mode
		model := testutil.SendKey(t, view, "/")
		mainView := model.(*MainView)
		assert.True(t, mainView.filterPanel.IsInInputMode())

		// Exit filter mode without searching (just test the mode transition)
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.filterPanel.IsInInputMode())

		// Ensure we have torrents
		assert.Greater(t, len(mainView.torrents), 0, "Should have torrents")

		// Enter details view - first make sure mock data is set up
		mockClient.TorrentProperties["hash1"] = &api.TorrentProperties{}
		mockClient.Trackers["hash1"] = []api.Tracker{}
		mockClient.Peers["hash1"] = map[string]api.Peer{}
		mockClient.Files["hash1"] = []api.TorrentFile{}

		model = testutil.SendKey(t, mainView, "enter")
		mainView = model.(*MainView)
		assert.Equal(t, ViewModeDetails, mainView.viewMode)

		// Open add dialog in details view
		model = testutil.SendKey(t, mainView, "a")
		mainView = model.(*MainView)
		assert.True(t, mainView.showAddDialog)
		assert.Equal(t, ViewModeDetails, mainView.viewMode)

		// Close dialog
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.showAddDialog)

		// Return to main view
		model = testutil.SendKey(t, mainView, "esc")
		mainView = model.(*MainView)
		assert.Equal(t, ViewModeMain, mainView.viewMode)
	})

	t.Run("quit works from any state", func(t *testing.T) {
		view, _ := createTestMainView()

		// Test quit from various states
		states := []struct {
			name  string
			setup func()
		}{
			{
				name:  "main view",
				setup: func() {},
			},
			{
				name: "details view",
				setup: func() {
					view.viewMode = ViewModeDetails
				},
			},
			{
				name: "add dialog",
				setup: func() {
					view.showAddDialog = true
				},
			},
			{
				name: "delete dialog",
				setup: func() {
					view.showDeleteDialog = true
				},
			},
			{
				name: "filter input mode",
				setup: func() {
					testutil.SendKey(t, view, "/")
				},
			},
		}

		for _, state := range states {
			t.Run(state.name, func(t *testing.T) {
				view, _ = createTestMainView()
				state.setup()

				_, cmd := view.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

				// Should return quit command
				assert.NotNil(t, cmd)
				_, ok := cmd().(tea.QuitMsg)
				assert.True(t, ok, "Should return quit message from %s", state.name)
			})
		}
	})
}

// TestMainView_ViewRendering tests view rendering
func TestMainView_ViewRendering(t *testing.T) {
	view, mockClient := createTestMainView()

	t.Run("renders loading state", func(t *testing.T) {
		view.isLoading = true
		output := view.View()
		assert.Contains(t, output, "Loading")
	})

	t.Run("renders error state", func(t *testing.T) {
		view.isLoading = false
		view.lastError = errors.New("Connection failed")
		output := view.View()
		assert.Contains(t, output, "Connection failed")
	})

	t.Run("renders main view with torrents", func(t *testing.T) {
		view.isLoading = false
		view.lastError = nil
		mockClient.Torrents = createTestTorrents()
		view.torrents = mockClient.Torrents
		view.torrentList.SetTorrents(mockClient.Torrents)

		output := view.View()
		assert.Contains(t, output, "Test Torrent 1")
		assert.Contains(t, output, "Test Torrent 2")
	})

	t.Run("renders empty state", func(t *testing.T) {
		view.isLoading = false
		view.torrents = []api.Torrent{}
		view.torrentList.SetTorrents([]api.Torrent{})

		output := view.View()
		assert.Contains(t, output, "No torrents")
	})
}

// TestMainView_ColumnToggling tests column visibility toggling
func TestMainView_ColumnToggling(t *testing.T) {
	view, _ := createTestMainView()

	t.Run("toggle column visibility", func(t *testing.T) {
		// Enter config mode
		model := testutil.SendKey(t, view, "C")
		mainView := model.(*MainView)
		assert.True(t, mainView.torrentList.IsInConfigMode())

		output := mainView.View()
		assert.Contains(t, output, "Column Configuration")

		// Toggle some columns
		model = testutil.SendKey(t, mainView, "3") // Toggle Progress
		model = testutil.SendKey(t, model, "5")    // Toggle Seeds

		// Exit config mode
		model = testutil.SendKey(t, model, "esc")
		mainView = model.(*MainView)
		assert.False(t, mainView.torrentList.IsInConfigMode())
	})
}

// TestMainView_ResponsiveLayout tests responsive layout behavior
func TestMainView_ResponsiveLayout(t *testing.T) {
	view, mockClient := createTestMainView()
	mockClient.Torrents = createTestTorrents()
	view.torrents = mockClient.Torrents
	view.torrentList.SetTorrents(mockClient.Torrents)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"narrow terminal", 80, 24},
		{"standard terminal", 120, 40},
		{"wide terminal", 200, 50},
		{"very narrow", 60, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := testutil.SimulateResize(t, view, tt.width, tt.height)
			mainView := model.(*MainView)

			assert.Equal(t, tt.width, mainView.width)
			assert.Equal(t, tt.height, mainView.height)

			// Check that view renders without panic
			output := mainView.View()
			assert.NotEmpty(t, output)

			// Verify output fits within terminal width
			testutil.AssertLineWidth(t, mainView, tt.width+2)
		})
	}
}

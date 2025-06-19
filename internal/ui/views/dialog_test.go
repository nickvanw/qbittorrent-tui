package views

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nickvanw/qbittorrent-tui/internal/api"

	// Removed unused import
	"github.com/stretchr/testify/assert"
)

// TestFileNavigator tests file navigation functionality
func TestFileNavigator(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test files and directories
	os.MkdirAll(filepath.Join(tmpDir, "subdir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "subdir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.torrent"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir1", "nested.torrent"), []byte("test"), 0644)

	t.Run("initialization", func(t *testing.T) {
		nav := NewFileNavigator(tmpDir)

		assert.Equal(t, tmpDir, nav.currentPath)
		assert.Equal(t, "*.torrent", nav.searchPattern)
		assert.False(t, nav.searchMode)
		assert.Greater(t, len(nav.allEntries), 0)
		assert.Greater(t, len(nav.filtered), 0)
	})

	t.Run("filter application", func(t *testing.T) {
		nav := NewFileNavigator(tmpDir)

		// Should show directories and .torrent files
		torrentCount := 0
		dirCount := 0
		for _, entry := range nav.filtered {
			if entry.isDir {
				dirCount++
			} else if filepath.Ext(entry.name) == ".torrent" {
				torrentCount++
			}
		}

		assert.Equal(t, 1, torrentCount)      // file1.torrent
		assert.GreaterOrEqual(t, dirCount, 2) // subdir1, subdir2, and possibly ".."
	})

	t.Run("navigation", func(t *testing.T) {
		nav := NewFileNavigator(tmpDir)

		// Find and select subdir1
		subdir1Idx := -1
		for i, entry := range nav.filtered {
			if entry.name == "subdir1" && entry.isDir {
				subdir1Idx = i
				break
			}
		}
		assert.NotEqual(t, -1, subdir1Idx)

		// Navigate to subdir1
		nav.selectedIdx = subdir1Idx
		selected := nav.filtered[nav.selectedIdx]
		nav.currentPath = selected.fullPath
		nav.readDirectory()

		// Should now see nested.torrent
		foundNested := false
		for _, entry := range nav.filtered {
			if entry.name == "nested.torrent" {
				foundNested = true
				break
			}
		}
		assert.True(t, foundNested)

		// Should have ".." entry to go back
		foundParent := false
		for _, entry := range nav.filtered {
			if entry.name == ".." && entry.isDir {
				foundParent = true
				break
			}
		}
		assert.True(t, foundParent)
	})

	t.Run("search pattern change", func(t *testing.T) {
		nav := NewFileNavigator(tmpDir)

		// Change pattern to show all files
		nav.searchPattern = "*"
		nav.applyFilter()

		// Should now show .txt file too
		foundTxt := false
		for _, entry := range nav.filtered {
			if entry.name == "file2.txt" {
				foundTxt = true
				break
			}
		}
		assert.True(t, foundTxt)

		// Change pattern to show only .txt
		nav.searchPattern = "*.txt"
		nav.applyFilter()

		// Should not show .torrent files
		foundTorrent := false
		for _, entry := range nav.filtered {
			if !entry.isDir && filepath.Ext(entry.name) == ".torrent" {
				foundTorrent = true
				break
			}
		}
		assert.False(t, foundTorrent)
	})
}

// TestFileNavigatorKeyHandling tests keyboard navigation
func TestFileNavigatorKeyHandling(t *testing.T) {
	view, _ := createTestMainView()
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "a.torrent"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.torrent"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "c.torrent"), []byte("test"), 0644)

	view.addDialog = &AddTorrentDialog{
		mode:     ModeFile,
		fileNav:  NewFileNavigator(tmpDir),
		urlInput: &URLInput{},
	}
	view.showAddDialog = true

	t.Run("cursor movement", func(t *testing.T) {
		nav := view.addDialog.fileNav
		initialIdx := nav.selectedIdx

		// Move down
		view.handleFileNavigatorKeys("down")
		assert.Greater(t, nav.selectedIdx, initialIdx)

		// Move up
		view.handleFileNavigatorKeys("up")
		assert.Equal(t, initialIdx, nav.selectedIdx)

		// Move to end
		for i := 0; i < 10; i++ {
			view.handleFileNavigatorKeys("down")
		}
		endIdx := nav.selectedIdx

		// Should be at last item
		view.handleFileNavigatorKeys("down")
		assert.Equal(t, endIdx, nav.selectedIdx)
	})

	t.Run("search mode", func(t *testing.T) {
		nav := view.addDialog.fileNav

		// Enter search mode
		view.handleFileNavigatorKeys("/")
		assert.True(t, nav.searchMode)
		assert.Equal(t, "", nav.searchPattern)

		// Type pattern
		view.handleFileNavigatorKeys("a")
		view.handleFileNavigatorKeys("*")
		assert.Equal(t, "a*", nav.searchPattern)

		// Exit search mode
		nav.searchMode = false
		nav.searchPattern = "*.torrent"
		nav.applyFilter()
	})

	t.Run("file selection", func(t *testing.T) {
		// MockClient handles the torrent file operations
		nav := view.addDialog.fileNav

		// Select first torrent file
		torrentIdx := -1
		for i, entry := range nav.filtered {
			if !entry.isDir && filepath.Ext(entry.name) == ".torrent" {
				torrentIdx = i
				break
			}
		}
		assert.NotEqual(t, -1, torrentIdx)

		nav.selectedIdx = torrentIdx
		// selectedFile would be: nav.filtered[torrentIdx].fullPath

		// MockClient will handle the add torrent call

		// Select file
		cmd := view.handleFileNavigatorKeys("enter")
		assert.NotNil(t, cmd)

		// Dialog should close after successful selection
		msg := cmd()
		_, isSuccess := msg.(successMsg)
		assert.True(t, isSuccess)
	})
}

// TestURLInput tests URL input functionality
func TestURLInput(t *testing.T) {
	view, _ := createTestMainView()
	view.addDialog = &AddTorrentDialog{
		mode:    ModeURL,
		fileNav: NewFileNavigator("."),
		urlInput: &URLInput{
			url:    "",
			cursor: 0,
		},
	}
	view.showAddDialog = true

	t.Run("text input", func(t *testing.T) {
		urlInput := view.addDialog.urlInput

		// Type URL
		keys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune("h")},
			{Type: tea.KeyRunes, Runes: []rune("t")},
			{Type: tea.KeyRunes, Runes: []rune("t")},
			{Type: tea.KeyRunes, Runes: []rune("p")},
			{Type: tea.KeyRunes, Runes: []rune(":")},
			{Type: tea.KeyRunes, Runes: []rune("/")},
			{Type: tea.KeyRunes, Runes: []rune("/")},
		}

		for _, key := range keys {
			view.handleURLInputKeys(key)
		}

		assert.Equal(t, "http://", urlInput.url)

		// Continue typing
		view.handleURLInputKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("example.com/file.torrent")})
		assert.Equal(t, "http://example.com/file.torrent", urlInput.url)
	})

	t.Run("backspace", func(t *testing.T) {
		urlInput := view.addDialog.urlInput
		urlInput.url = "test"

		view.handleURLInputKeys(tea.KeyMsg{Type: tea.KeyBackspace})
		assert.Equal(t, "tes", urlInput.url)

		view.handleURLInputKeys(tea.KeyMsg{Type: tea.KeyBackspace})
		assert.Equal(t, "te", urlInput.url)
	})

	t.Run("clear all", func(t *testing.T) {
		urlInput := view.addDialog.urlInput
		urlInput.url = "long test string"

		view.handleURLInputKeys(tea.KeyMsg{Type: tea.KeyCtrlA})
		assert.Equal(t, "", urlInput.url)
	})

	t.Run("paste simulation", func(t *testing.T) {
		urlInput := view.addDialog.urlInput
		urlInput.url = ""

		// Simulate paste with multi-character input
		pasteMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("https://example.com/ubuntu.iso.torrent"),
		}

		view.handleURLInputKeys(pasteMsg)
		assert.Equal(t, "https://example.com/ubuntu.iso.torrent", urlInput.url)
	})

	t.Run("URL submission", func(t *testing.T) {
		// MockClient handles the torrent URL operations
		urlInput := view.addDialog.urlInput
		urlInput.url = "https://example.com/test.torrent"

		// MockClient will handle the add torrent call

		// Submit URL
		cmd := view.handleURLInputKeys(tea.KeyMsg{Type: tea.KeyEnter})
		assert.NotNil(t, cmd)

		// Should get success message
		msg := cmd()
		_, isSuccess := msg.(successMsg)
		assert.True(t, isSuccess)
	})

	t.Run("empty URL submission", func(t *testing.T) {
		urlInput := view.addDialog.urlInput
		urlInput.url = ""

		// Try to submit empty URL
		cmd := view.handleURLInputKeys(tea.KeyMsg{Type: tea.KeyEnter})
		assert.Nil(t, cmd) // Should not submit
	})
}

// TestAddTorrentDialog tests the add torrent dialog
func TestAddTorrentDialog(t *testing.T) {
	t.Run("mode switching", func(t *testing.T) {
		dialog := NewAddTorrentDialog(".")

		assert.Equal(t, ModeFile, dialog.mode)

		// Switch to URL mode
		dialog.mode = ModeURL
		assert.Equal(t, ModeURL, dialog.mode)

		// Switch back
		dialog.mode = ModeFile
		assert.Equal(t, ModeFile, dialog.mode)
	})

	t.Run("file entry formatting", func(t *testing.T) {
		view, _ := createTestMainView()

		tests := []struct {
			name     string
			entry    FileEntry
			selected bool
			expected []string
		}{
			{
				name: "directory",
				entry: FileEntry{
					name:  "Downloads",
					isDir: true,
				},
				selected: false,
				expected: []string{"📁", "Downloads", "DIR"},
			},
			{
				name: "parent directory",
				entry: FileEntry{
					name:  "..",
					isDir: true,
				},
				selected: false,
				expected: []string{"📁", ".."},
			},
			{
				name: "torrent file",
				entry: FileEntry{
					name:  "ubuntu.iso.torrent",
					isDir: false,
					size:  12345,
				},
				selected: true,
				expected: []string{"📄", "ubuntu.iso.torrent", "12.1 KB"},
			},
			{
				name: "long filename",
				entry: FileEntry{
					name:  "this-is-a-very-long-filename-that-should-be-truncated-properly.torrent",
					isDir: false,
					size:  1048576,
				},
				selected: false,
				expected: []string{"📄", "...", "1.0 MB"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				formatted := view.formatFileEntry(tt.entry, tt.selected)

				for _, exp := range tt.expected {
					assert.Contains(t, formatted, exp)
				}
			})
		}
	})
}

// TestDialogRendering tests dialog view rendering
func TestDialogRendering(t *testing.T) {
	view, _ := createTestMainView()
	view.width = 120
	view.height = 40

	t.Run("add dialog file mode", func(t *testing.T) {
		view.showAddDialog = true
		view.addDialog.mode = ModeFile

		output := view.View()
		assert.Contains(t, output, "Add Torrent")
		assert.Contains(t, output, "[File Browser]")
		assert.Contains(t, output, "📁")
		assert.Contains(t, output, "Tab: Switch mode")
	})

	t.Run("add dialog URL mode", func(t *testing.T) {
		view.showAddDialog = true
		view.addDialog.mode = ModeURL
		view.addDialog.urlInput.url = "https://example.com/"

		output := view.View()
		assert.Contains(t, output, "Add Torrent")
		assert.Contains(t, output, "[URL]")
		assert.Contains(t, output, "https://example.com/")
		assert.Contains(t, output, "Enter URL to .torrent file")
	})

	t.Run("delete dialog", func(t *testing.T) {
		view.showAddDialog = false
		view.showDeleteDialog = true
		view.deleteTarget = &api.Torrent{
			Name: "Test Torrent",
			Hash: "abc123",
		}
		view.deleteWithFiles = false

		output := view.View()
		assert.Contains(t, output, "Delete Torrent")
		assert.Contains(t, output, "Test Torrent")
		assert.Contains(t, output, "Delete files from disk")
		assert.Contains(t, output, "Y/Enter: Confirm")
		assert.Contains(t, output, "N/Esc: Cancel")
		assert.Contains(t, output, "F: Toggle files")

		// Toggle file deletion
		view.deleteWithFiles = true
		output = view.View()
		assert.Contains(t, output, "[✓] Delete files from disk")
	})
}

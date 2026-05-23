package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/config"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/components"
)

func TestAppendPrintable(t *testing.T) {
	tests := []struct {
		name string
		dst  string
		s    string
		want string
	}{
		{"empty", "", "", ""},
		{"magnet link", "", "magnet:?xt=urn:btih:abc123&dn=test", "magnet:?xt=urn:btih:abc123&dn=test"},
		{"appends to existing", "http://", "example.com/file.torrent", "http://example.com/file.torrent"},
		{"strips non-ASCII", "", "hello\x00world\x01!", "helloworld!"},
		{"strips high unicode", "", "testéfile", "testfile"},
		{"preserves URL special chars", "", "https://a.co/path?q=1&x=2#frag", "https://a.co/path?q=1&x=2#frag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendPrintable(tt.dst, tt.s)
			if got != tt.want {
				t.Errorf("appendPrintable(%q, %q) = %q, want %q", tt.dst, tt.s, got, tt.want)
			}
		})
	}
}

func newTestMainView() *MainView {
	cfg := &config.Config{}
	return &MainView{
		config:      cfg,
		torrentList: components.NewTorrentListWithColumns(nil, "", ""),
		statsPanel:  components.NewStatsPanel(),
		filterPanel: components.NewFilterPanel(),
		keys:        DefaultKeyMap(),
		viewMode:    ViewModeMain,
		addDialog: &AddTorrentDialog{
			mode:     ModeURL,
			urlInput: &URLInput{url: "", cursor: 0},
		},
		torrentMap: make(map[string]api.Torrent),
	}
}

func TestPasteIntoURLInput(t *testing.T) {
	m := newTestMainView()
	m.showAddDialog = true
	m.addDialog.mode = ModeURL

	magnet := "magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a&dn=test"
	m.Update(tea.PasteMsg{Content: magnet})

	if m.addDialog.urlInput.url != magnet {
		t.Errorf("paste into URL input: got %q, want %q", m.addDialog.urlInput.url, magnet)
	}
}

func TestPasteIntoURLInputAppends(t *testing.T) {
	m := newTestMainView()
	m.showAddDialog = true
	m.addDialog.mode = ModeURL
	m.addDialog.urlInput.url = "http://"

	m.Update(tea.PasteMsg{Content: "example.com/file.torrent"})

	want := "http://example.com/file.torrent"
	if m.addDialog.urlInput.url != want {
		t.Errorf("paste append: got %q, want %q", m.addDialog.urlInput.url, want)
	}
}

func TestPasteIntoPathInput(t *testing.T) {
	m := newTestMainView()
	m.showLocationDialog = true
	m.locationDialog = &LocationDialog{
		mode:      LocationModeText,
		pathInput: &PathInput{path: "", cursor: 0},
	}

	path := "/mnt/data/downloads"
	m.Update(tea.PasteMsg{Content: path})

	if m.locationDialog.pathInput.path != path {
		t.Errorf("paste into path input: got %q, want %q", m.locationDialog.pathInput.path, path)
	}
	if m.locationDialog.pathInput.cursor != len(path) {
		t.Errorf("cursor after paste: got %d, want %d", m.locationDialog.pathInput.cursor, len(path))
	}
}

func TestPasteIgnoredWhenNoInputActive(t *testing.T) {
	m := newTestMainView()
	m.showAddDialog = false
	m.showLocationDialog = false

	m.Update(tea.PasteMsg{Content: "should be ignored"})

	if m.addDialog.urlInput.url != "" {
		t.Error("paste should be ignored when no dialog is open")
	}
}

func TestPasteIgnoredInFileMode(t *testing.T) {
	m := newTestMainView()
	m.showAddDialog = true
	m.addDialog.mode = ModeFile

	m.Update(tea.PasteMsg{Content: "should be ignored"})

	if m.addDialog.urlInput.url != "" {
		t.Error("paste should be ignored when add dialog is in file mode")
	}
}

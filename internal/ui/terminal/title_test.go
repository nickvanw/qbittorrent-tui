package terminal

import (
	"strings"
	"testing"

	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTitle(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     TitleData
		contains []string // Strings that should be in the result
		wantErr  bool
	}{
		{
			name:     "simple speed template",
			template: "qbt - ↓{dl_speed} ↑{up_speed}",
			data: TitleData{
				DlSpeed: 1024 * 1024, // 1 MB/s
				UpSpeed: 512 * 1024,  // 512 KB/s
			},
			contains: []string{"qbt - ↓", "↑"},
			wantErr:  false,
		},
		{
			name:     "torrent counts template",
			template: "{active_torrents}/{total_torrents} active",
			data: TitleData{
				ActiveTorrents: 5,
				TotalTorrents:  20,
			},
			contains: []string{"5/20 active"},
			wantErr:  false,
		},
		{
			name:     "comprehensive template",
			template: "qbt-tui - ↓{dl_speed} ↑{up_speed} - {active_torrents} active - {server_url}",
			data: TitleData{
				DlSpeed:        2048 * 1024,
				UpSpeed:        1024 * 1024,
				ServerURL:      "localhost:8080",
				ActiveTorrents: 3,
			},
			contains: []string{"qbt-tui - ↓", "↑", "3 active", "localhost:8080"},
			wantErr:  false,
		},
		{
			name:     "all count variables",
			template: "D:{dl_torrents} U:{up_torrents} P:{paused_torrents} T:{total_torrents}",
			data: TitleData{
				DlTorrents:     2,
				UpTorrents:     5,
				PausedTorrents: 3,
				TotalTorrents:  10,
			},
			contains: []string{"D:2", "U:5", "P:3", "T:10"},
			wantErr:  false,
		},
		{
			name:     "session totals",
			template: "Session: ↓{session_downloaded} ↑{session_uploaded}",
			data: TitleData{
				SessionDownloaded: 5 * 1024 * 1024 * 1024, // 5 GB
				SessionUploaded:   2 * 1024 * 1024 * 1024, // 2 GB
			},
			contains: []string{"Session: ↓", "↑"},
			wantErr:  false,
		},
		{
			name:     "all variables combined",
			template: "{server_url} [{active_torrents}/{total_torrents}] ↓{dl_speed} ↑{up_speed} D:{dl_torrents} U:{up_torrents} P:{paused_torrents} Session:↓{session_downloaded}↑{session_uploaded}",
			data: TitleData{
				DlSpeed:           1024 * 1024,
				UpSpeed:           512 * 1024,
				SessionDownloaded: 1024 * 1024 * 1024,
				SessionUploaded:   512 * 1024 * 1024,
				ServerURL:         "http://localhost:8080",
				ActiveTorrents:    3,
				TotalTorrents:     10,
				DlTorrents:        2,
				UpTorrents:        5,
				PausedTorrents:    3,
			},
			contains: []string{
				"http://localhost:8080",
				"[3/10]",
				"D:2", "U:5", "P:3",
			},
			wantErr: false,
		},
		{
			name:     "no variables static title",
			template: "qBittorrent TUI",
			data:     TitleData{},
			contains: []string{"qBittorrent TUI"},
			wantErr:  false,
		},
		{
			name:     "zero values",
			template: "{active_torrents}/{total_torrents}",
			data: TitleData{
				ActiveTorrents: 0,
				TotalTorrents:  0,
			},
			contains: []string{"0/0"},
			wantErr:  false,
		},
		{
			name:     "empty template error",
			template: "",
			data:     TitleData{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTitle(tt.template, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				for _, expected := range tt.contains {
					assert.Contains(t, result, expected)
				}
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid template with speeds",
			template: "↓{dl_speed} ↑{up_speed}",
			wantErr:  false,
		},
		{
			name:     "valid template with counts",
			template: "{active_torrents}/{total_torrents}",
			wantErr:  false,
		},
		{
			name:     "valid template with all variables",
			template: "{dl_speed} {up_speed} {server_url} {active_torrents} {total_torrents} {dl_torrents} {up_torrents} {paused_torrents} {session_downloaded} {session_uploaded}",
			wantErr:  false,
		},
		{
			name:     "invalid variable",
			template: "qbt - {invalid_var}",
			wantErr:  true,
			errMsg:   "unknown variable: {invalid_var}",
		},
		{
			name:     "multiple invalid variables",
			template: "{foo} and {bar}",
			wantErr:  true,
			errMsg:   "unknown variable:",
		},
		{
			name:     "empty template",
			template: "",
			wantErr:  false,
		},
		{
			name:     "no variables",
			template: "qBittorrent TUI",
			wantErr:  false,
		},
		{
			name:     "partial variable match should fail",
			template: "{dl_speed_extra}",
			wantErr:  true,
			errMsg:   "unknown variable: {dl_speed_extra}",
		},
		{
			name:     "typo in variable name",
			template: "{dlspeed}",
			wantErr:  true,
			errMsg:   "unknown variable: {dlspeed}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.template)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateTorrentCounts(t *testing.T) {
	tests := []struct {
		name         string
		torrents     []api.Torrent
		wantActive   int
		wantDownload int
		wantUpload   int
		wantPaused   int
	}{
		{
			name: "mixed states",
			torrents: []api.Torrent{
				{State: "downloading"},
				{State: "downloading"},
				{State: "uploading"},
				{State: "uploading"},
				{State: "uploading"},
				{State: "pausedDL"},
				{State: "pausedUP"},
				{State: "stalledUP"},
			},
			wantActive:   6, // 2 downloading + 3 uploading + 1 stalledUP (stalledUP counts as uploading and active)
			wantDownload: 2,
			wantUpload:   4, // 3 uploading + 1 stalledUP
			wantPaused:   2,
		},
		{
			name: "all downloading",
			torrents: []api.Torrent{
				{State: "downloading"},
				{State: "downloading"},
				{State: "downloading"},
			},
			wantActive:   3,
			wantDownload: 3,
			wantUpload:   0,
			wantPaused:   0,
		},
		{
			name: "all paused",
			torrents: []api.Torrent{
				{State: "pausedDL"},
				{State: "pausedUP"},
			},
			wantActive:   0,
			wantDownload: 0,
			wantUpload:   0,
			wantPaused:   2,
		},
		{
			name:         "empty list",
			torrents:     []api.Torrent{},
			wantActive:   0,
			wantDownload: 0,
			wantUpload:   0,
			wantPaused:   0,
		},
		{
			name: "seeding torrents",
			torrents: []api.Torrent{
				{State: "uploading"},
				{State: "uploading"},
				{State: "uploading"},
				{State: "uploading"},
				{State: "stalledUP"},
			},
			wantActive:   5, // all 5 are active (stalledUP counts as uploading and thus active)
			wantDownload: 0,
			wantUpload:   5, // all of them are uploading (including stalledUP)
			wantPaused:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			active, dl, up, paused := CalculateTorrentCounts(tt.torrents)

			assert.Equal(t, tt.wantActive, active, "active count mismatch")
			assert.Equal(t, tt.wantDownload, dl, "download count mismatch")
			assert.Equal(t, tt.wantUpload, up, "upload count mismatch")
			assert.Equal(t, tt.wantPaused, paused, "paused count mismatch")
		})
	}
}

func TestSetTerminalTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "simple title",
			title: "Test Title",
		},
		{
			name:  "title with special characters",
			title: "qbt-tui - ↓1.5 MB/s ↑512 KB/s",
		},
		{
			name:  "empty title",
			title: "",
		},
		{
			name:  "title with unicode",
			title: "qBittorrent 🚀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetTerminalTitle(tt.title)

			// Should contain ANSI escape sequence
			assert.True(t, strings.HasPrefix(result, "\033]0;"))
			assert.True(t, strings.HasSuffix(result, "\007"))
			assert.Contains(t, result, tt.title)

			// Should be the expected format: ESC]0;title\007
			expected := "\033]0;" + tt.title + "\007"
			assert.Equal(t, expected, result)
		})
	}
}

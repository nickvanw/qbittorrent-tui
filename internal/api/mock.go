package api

import (
	"context"
	"fmt"
	"time"
)

// MockClient provides a mock implementation of the qBittorrent API client for testing
type MockClient struct {
	Torrents          []Torrent
	GlobalStats       *GlobalStats
	TorrentProperties map[string]*TorrentProperties
	Categories        map[string]interface{}
	Tags              []string
	LoginError        error
	GetError          error
	LoggedIn          bool
}

// NewMockClient creates a new mock client with default test data
func NewMockClient() *MockClient {
	return &MockClient{
		Torrents:          []Torrent{},
		TorrentProperties: make(map[string]*TorrentProperties),
		Categories:        make(map[string]interface{}),
		Tags:              []string{},
		GlobalStats: &GlobalStats{
			DlInfoSpeed:      1024 * 1024,
			UpInfoSpeed:      512 * 1024,
			NumTorrents:      10,
			NumActiveItems:   3,
			ConnectionStatus: "connected",
			DHTNodes:         150,
			TorrentsCount:    10,
			FreeSpaceOnDisk:  1024 * 1024 * 1024 * 500, // 500GB
		},
	}
}

func (m *MockClient) Login(username, password string) error {
	if m.LoginError != nil {
		return m.LoginError
	}
	m.LoggedIn = true
	return nil
}

func (m *MockClient) GetTorrents(ctx context.Context) ([]Torrent, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	return m.Torrents, nil
}

func (m *MockClient) GetTorrentsFiltered(ctx context.Context, filter map[string]string) ([]Torrent, error) {
	return m.GetTorrents(ctx)
}

func (m *MockClient) GetGlobalStats(ctx context.Context) (*GlobalStats, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	return m.GlobalStats, nil
}

func (m *MockClient) GetTorrentProperties(ctx context.Context, hash string) (*TorrentProperties, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}

	props, ok := m.TorrentProperties[hash]
	if !ok {
		return nil, fmt.Errorf("torrent not found")
	}
	return props, nil
}

func (m *MockClient) GetCategories(ctx context.Context) (map[string]interface{}, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	return m.Categories, nil
}

func (m *MockClient) GetTags(ctx context.Context) ([]string, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	return m.Tags, nil
}

// GenerateMockTorrents creates a slice of mock torrents for testing
func GenerateMockTorrents(count int) []Torrent {
	states := []TorrentState{
		StateDownloading,
		StateUploading,
		StatePausedDL,
		StatePausedUP,
		StateStalledDL,
		StateStalledUP,
	}

	categories := []string{"", "movies", "tv", "music", "software"}
	trackers := []string{
		"https://tracker1.com/announce",
		"https://tracker2.org/announce",
		"https://tracker3.net/announce",
		"",
	}
	tags := []string{"", "hd", "4k", "favorite", "new"}

	torrents := make([]Torrent, count)
	for i := 0; i < count; i++ {
		state := states[i%len(states)]
		progress := 0.0
		dlSpeed := int64(0)
		upSpeed := int64(0)

		switch state {
		case StateDownloading:
			progress = float64(i%100) / 100.0
			dlSpeed = int64(1024 * 1024 * (i%10 + 1))
		case StateUploading:
			progress = 1.0
			upSpeed = int64(512 * 1024 * (i%5 + 1))
		case StateStalledDL:
			progress = 0.5
		case StateStalledUP:
			progress = 1.0
		}

		torrents[i] = Torrent{
			Hash:          fmt.Sprintf("%040x", i),
			Name:          fmt.Sprintf("Test Torrent %d", i+1),
			Size:          int64(1024 * 1024 * 1024 * (i%10 + 1)),
			Progress:      progress,
			DlSpeed:       dlSpeed,
			UpSpeed:       upSpeed,
			NumSeeds:      i % 50,
			NumLeeches:    i % 30,
			NumComplete:   i % 50,
			NumIncomplete: i % 30,
			Ratio:         float64(i%200) / 100.0,
			State:         state.String(),
			Category:      categories[i%len(categories)],
			Tags:          tags[i%len(tags)],
			Tracker:       trackers[i%len(trackers)],
			AddedOn:       time.Now().Add(-time.Duration(i*24) * time.Hour).Unix(),
		}
	}

	return torrents
}

// SetupMockClientWithData creates a mock client with predefined test data
func SetupMockClientWithData() *MockClient {
	client := NewMockClient()
	client.Torrents = GenerateMockTorrents(5)
	client.Categories = map[string]interface{}{
		"movies": map[string]interface{}{"name": "Movies", "savePath": "/downloads/movies"},
		"tv":     map[string]interface{}{"name": "TV Shows", "savePath": "/downloads/tv"},
	}
	client.Tags = []string{"hd", "4k", "favorite"}
	return client
}

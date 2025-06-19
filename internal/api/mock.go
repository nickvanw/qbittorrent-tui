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
	Trackers          map[string][]Tracker
	Peers             map[string]map[string]Peer
	Files             map[string][]TorrentFile
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
		Trackers:          make(map[string][]Tracker),
		Peers:             make(map[string]map[string]Peer),
		Files:             make(map[string][]TorrentFile),
		GlobalStats: &GlobalStats{
			DlInfoSpeed:      1024 * 1024,
			UpInfoSpeed:      512 * 1024,
			ConnectionStatus: "connected",
			DHTNodes:         150,
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

func (m *MockClient) GetTorrentTrackers(ctx context.Context, hash string) ([]Tracker, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	if trackers, exists := m.Trackers[hash]; exists {
		return trackers, nil
	}
	// Return mock data if no specific data is set
	return []Tracker{
		{
			URL:           "http://tracker.example.com:8080/announce",
			Status:        2, // Working
			Tier:          0,
			NumPeers:      42,
			NumSeeds:      15,
			NumLeeches:    27,
			NumDownloaded: 125,
			Msg:           "Working",
		},
	}, nil
}

func (m *MockClient) GetTorrentPeers(ctx context.Context, hash string) (map[string]Peer, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	if peers, exists := m.Peers[hash]; exists {
		return peers, nil
	}
	// Return mock data if no specific data is set
	return map[string]Peer{
		"192.168.1.100:51413": {
			IP:         "192.168.1.100",
			Port:       51413,
			Country:    "US",
			Connection: "uTP",
			Flags:      "u D",
			Client:     "qBittorrent/4.5.0",
			Progress:   0.75,
			DlSpeed:    1024 * 100,
			UpSpeed:    1024 * 50,
			Downloaded: 1024 * 1024 * 500,
			Uploaded:   1024 * 1024 * 200,
			Relevance:  1.0,
		},
	}, nil
}

func (m *MockClient) GetTorrentFiles(ctx context.Context, hash string) ([]TorrentFile, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if !m.LoggedIn {
		return nil, fmt.Errorf("authentication required")
	}
	if files, exists := m.Files[hash]; exists {
		return files, nil
	}
	// Return mock data if no specific data is set
	return []TorrentFile{
		{
			Index:        0,
			Name:         "ubuntu-22.04.3-desktop-amd64.iso",
			Size:         4800000000, // ~4.8GB
			Progress:     0.85,
			Priority:     1, // Normal priority
			IsSeed:       false,
			PieceRange:   []int{0, 1199},
			Availability: 0.95,
		},
		{
			Index:        1,
			Name:         "README.txt",
			Size:         2048,
			Progress:     1.0,
			Priority:     1,
			IsSeed:       true,
			PieceRange:   []int{1200, 1200},
			Availability: 1.0,
		},
	}, nil
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

// PauseTorrents simulates pausing torrents
func (m *MockClient) PauseTorrents(ctx context.Context, hashes []string) error {
	if m.GetError != nil {
		return m.GetError
	}
	if !m.LoggedIn {
		return fmt.Errorf("authentication required")
	}
	// Mock implementation - in real usage this would pause the torrents
	return nil
}

// ResumeTorrents simulates resuming torrents
func (m *MockClient) ResumeTorrents(ctx context.Context, hashes []string) error {
	if m.GetError != nil {
		return m.GetError
	}
	if !m.LoggedIn {
		return fmt.Errorf("authentication required")
	}
	// Mock implementation - in real usage this would resume the torrents
	return nil
}

// DeleteTorrents simulates deleting torrents
func (m *MockClient) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	if m.GetError != nil {
		return m.GetError
	}
	if !m.LoggedIn {
		return fmt.Errorf("authentication required")
	}
	// Mock implementation - in real usage this would delete the torrents
	return nil
}

// AddTorrentFile simulates adding a torrent from a file
func (m *MockClient) AddTorrentFile(ctx context.Context, filePath string) error {
	if m.GetError != nil {
		return m.GetError
	}
	if !m.LoggedIn {
		return fmt.Errorf("authentication required")
	}
	// Mock implementation - in real usage this would add the torrent file
	return nil
}

// AddTorrentURL simulates adding a torrent from a URL
func (m *MockClient) AddTorrentURL(ctx context.Context, url string) error {
	if m.GetError != nil {
		return m.GetError
	}
	if !m.LoggedIn {
		return fmt.Errorf("authentication required")
	}
	// Mock implementation - in real usage this would add the torrent from URL
	return nil
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

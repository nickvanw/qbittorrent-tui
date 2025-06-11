package api

import "context"

// ClientInterface defines the interface for qBittorrent API clients
type ClientInterface interface {
	// Authentication
	Login(username, password string) error

	// Torrent operations
	GetTorrents(ctx context.Context) ([]Torrent, error)
	GetTorrentsFiltered(ctx context.Context, filter map[string]string) ([]Torrent, error)
	GetTorrentProperties(ctx context.Context, hash string) (*TorrentProperties, error)

	// Torrent details
	GetTorrentTrackers(ctx context.Context, hash string) ([]Tracker, error)
	GetTorrentPeers(ctx context.Context, hash string) (map[string]Peer, error)
	GetTorrentFiles(ctx context.Context, hash string) ([]TorrentFile, error)

	// Global operations
	GetGlobalStats(ctx context.Context) (*GlobalStats, error)
	GetCategories(ctx context.Context) (map[string]interface{}, error)
	GetTags(ctx context.Context) ([]string, error)
}

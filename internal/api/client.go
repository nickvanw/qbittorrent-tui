package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cookie     string
}

func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) Login(username, password string) error {
	data := url.Values{
		"username": {username},
		"password": {password},
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v2/auth/login", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	responseStr := strings.TrimSpace(string(body))
	if responseStr == "Fails." {
		return fmt.Errorf("invalid username or password")
	}

	for _, cookie := range c.httpClient.Jar.Cookies(resp.Request.URL) {
		if cookie.Name == "SID" {
			c.cookie = cookie.Value
			return nil
		}
	}

	return nil
}

func (c *Client) GetTorrents(ctx context.Context) ([]Torrent, error) {
	return c.GetTorrentsFiltered(ctx, nil)
}

func (c *Client) GetTorrentsFiltered(ctx context.Context, filter map[string]string) ([]Torrent, error) {
	endpoint := "/api/v2/torrents/info"
	if len(filter) > 0 {
		params := url.Values{}
		for k, v := range filter {
			params.Set(k, v)
		}
		endpoint += "?" + params.Encode()
	}

	var torrents []Torrent
	if err := c.get(ctx, endpoint, &torrents); err != nil {
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	return torrents, nil
}

func (c *Client) GetGlobalStats(ctx context.Context) (*GlobalStats, error) {
	// Get transfer info
	var stats GlobalStats
	if err := c.get(ctx, "/api/v2/transfer/info", &stats); err != nil {
		return nil, fmt.Errorf("failed to get transfer info: %w", err)
	}

	// Get maindata for free disk space
	var mainData MainData
	if err := c.get(ctx, "/api/v2/sync/maindata", &mainData); err != nil {
		// Log warning but don't fail - free space is not critical
		// In a real app you'd use a proper logger here
		// For now, just continue with stats.FreeSpaceOnDisk = 0
	} else {
		// Merge free space from maindata
		stats.FreeSpaceOnDisk = mainData.ServerState.FreeSpaceOnDisk
	}

	return &stats, nil
}

func (c *Client) GetTorrentProperties(ctx context.Context, hash string) (*TorrentProperties, error) {
	endpoint := fmt.Sprintf("/api/v2/torrents/properties?hash=%s", hash)

	var props TorrentProperties
	if err := c.get(ctx, endpoint, &props); err != nil {
		return nil, fmt.Errorf("failed to get torrent properties: %w", err)
	}

	return &props, nil
}

func (c *Client) GetCategories(ctx context.Context) (map[string]interface{}, error) {
	var categories map[string]interface{}
	if err := c.get(ctx, "/api/v2/torrents/categories", &categories); err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (c *Client) GetTags(ctx context.Context) ([]string, error) {
	var tags []string
	if err := c.get(ctx, "/api/v2/torrents/tags", &tags); err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

// GetTorrentTrackers retrieves trackers for a specific torrent
func (c *Client) GetTorrentTrackers(ctx context.Context, hash string) ([]Tracker, error) {
	endpoint := fmt.Sprintf("/api/v2/torrents/trackers?hash=%s", hash)
	var trackers []Tracker
	if err := c.get(ctx, endpoint, &trackers); err != nil {
		return nil, fmt.Errorf("failed to get torrent trackers: %w", err)
	}
	return trackers, nil
}

// GetTorrentPeers retrieves peers for a specific torrent
func (c *Client) GetTorrentPeers(ctx context.Context, hash string) (map[string]Peer, error) {
	endpoint := fmt.Sprintf("/api/v2/sync/torrentPeers?hash=%s", hash)

	// The API returns a complex object, we need to extract the peers map
	var response struct {
		Peers map[string]Peer `json:"peers"`
	}

	if err := c.get(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get torrent peers: %w", err)
	}

	return response.Peers, nil
}

// GetTorrentFiles retrieves files for a specific torrent
func (c *Client) GetTorrentFiles(ctx context.Context, hash string) ([]TorrentFile, error) {
	endpoint := fmt.Sprintf("/api/v2/torrents/files?hash=%s", hash)
	var files []TorrentFile
	if err := c.get(ctx, endpoint, &files); err != nil {
		return nil, fmt.Errorf("failed to get torrent files: %w", err)
	}
	return files, nil
}

// PauseTorrents pauses one or more torrents
func (c *Client) PauseTorrents(ctx context.Context, hashes []string) error {
	hashParam := strings.Join(hashes, "|")

	data := url.Values{
		"hashes": {hashParam},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v2/torrents/stop", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create pause request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("pause request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pause failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ResumeTorrents resumes one or more torrents
func (c *Client) ResumeTorrents(ctx context.Context, hashes []string) error {
	hashParam := strings.Join(hashes, "|")

	data := url.Values{
		"hashes": {hashParam},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v2/torrents/start", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create resume request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("resume request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resume failed with status: %d", resp.StatusCode)
	}

	return nil
}

// DeleteTorrents deletes one or more torrents
func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	hashParam := strings.Join(hashes, "|")
	deleteFilesParam := "false"
	if deleteFiles {
		deleteFilesParam = "true"
	}

	data := url.Values{
		"hashes":      {hashParam},
		"deleteFiles": {deleteFilesParam},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v2/torrents/delete", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

// AddTorrentFile adds a torrent from a local .torrent file
func (c *Client) AddTorrentFile(ctx context.Context, filePath string) error {
	// Read the torrent file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open torrent file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add the torrent file
	part, err := writer.CreateFormFile("torrents", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v2/torrents/add", &body)
	if err != nil {
		return fmt.Errorf("failed to create add request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Referer", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("add torrent request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add torrent failed with status: %d", resp.StatusCode)
	}

	return nil
}

// AddTorrentURL adds a torrent from a URL
func (c *Client) AddTorrentURL(ctx context.Context, torrentURL string) error {
	data := url.Values{
		"urls": {torrentURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v2/torrents/add", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create add URL request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("add torrent URL request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add torrent URL failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) get(ctx context.Context, endpoint string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Referer", c.baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication required (403 Forbidden)")
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("request failed with status %d (failed to read body: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

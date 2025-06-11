package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("http://localhost:8080")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.httpClient.Jar)
}

func TestLoginWithActualServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual server test in short mode")
	}

	// This test helps debug actual login issues
	serverURL := "http://localhost:8181"
	password := os.Getenv("QBT_TEST_PASSWORD")
	if password == "" {
		t.Skip("QBT_TEST_PASSWORD not set")
	}

	// Test with curl command to verify
	t.Logf("Testing with curl first...")
	cmd := exec.Command("curl", "-v", "-X", "POST",
		serverURL+"/api/v2/auth/login",
		"-H", "Referer: "+serverURL,
		"-d", "username=admin",
		"-d", "password="+password)

	output, _ := cmd.CombinedOutput()
	t.Logf("Curl output:\n%s", string(output))

	// Now test with our client
	client, err := NewClient(serverURL)
	require.NoError(t, err)

	err = client.Login("admin", password)
	if err != nil {
		t.Errorf("Login failed: %v", err)
	}
}

func TestClientLogin(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		response      string
		statusCode    int
		setCookie     bool
		expectError   bool
		errorContains string
	}{
		{
			name:       "successful login",
			username:   "admin",
			password:   "password",
			response:   "Ok.",
			statusCode: http.StatusOK,
			setCookie:  true,
		},
		{
			name:          "invalid credentials",
			username:      "admin",
			password:      "wrong",
			response:      "Fails.",
			statusCode:    http.StatusOK,
			setCookie:     false,
			expectError:   true,
			errorContains: "invalid username or password",
		},
		{
			name:          "server error",
			username:      "admin",
			password:      "password",
			response:      "Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "login failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v2/auth/login", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.username, r.FormValue("username"))
				assert.Equal(t, tt.password, r.FormValue("password"))

				if tt.setCookie {
					http.SetCookie(w, &http.Cookie{
						Name:  "SID",
						Value: "test-session-id",
						Path:  "/",
					})
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			require.NoError(t, err)

			err = client.Login(tt.username, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.setCookie {
					serverURL, err := url.Parse(server.URL)
					require.NoError(t, err)
					cookies := client.httpClient.Jar.Cookies(serverURL)
					found := false
					for _, c := range cookies {
						if c.Name == "SID" {
							found = true
							assert.Equal(t, "test-session-id", c.Value)
							break
						}
					}
					assert.True(t, found, "SID cookie not found")
				}
			}
		})
	}
}

func TestClientGetTorrents(t *testing.T) {
	mockTorrents := []Torrent{
		{
			Hash:     "abc123",
			Name:     "Test Torrent 1",
			Size:     1024 * 1024 * 1024,
			Progress: 0.5,
			State:    "downloading",
		},
		{
			Hash:     "def456",
			Name:     "Test Torrent 2",
			Size:     2 * 1024 * 1024 * 1024,
			Progress: 1.0,
			State:    "seeding",
		},
	}

	tests := []struct {
		name          string
		filter        map[string]string
		statusCode    int
		response      interface{}
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:          "successful get torrents",
			statusCode:    http.StatusOK,
			response:      mockTorrents,
			expectedCount: 2,
		},
		{
			name:          "successful with filter",
			filter:        map[string]string{"filter": "downloading"},
			statusCode:    http.StatusOK,
			response:      []Torrent{mockTorrents[0]},
			expectedCount: 1,
		},
		{
			name:          "authentication error",
			statusCode:    http.StatusForbidden,
			expectError:   true,
			errorContains: "authentication required",
		},
		{
			name:          "server error",
			statusCode:    http.StatusInternalServerError,
			response:      "Internal Server Error",
			expectError:   true,
			errorContains: "request failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v2/torrents/info", r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				if tt.filter != nil {
					for k, v := range tt.filter {
						assert.Equal(t, v, r.URL.Query().Get(k))
					}
				}

				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					if str, ok := tt.response.(string); ok {
						w.Write([]byte(str))
					} else {
						json.NewEncoder(w).Encode(tt.response)
					}
				}
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			require.NoError(t, err)

			ctx := context.Background()
			var torrents []Torrent

			if tt.filter != nil {
				torrents, err = client.GetTorrentsFiltered(ctx, tt.filter)
			} else {
				torrents, err = client.GetTorrents(ctx)
			}

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, torrents, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, mockTorrents[0].Hash, torrents[0].Hash)
				}
			}
		})
	}
}

func TestClientGetGlobalStats(t *testing.T) {
	mockStats := &GlobalStats{
		DlInfoSpeed:    1024 * 1024,
		UpInfoSpeed:    512 * 1024,
		NumTorrents:    10,
		NumActiveItems: 3,
	}

	tests := []struct {
		name          string
		statusCode    int
		response      interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:       "successful get stats",
			statusCode: http.StatusOK,
			response:   mockStats,
		},
		{
			name:          "authentication error",
			statusCode:    http.StatusForbidden,
			expectError:   true,
			errorContains: "authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)

				// Handle both endpoints that GetGlobalStats calls
				if r.URL.Path == "/api/v2/transfer/info" {
					w.WriteHeader(tt.statusCode)
					if tt.response != nil {
						json.NewEncoder(w).Encode(tt.response)
					}
				} else if r.URL.Path == "/api/v2/sync/maindata" {
					// Return maindata with free space
					mainData := map[string]interface{}{
						"server_state": map[string]interface{}{
							"free_space_on_disk": int64(1024 * 1024 * 1024 * 100), // 100GB
						},
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(mainData)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			require.NoError(t, err)

			ctx := context.Background()
			stats, err := client.GetGlobalStats(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, mockStats.DlInfoSpeed, stats.DlInfoSpeed)
				assert.Equal(t, mockStats.NumTorrents, stats.NumTorrents)
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)
	client.httpClient.Timeout = 100 * time.Millisecond

	ctx := context.Background()
	_, err = client.GetTorrents(ctx)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "context deadline exceeded") ||
		strings.Contains(err.Error(), "Client.Timeout exceeded"))
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = client.GetTorrents(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClientGetTorrentProperties(t *testing.T) {
	mockProps := &TorrentProperties{
		SavePath:    "/downloads/",
		TotalSize:   1024 * 1024 * 1024,
		DlSpeed:     1024 * 1024,
		UpSpeed:     512 * 1024,
		TimeElapsed: 3600,
		ShareRatio:  1.5,
		Peers:       10,
		Seeds:       5,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/torrents/properties", r.URL.Path)
		assert.Equal(t, "abc123", r.URL.Query().Get("hash"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockProps)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	props, err := client.GetTorrentProperties(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, mockProps.SavePath, props.SavePath)
	assert.Equal(t, mockProps.ShareRatio, props.ShareRatio)
}

func TestClientGetCategories(t *testing.T) {
	mockCategories := map[string]interface{}{
		"movies": map[string]interface{}{
			"name":     "movies",
			"savePath": "/downloads/movies",
		},
		"tv": map[string]interface{}{
			"name":     "tv",
			"savePath": "/downloads/tv",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/torrents/categories", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockCategories)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	categories, err := client.GetCategories(ctx)
	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Contains(t, categories, "movies")
	assert.Contains(t, categories, "tv")
}

func TestClientGetTags(t *testing.T) {
	mockTags := []string{"hd", "4k", "favorite", "new"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/torrents/tags", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockTags)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	tags, err := client.GetTags(ctx)
	assert.NoError(t, err)
	assert.Equal(t, mockTags, tags)
}

func TestTorrentStateHelpers(t *testing.T) {
	tests := []struct {
		state         TorrentState
		isDownloading bool
		isUploading   bool
		isPaused      bool
		isActive      bool
	}{
		{StateDownloading, true, false, false, true},
		{StateMetaDL, true, false, false, true},
		{StateForcedDL, true, false, false, true},
		{StateAllocating, true, false, false, true},
		{StateUploading, false, true, false, true},
		{StateForcedUP, false, true, false, true},
		{StateStalledUP, false, true, false, true},
		{StatePausedDL, false, false, true, false},
		{StatePausedUP, false, false, true, false},
		{StateQueuedDL, false, false, false, false},
		{StateError, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			assert.Equal(t, string(tt.state), tt.state.String())
			assert.Equal(t, tt.isDownloading, tt.state.IsDownloading())
			assert.Equal(t, tt.isUploading, tt.state.IsUploading())
			assert.Equal(t, tt.isPaused, tt.state.IsPaused())
			assert.Equal(t, tt.isActive, tt.state.IsActive())
		})
	}
}

func TestMockClient(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	// Test login
	err := mock.Login("admin", "password")
	assert.NoError(t, err)
	assert.True(t, mock.LoggedIn)

	// Test with login error
	mock.LoginError = fmt.Errorf("auth failed")
	err = mock.Login("admin", "wrong")
	assert.Error(t, err)
	assert.Equal(t, "auth failed", err.Error())

	// Reset for other tests
	mock.LoginError = nil
	mock.LoggedIn = true

	// Test GetTorrents
	mock.Torrents = GenerateMockTorrents(5)
	torrents, err := mock.GetTorrents(ctx)
	assert.NoError(t, err)
	assert.Len(t, torrents, 5)

	// Test GetGlobalStats
	stats, err := mock.GetGlobalStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Test GetCategories
	mock.Categories["test"] = map[string]interface{}{"name": "test"}
	categories, err := mock.GetCategories(ctx)
	assert.NoError(t, err)
	assert.Contains(t, categories, "test")

	// Test GetTags
	mock.Tags = []string{"tag1", "tag2"}
	tags, err := mock.GetTags(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"tag1", "tag2"}, tags)

	// Test auth required
	mock.LoggedIn = false
	_, err = mock.GetTorrents(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestGenerateMockTorrents(t *testing.T) {
	torrents := GenerateMockTorrents(10)
	assert.Len(t, torrents, 10)

	// Check that torrents have varied states
	states := make(map[string]bool)
	for _, torrent := range torrents {
		states[torrent.State] = true
		assert.NotEmpty(t, torrent.Hash)
		assert.NotEmpty(t, torrent.Name)
		assert.Greater(t, torrent.Size, int64(0))
	}
	assert.Greater(t, len(states), 1, "Should have multiple states")
}

func TestTorrentControlMethods(t *testing.T) {
	mock := NewMockClient()
	mock.LoggedIn = true
	ctx := context.Background()

	hashes := []string{"hash1", "hash2"}

	// Test PauseTorrents
	err := mock.PauseTorrents(ctx, hashes)
	assert.NoError(t, err)

	// Test ResumeTorrents
	err = mock.ResumeTorrents(ctx, hashes)
	assert.NoError(t, err)

	// Test DeleteTorrents without files
	err = mock.DeleteTorrents(ctx, hashes, false)
	assert.NoError(t, err)

	// Test DeleteTorrents with files
	err = mock.DeleteTorrents(ctx, hashes, true)
	assert.NoError(t, err)

	// Test authentication required
	mock.LoggedIn = false
	err = mock.PauseTorrents(ctx, hashes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")

	err = mock.ResumeTorrents(ctx, hashes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")

	err = mock.DeleteTorrents(ctx, hashes, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

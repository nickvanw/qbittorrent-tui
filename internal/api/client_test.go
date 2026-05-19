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

func TestNewClientWithAPIKey(t *testing.T) {
	client, err := NewClientWithAPIKey("http://localhost:8080/", "qbt_testkeytestkeytestkeytestkey")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL, "trailing slash should be trimmed")
	assert.NotNil(t, client.httpClient)
	assert.Nil(t, client.httpClient.Jar, "API-key clients should not attach a cookie jar")
	require.NotNil(t, client.httpClient.Transport, "transport should be wired for bearer auth")
	bt, ok := client.httpClient.Transport.(*bearerAuthTransport)
	require.True(t, ok, "transport should be *bearerAuthTransport")
	assert.Equal(t, "qbt_testkeytestkeytestkeytestkey", bt.key)
	assert.Equal(t, "localhost:8080", bt.host, "transport host gate should match base URL host")
}

func TestNewClientWithAPIKeyInvalidURL(t *testing.T) {
	_, err := NewClientWithAPIKey("not a url", "qbt_x")
	assert.Error(t, err)
}

// TestClientAPIKeyHeader verifies that a client built with NewClientWithAPIKey
// sends "Authorization: Bearer <key>" on every outgoing request — both reads
// and write actions — and that no session cookie is established (qBittorrent's
// API-key auth is stateless).
func TestClientAPIKeyHeader(t *testing.T) {
	const key = "qbt_testkeytestkeytestkeytestkey"
	wantAuth := "Bearer " + key

	cases := []struct {
		name string
		path string
		call func(c *Client, ctx context.Context) error
	}{
		{"GetTorrents", "/api/v2/torrents/info", func(c *Client, ctx context.Context) error {
			_, err := c.GetTorrents(ctx)
			return err
		}},
		{"PauseTorrents", "/api/v2/torrents/stop", func(c *Client, ctx context.Context) error {
			return c.PauseTorrents(ctx, []string{"hash1"})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sawAuth string
			var sawCookies []*http.Cookie
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.path, r.URL.Path)
				sawAuth = r.Header.Get("Authorization")
				sawCookies = r.Cookies()
				// GetTorrents needs a JSON body; the action endpoint is fine with 204.
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("[]"))
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client, err := NewClientWithAPIKey(server.URL, key)
			require.NoError(t, err)

			err = tc.call(client, context.Background())
			require.NoError(t, err)

			assert.Equal(t, wantAuth, sawAuth, "Authorization header missing or wrong")
			assert.Empty(t, sawCookies, "API-key auth must not send cookies")
		})
	}
}

// TestClientAPIKeyIgnoresSetCookie verifies that even if the server emits a
// Set-Cookie header, the API-key client neither stores it nor replays it on
// subsequent requests — the jar is nil, so there is no persistence surface.
func TestClientAPIKeyIgnoresSetCookie(t *testing.T) {
	const key = "qbt_testkeytestkeytestkeytestkey"

	var requestCount int
	var sawCookiesPerRequest [][]*http.Cookie
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		sawCookiesPerRequest = append(sawCookiesPerRequest, r.Cookies())
		// Try to plant a cookie. A nil jar should drop it on the floor.
		http.SetCookie(w, &http.Cookie{Name: "SID", Value: "should-not-persist", Path: "/"})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client, err := NewClientWithAPIKey(server.URL, key)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.GetTorrents(ctx)
	require.NoError(t, err)
	_, err = client.GetTorrents(ctx)
	require.NoError(t, err)

	require.Equal(t, 2, requestCount, "expected two server-observed requests")
	assert.Empty(t, sawCookiesPerRequest[0], "first request must not send cookies")
	assert.Empty(t, sawCookiesPerRequest[1], "second request must not replay a server-issued cookie")
	assert.Nil(t, client.httpClient.Jar, "client must not grow a cookie jar")
}

// TestClientAPIKeyHostMismatch verifies the transport drops the Authorization
// header for requests whose host doesn't match the configured base URL host.
// This protects against credential leaks on cross-host redirects, which Go's
// built-in protection cannot strip when the header is added inside a
// RoundTripper rather than on the original request.
func TestClientAPIKeyHostMismatch(t *testing.T) {
	const key = "qbt_testkeytestkeytestkeytestkey"

	tr := &bearerAuthTransport{
		key:  key,
		host: "qbittorrent.example.com:8080",
		base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: http.NoBody, Header: make(http.Header), Request: req}, nil
		}),
	}

	matching, _ := http.NewRequest(http.MethodGet, "http://qbittorrent.example.com:8080/api/v2/torrents/info", nil)
	resp, err := tr.RoundTrip(matching)
	require.NoError(t, err)
	assert.Equal(t, "Bearer "+key, resp.Request.Header.Get("Authorization"), "matching host should carry the bearer token")

	foreign, _ := http.NewRequest(http.MethodGet, "http://evil.example.com/steal", nil)
	resp, err = tr.RoundTrip(foreign)
	require.NoError(t, err)
	assert.Empty(t, resp.Request.Header.Get("Authorization"), "foreign host must not receive the bearer token")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

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
		cookieName    string
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
			cookieName: "SID",
		},
		{
			// qBittorrent 5.2.0+ returns 204 No Content with a port-suffixed
			// QBT_SID cookie instead of 200 OK + SID.
			name:       "successful login (qBittorrent 5.2.0)",
			username:   "admin",
			password:   "password",
			response:   "",
			statusCode: http.StatusNoContent,
			setCookie:  true,
			cookieName: "QBT_SID_8112",
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
			errorContains: "server error",
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
						Name:  tt.cookieName,
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
						if c.Name == tt.cookieName {
							found = true
							assert.Equal(t, "test-session-id", c.Value)
							break
						}
					}
					assert.True(t, found, "session cookie %q not found", tt.cookieName)
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
			errorContains: "server error",
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
		DlInfoSpeed:      1024 * 1024,
		UpInfoSpeed:      512 * 1024,
		ConnectionStatus: "connected",
		DHTNodes:         150,
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
				assert.Equal(t, mockStats.ConnectionStatus, stats.ConnectionStatus)
			}
		})
	}
}

// TestClientActionEndpointsAcceptStatus verifies that all action (write)
// endpoints accept both 200 OK (qBittorrent <5.2) and 204 No Content
// (qBittorrent 5.2.0+). qBittorrent 5.2 changed write endpoints to return
// 204 since they have no response body.
func TestClientActionEndpointsAcceptStatus(t *testing.T) {
	statusCodes := []int{http.StatusOK, http.StatusNoContent}

	actions := []struct {
		name string
		path string
		call func(c *Client, ctx context.Context) error
	}{
		{"PauseTorrents", "/api/v2/torrents/stop", func(c *Client, ctx context.Context) error {
			return c.PauseTorrents(ctx, []string{"hash1"})
		}},
		{"ResumeTorrents", "/api/v2/torrents/start", func(c *Client, ctx context.Context) error {
			return c.ResumeTorrents(ctx, []string{"hash1"})
		}},
		{"DeleteTorrents", "/api/v2/torrents/delete", func(c *Client, ctx context.Context) error {
			return c.DeleteTorrents(ctx, []string{"hash1"}, false)
		}},
		{"AddTorrentURL", "/api/v2/torrents/add", func(c *Client, ctx context.Context) error {
			return c.AddTorrentURL(ctx, "magnet:?xt=urn:btih:abc")
		}},
		{"SetTorrentLocation", "/api/v2/torrents/setLocation", func(c *Client, ctx context.Context) error {
			return c.SetTorrentLocation(ctx, []string{"hash1"}, "/downloads")
		}},
	}

	for _, action := range actions {
		for _, code := range statusCodes {
			t.Run(fmt.Sprintf("%s/%d", action.name, code), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, action.path, r.URL.Path)
					assert.Equal(t, "POST", r.Method)
					w.WriteHeader(code)
				}))
				defer server.Close()

				client, err := NewClient(server.URL)
				require.NoError(t, err)

				err = action.call(client, context.Background())
				assert.NoError(t, err)
			})
		}
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

	// Test SetTorrentLocation
	err = mock.SetTorrentLocation(ctx, hashes, "/new/location")
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

	err = mock.SetTorrentLocation(ctx, hashes, "/new/location")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestGetDirectoryContent(t *testing.T) {
	mock := NewMockClient()
	mock.LoggedIn = true
	ctx := context.Background()

	// Test root directory
	dirs, err := mock.GetDirectoryContent(ctx, "/", "dirs")
	assert.NoError(t, err)
	assert.NotEmpty(t, dirs)
	assert.Contains(t, dirs, "/downloads")

	// Test subdirectory
	dirs, err = mock.GetDirectoryContent(ctx, "/downloads", "dirs")
	assert.NoError(t, err)
	assert.NotEmpty(t, dirs)

	// Test authentication required
	mock.LoggedIn = false
	_, err = mock.GetDirectoryContent(ctx, "/", "dirs")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestSyncMainData(t *testing.T) {
	mock := NewMockClient()
	mock.LoggedIn = true
	mock.Torrents = GenerateMockTorrents(3)
	mock.Tags = []string{"test", "hd"}
	mock.Categories = map[string]interface{}{
		"movies": map[string]interface{}{"name": "Movies", "savePath": "/downloads/movies"},
	}
	ctx := context.Background()

	// Test first request (rid=0) should return full update
	syncData, err := mock.SyncMainData(ctx, 0)
	assert.NoError(t, err)
	assert.NotNil(t, syncData)
	assert.True(t, syncData.FullUpdate, "First request should be a full update")
	assert.Equal(t, 1, syncData.RID, "RID should be incremented")
	assert.Len(t, syncData.Torrents, 3, "Should return all 3 torrents")
	assert.Len(t, syncData.Tags, 2, "Should return all tags")
	assert.Len(t, syncData.Categories, 1, "Should return all categories")
	assert.NotNil(t, syncData.ServerState, "ServerState should be populated")

	// Test subsequent request (rid > 0) should return incremental update
	syncData2, err := mock.SyncMainData(ctx, syncData.RID)
	assert.NoError(t, err)
	assert.NotNil(t, syncData2)
	assert.False(t, syncData2.FullUpdate, "Subsequent request should be incremental")
	assert.Equal(t, 2, syncData2.RID, "RID should be incremented again")

	// Test authentication required
	mock.LoggedIn = false
	_, err = mock.SyncMainData(ctx, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")

	// Test with error
	mock.LoggedIn = true
	mock.GetError = assert.AnError
	_, err = mock.SyncMainData(ctx, 0)
	assert.Error(t, err)
}

func TestPartialTorrentApplyTo(t *testing.T) {
	// Test that ApplyTo correctly distinguishes between nil (not present) and zero values
	t.Run("nil fields are not applied", func(t *testing.T) {
		existing := Torrent{
			Hash:     "abc123",
			Name:     "Original Name",
			Size:     1000,
			Progress: 0.5,
			DlSpeed:  5000,
			UpSpeed:  2000,
			State:    "downloading",
		}

		// Partial with only DlSpeed set (to 0 - a valid value)
		zero := int64(0)
		partial := PartialTorrent{
			DlSpeed: &zero, // Explicitly set to 0
			// All other fields are nil (not present in JSON)
		}

		partial.ApplyTo(&existing)

		// DlSpeed should be updated to 0 (was explicitly set)
		assert.Equal(t, int64(0), existing.DlSpeed, "DlSpeed should be updated to 0")

		// All other fields should remain unchanged (were nil in partial)
		assert.Equal(t, "abc123", existing.Hash, "Hash should be unchanged")
		assert.Equal(t, "Original Name", existing.Name, "Name should be unchanged")
		assert.Equal(t, int64(1000), existing.Size, "Size should be unchanged")
		assert.Equal(t, 0.5, existing.Progress, "Progress should be unchanged")
		assert.Equal(t, int64(2000), existing.UpSpeed, "UpSpeed should be unchanged")
		assert.Equal(t, "downloading", existing.State, "State should be unchanged")
	})

	t.Run("zero values are applied when field is present", func(t *testing.T) {
		existing := Torrent{
			Hash:     "abc123",
			Name:     "Original Name",
			Progress: 0.75,
			DlSpeed:  5000,
			Ratio:    1.5,
		}

		// Partial with multiple fields explicitly set to zero
		zero64 := int64(0)
		zeroFloat := 0.0
		emptyStr := ""
		partial := PartialTorrent{
			DlSpeed:  &zero64,    // Torrent stopped downloading
			Progress: &zeroFloat, // This would be unusual but should work
			Category: &emptyStr,  // Category removed (empty string)
		}

		partial.ApplyTo(&existing)

		assert.Equal(t, int64(0), existing.DlSpeed, "DlSpeed should be 0")
		assert.Equal(t, 0.0, existing.Progress, "Progress should be 0")
		assert.Equal(t, "", existing.Category, "Category should be empty")

		// Unchanged fields
		assert.Equal(t, "abc123", existing.Hash)
		assert.Equal(t, "Original Name", existing.Name)
		assert.Equal(t, 1.5, existing.Ratio)
	})

	t.Run("ToTorrent creates torrent from partial", func(t *testing.T) {
		name := "Test Torrent"
		size := int64(1024)
		progress := 0.5
		state := "downloading"

		partial := PartialTorrent{
			Name:     &name,
			Size:     &size,
			Progress: &progress,
			State:    &state,
		}

		torrent := partial.ToTorrent()

		assert.Equal(t, "Test Torrent", torrent.Name)
		assert.Equal(t, int64(1024), torrent.Size)
		assert.Equal(t, 0.5, torrent.Progress)
		assert.Equal(t, "downloading", torrent.State)

		// Fields not set should be zero values
		assert.Equal(t, "", torrent.Hash)
		assert.Equal(t, int64(0), torrent.DlSpeed)
	})
}

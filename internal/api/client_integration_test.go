//go:build integration

package api

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get test server URL from env or use default
	serverURL := os.Getenv("QBT_TEST_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8181"
	}

	// Check if auth is bypassed (for subnet whitelist config)
	authBypassed := false
	resp, err := http.Get(serverURL + "/api/v2/app/version")
	if err == nil && resp.StatusCode == 200 {
		authBypassed = true
		resp.Body.Close()
		t.Log("Authentication is bypassed - using subnet whitelist")
	}

	// Use credentials from environment or defaults
	username := os.Getenv("QBT_TEST_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("QBT_TEST_PASSWORD")
	if password == "" && !authBypassed {
		t.Fatal("QBT_TEST_PASSWORD environment variable is required when auth is not bypassed.")
	}
	if password == "" {
		password = "dummy" // Won't be used but API client expects something
	}

	t.Logf("Testing with URL: %s, Auth bypassed: %v", serverURL, authBypassed)

	// First check if the server is reachable
	resp2, err2 := http.Get(serverURL)
	if err2 != nil {
		t.Skipf("qBittorrent server not reachable at %s: %v", serverURL, err2)
	}
	resp2.Body.Close()

	client, err := NewClient(serverURL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Login", func(t *testing.T) {
		err := client.Login(username, password)
		if authBypassed && err != nil {
			// Login might fail but API calls should still work
			t.Logf("Login returned error (expected with auth bypass): %v", err)
		} else if !authBypassed && err != nil {
			t.Fatalf("Login failed: %v", err)
		} else {
			t.Log("Login successful")
		}
	})

	t.Run("GetTorrents", func(t *testing.T) {
		torrents, err := client.GetTorrents(ctx)
		if assert.NoError(t, err) {
			assert.NotNil(t, torrents)
			// May be empty if no torrents
			t.Logf("Found %d torrents", len(torrents))
		}
	})

	t.Run("GetGlobalStats", func(t *testing.T) {
		stats, err := client.GetGlobalStats(ctx)
		if assert.NoError(t, err) {
			assert.NotNil(t, stats)
			assert.GreaterOrEqual(t, stats.DlInfoSpeed, int64(0))
			t.Logf("Stats: DL: %d B/s, UP: %d B/s, DHT: %d nodes",
				stats.DlInfoSpeed, stats.UpInfoSpeed, stats.DHTNodes)
		}
	})

	t.Run("GetCategories", func(t *testing.T) {
		categories, err := client.GetCategories(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		t.Logf("Found %d categories", len(categories))
	})

	t.Run("GetTags", func(t *testing.T) {
		tags, err := client.GetTags(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, tags)
		t.Logf("Found %d tags", len(tags))
	})

	t.Run("RapidRequests", func(t *testing.T) {
		// Test making multiple requests in quick succession
		for i := 0; i < 5; i++ {
			_, err := client.GetGlobalStats(ctx)
			assert.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
		}
	})
}

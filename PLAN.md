# qBittorrent TUI Project Plan

## Project Overview
Build a read-only terminal UI (TUI) client for qBittorrent focused on monitoring and filtering torrents. The application will feature a split-pane interface similar to the web UI, with real-time updates and powerful filtering capabilities.

## Core Requirements
- **Read-only monitoring** of torrent states and speeds
- **Split-pane layout**: main torrent list + sidebar with filters
- **Real-time updates** with configurable refresh rate (default: 2-3 seconds)
- **Search and filtering** by state, tracker, category, and tags
- **Drill-down view** for individual torrent details
- **Configuration file** support (TOML) for connection settings
- **Cross-platform** support (Linux/macOS primary)
- **Comprehensive test coverage**

## Technology Stack
- **Language**: Go
- **TUI Framework**: Bubble Tea + Lipgloss + Bubbles
- **HTTP Client**: Standard library (net/http)
- **Configuration**: Viper (TOML support)
- **Testing**: Standard library + testify + golden files

## Project Structure
```
qbittorrent-tui/
├── cmd/
│   └── qbt-tui/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── client.go        # API client implementation
│   │   ├── client_test.go   # API client tests
│   │   ├── types.go         # API request/response types
│   │   └── mock.go          # Mock client for testing
│   ├── ui/
│   │   ├── app.go           # Main application model
│   │   ├── app_test.go      # Application tests
│   │   ├── views/
│   │   │   ├── torrents.go      # Main torrent list view
│   │   │   ├── details.go       # Torrent details view
│   │   │   ├── sidebar.go       # Filter sidebar
│   │   │   └── views_test.go    # View tests
│   │   ├── components/
│   │   │   ├── table.go         # Torrent table component
│   │   │   ├── filter.go        # Filter component
│   │   │   └── statusbar.go     # Status bar component
│   │   └── styles/
│   │       └── theme.go         # Lipgloss styles
│   ├── config/
│   │   ├── config.go        # Configuration management
│   │   └── config_test.go   # Configuration tests
│   ├── filter/
│   │   ├── filter.go        # Filter logic
│   │   └── filter_test.go   # Filter tests
│   └── testutil/
│       ├── testutil.go      # Test utilities
│       └── docker.go        # Docker test environment
├── testdata/
│   └── golden/              # Golden files for UI tests
├── config.example.toml      # Example configuration
├── docker-compose.test.yml  # Test environment
├── Makefile                 # Build and test automation
├── go.mod
├── go.sum
├── README.md
├── PLAN.md                  # This file
└── Makefile                 # Build system and validation
```

## Implementation Phases

### Phase 1: Foundation (MUST COMPLETE FIRST)
**Goal**: Set up project structure, configuration, and API client with full test coverage.

#### 1.1 Project Setup
```bash
# Commands for Claude Code to run
go mod init github.com/yourusername/qbittorrent-tui
mkdir -p cmd/qbt-tui internal/{api,ui/views,ui/components,config,filter,testutil} testdata/golden

# Install dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/spf13/viper
go get github.com/stretchr/testify
```

#### 1.2 Configuration Management
Create `internal/config/config.go`:
```go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/spf13/viper"
)

type Config struct {
    Server struct {
        URL      string `mapstructure:"url"`
        Username string `mapstructure:"username"`
        Password string `mapstructure:"password"`
    } `mapstructure:"server"`
    
    UI struct {
        RefreshInterval int    `mapstructure:"refresh_interval"` // seconds
        Theme          string `mapstructure:"theme"`
    } `mapstructure:"ui"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("toml")
    
    // Config search paths
    viper.AddConfigPath(".")
    viper.AddConfigPath("$HOME/.config/qbt-tui")
    
    // Environment variables
    viper.SetEnvPrefix("QBT")
    viper.AutomaticEnv()
    
    // Defaults
    viper.SetDefault("ui.refresh_interval", 3)
    viper.SetDefault("ui.theme", "default")
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

Create `config.example.toml`:
```toml
[server]
url = "http://localhost:8080"
username = "admin"
password = "adminpass"

[ui]
refresh_interval = 3  # seconds
theme = "default"
```

#### 1.3 API Types
Create `internal/api/types.go`:
```go
package api

import "time"

type Torrent struct {
    Hash       string  `json:"hash"`
    Name       string  `json:"name"`
    Size       int64   `json:"size"`
    Progress   float64 `json:"progress"`
    DlSpeed    int64   `json:"dlspeed"`
    UpSpeed    int64   `json:"upspeed"`
    Priority   int     `json:"priority"`
    NumSeeds   int     `json:"num_seeds"`
    NumLeeches int     `json:"num_leechs"`
    Ratio      float64 `json:"ratio"`
    ETA        int64   `json:"eta"`
    State      string  `json:"state"`
    Category   string  `json:"category"`
    Tags       string  `json:"tags"`
    AddedOn    int64   `json:"added_on"`
    Tracker    string  `json:"tracker"`
}

type GlobalStats struct {
    DlInfoSpeed     int64 `json:"dl_info_speed"`
    UpInfoSpeed     int64 `json:"up_info_speed"`
    DlInfoData      int64 `json:"dl_info_data"`
    UpInfoData      int64 `json:"up_info_data"`
    NumTorrents     int   `json:"num_torrents"`
    NumActiveItems  int   `json:"num_active_torrents"`
}
```

#### 1.4 API Client
Create `internal/api/client.go`:
```go
package api

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/cookiejar"
    "net/url"
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
        return nil, err
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
    
    resp, err := c.httpClient.PostForm(c.baseURL+"/api/v2/auth/login", data)
    if err != nil {
        return fmt.Errorf("login request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("login failed: %s", string(body))
    }
    
    // Store cookie for future requests
    for _, cookie := range resp.Cookies() {
        if cookie.Name == "SID" {
            c.cookie = cookie.Value
            return nil
        }
    }
    
    return fmt.Errorf("no session cookie received")
}

func (c *Client) GetTorrents() ([]Torrent, error) {
    resp, err := c.httpClient.Get(c.baseURL + "/api/v2/torrents/info")
    if err != nil {
        return nil, fmt.Errorf("get torrents failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("get torrents failed with status: %d", resp.StatusCode)
    }
    
    var torrents []Torrent
    if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
        return nil, fmt.Errorf("decode torrents failed: %w", err)
    }
    
    return torrents, nil
}

func (c *Client) GetGlobalStats() (*GlobalStats, error) {
    resp, err := c.httpClient.Get(c.baseURL + "/api/v2/transfer/info")
    if err != nil {
        return nil, fmt.Errorf("get stats failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("get stats failed with status: %d", resp.StatusCode)
    }
    
    var stats GlobalStats
    if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
        return nil, fmt.Errorf("decode stats failed: %w", err)
    }
    
    return &stats, nil
}
```

#### 1.5 Test Infrastructure
Create `docker-compose.test.yml`:
```yaml
version: '3'
services:
  qbittorrent:
    image: linuxserver/qbittorrent:latest
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
      - WEBUI_PORT=8080
    ports:
      - "8080:8080"
    volumes:
      - ./testdata/qbt-config:/config
      - ./testdata/downloads:/downloads
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080"]
      interval: 5s
      timeout: 3s
      retries: 5
```

Development workflow uses `Makefile` targets:
```makefile
# Run full validation suite
validate: clean lint test-coverage
    @echo "✅ All validations passed!"

# Quick development checks  
check: fmt
    @go vet ./...
    @go test -short ./...
```

#### Phase 1 Validation
Claude Code should run `make validate` and ensure:
- [ ] All code compiles
- [ ] Configuration can be loaded from TOML
- [ ] API client can authenticate
- [ ] API client can fetch torrents
- [ ] All tests pass
- [ ] Test coverage > 80% for api package

### Phase 2: Filter System
**Goal**: Implement the filter logic as pure functions with comprehensive tests.

#### 2.1 Filter Implementation
Create `internal/filter/filter.go`:
```go
package filter

import (
    "strings"
    "github.com/yourusername/qbittorrent-tui/internal/api"
)

type Filter struct {
    States   []string // downloading, seeding, paused, etc.
    Trackers []string // tracker domains
    Category string
    Tags     []string
    Search   string // text search
}

func (f *Filter) IsEmpty() bool {
    return len(f.States) == 0 && 
           len(f.Trackers) == 0 && 
           f.Category == "" && 
           len(f.Tags) == 0 && 
           f.Search == ""
}

func (f *Filter) Apply(torrents []api.Torrent) []api.Torrent {
    if f.IsEmpty() {
        return torrents
    }
    
    var filtered []api.Torrent
    for _, t := range torrents {
        if f.matches(t) {
            filtered = append(filtered, t)
        }
    }
    return filtered
}

func (f *Filter) matches(t api.Torrent) bool {
    // State filter
    if len(f.States) > 0 && !contains(f.States, t.State) {
        return false
    }
    
    // Tracker filter (by domain)
    if len(f.Trackers) > 0 {
        domain := extractDomain(t.Tracker)
        if !contains(f.Trackers, domain) {
            return false
        }
    }
    
    // Category filter
    if f.Category != "" && t.Category != f.Category {
        return false
    }
    
    // Tags filter
    if len(f.Tags) > 0 {
        torrentTags := strings.Split(t.Tags, ",")
        if !hasAnyTag(torrentTags, f.Tags) {
            return false
        }
    }
    
    // Search filter
    if f.Search != "" {
        searchLower := strings.ToLower(f.Search)
        nameLower := strings.ToLower(t.Name)
        if !strings.Contains(nameLower, searchLower) {
            return false
        }
    }
    
    return true
}

// ExtractUniqueTrackers gets unique tracker domains from torrents
func ExtractUniqueTrackers(torrents []api.Torrent) []string {
    domains := make(map[string]bool)
    for _, t := range torrents {
        domain := extractDomain(t.Tracker)
        if domain != "" {
            domains[domain] = true
        }
    }
    
    var result []string
    for d := range domains {
        result = append(result, d)
    }
    return result
}
```

#### 2.2 Filter Tests
Create comprehensive tests to ensure filtering works correctly.

#### Phase 2 Validation
- [ ] Filter tests cover all combinations
- [ ] Filter performance is acceptable with 1000+ torrents
- [ ] Edge cases handled (empty filters, no matches)

### Phase 3: TUI Foundation
**Goal**: Build the basic TUI structure with mock data.

#### 3.1 Application Model
Create `internal/ui/app.go`:
```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yourusername/qbittorrent-tui/internal/api"
    "github.com/yourusername/qbittorrent-tui/internal/filter"
)

type View int

const (
    ViewTorrentList View = iota
    ViewTorrentDetails
)

type Model struct {
    currentView    View
    torrents       []api.Torrent
    filteredTorrents []api.Torrent
    activeFilter   filter.Filter
    globalStats    *api.GlobalStats
    
    // Sub-views
    torrentList    torrentListModel
    torrentDetails torrentDetailsModel
    sidebar        sidebarModel
    
    // UI state
    width  int
    height int
    err    error
}

func New(client *api.Client) Model {
    return Model{
        currentView: ViewTorrentList,
        torrentList: newTorrentListModel(),
        sidebar:     newSidebarModel(),
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        fetchTorrents(m.client),
        tea.Every(m.refreshInterval, func(t time.Time) tea.Msg {
            return refreshMsg(t)
        }),
    )
}
```

#### 3.2 Main View (Split Pane)
Implement the split-pane layout with proper responsive sizing.

#### 3.3 Keybindings
```go
func (m Model) handleKeyPress(key tea.KeyMsg) (Model, tea.Cmd) {
    switch key.String() {
    case "/":
        // Start search
        m.sidebar.startSearch()
    case "f":
        // Open filter menu
        m.sidebar.openFilterMenu()
    case "tab":
        // Switch focus between panes
        m.switchFocus()
    case "q", "ctrl+c":
        // Quit
        return m, tea.Quit
    }
    
    // Delegate to active pane
    switch m.focusedPane {
    case focusMain:
        m.torrentList, cmd = m.torrentList.Update(msg)
    case focusSidebar:
        m.sidebar, cmd = m.sidebar.Update(msg)
    }
    
    return m, cmd
}
```

#### Phase 3 Validation
- [ ] TUI renders with mock data
- [ ] Split pane layout works
- [ ] Keybindings respond correctly
- [ ] UI is responsive to terminal resize

### Phase 4: Integration
**Goal**: Connect the TUI to the real API with proper error handling.

#### 4.1 Connect API to TUI
- Replace mock data with real API calls
- Implement refresh logic
- Handle connection errors gracefully

#### 4.2 Status Bar
Show global stats, connection status, and active filters.

#### 4.3 Golden File Tests
Create snapshot tests for all UI states.

#### Phase 4 Validation
- [ ] Real data displays correctly
- [ ] Refresh works at configured interval
- [ ] Connection errors show user-friendly messages
- [ ] All golden tests pass

### Phase 5: Details View
**Goal**: Implement drill-down view for individual torrents.

#### 5.1 Details View Model
Show comprehensive torrent information with ability to return to list.

#### 5.2 Navigation
Implement smooth transitions between views.

#### Phase 5 Validation
- [ ] Can navigate to details and back
- [ ] Details show all relevant information
- [ ] Navigation preserves filter state

## Test Strategy

### Unit Tests (Run with -short flag)
- API client methods (mocked HTTP)
- Filter logic (pure functions)
- Configuration loading
- UI component rendering

### Integration Tests (Require Docker)
- Real qBittorrent instance
- End-to-end workflows
- Performance with many torrents

### Golden/Snapshot Tests
- Each UI state has a golden file
- Update with: `UPDATE_GOLDEN=true go test ./...`

### Example Test Pattern
```go
func TestTorrentFiltering(t *testing.T) {
    torrents := []api.Torrent{
        {Name: "Ubuntu 22.04", State: "downloading", Tracker: "https://tracker1.com/announce"},
        {Name: "Debian 12", State: "seeding", Tracker: "https://tracker2.org/announce"},
    }
    
    tests := []struct {
        name     string
        filter   filter.Filter
        expected int
    }{
        {
            name:     "filter by state",
            filter:   filter.Filter{States: []string{"downloading"}},
            expected: 1,
        },
        {
            name:     "filter by tracker domain",
            filter:   filter.Filter{Trackers: []string{"tracker1.com"}},
            expected: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.filter.Apply(torrents)
            assert.Len(t, result, tt.expected)
        })
    }
}
```

## Success Criteria

### Phase 1 Success
- [ ] Can connect to qBittorrent API
- [ ] Can fetch and display torrent data
- [ ] Configuration works from TOML file
- [ ] All tests pass

### Phase 2 Success
- [ ] Filters work correctly for all criteria
- [ ] Filter combinations work (AND logic)
- [ ] Performance acceptable with 1000+ torrents

### Phase 3 Success
- [ ] TUI renders correctly
- [ ] Split-pane layout works
- [ ] Keybindings are responsive
- [ ] Can navigate with keyboard

### Phase 4 Success
- [ ] Real-time updates work
- [ ] Error handling is graceful
- [ ] Status bar shows correct information

### Phase 5 Success
- [ ] Can drill down into torrent details
- [ ] Navigation is smooth
- [ ] State is preserved when returning

## Common Pitfalls to Avoid

1. **API Authentication**: qBittorrent uses cookie-based auth. Store and reuse the SID cookie.

2. **Tracker Extraction**: Trackers include full URLs. Extract just the domain for grouping.

3. **Terminal Compatibility**: Test with different terminal sizes. Handle minimum dimensions.

4. **Update Frequency**: Don't hammer the API. Respect the configured interval.

5. **Error States**: Always show user-friendly errors, never panic.

## Incremental Development Approach

1. **Always run tests** after each change: `make validate`
2. **Commit working code** after each successful phase
3. **Use mock data** until API integration is complete
4. **Test with real qBittorrent** instance regularly
5. **Update golden files** when UI changes are intentional

## Example Commands for Claude Code

```bash
# Start the project
cd /path/to/new/repo
go mod init github.com/yourusername/qbittorrent-tui

# After implementing each phase
make validate

# Run specific tests
go test ./internal/api -v
go test ./internal/filter -v

# Update golden files after UI changes
UPDATE_GOLDEN=true go test ./internal/ui/...

# Test with real qBittorrent
docker-compose -f docker-compose.test.yml up -d
go run cmd/qbt-tui/main.go
```

## Final Notes

- Start with Phase 1 and don't proceed until all tests pass
- Each phase builds on the previous one
- The validation script is your friend - run it often
- When in doubt, add more tests
- Keep the UI simple and focused on monitoring

This plan prioritizes a working, well-tested implementation over feature completeness. The read-only constraint significantly simplifies the project while still delivering a useful tool.

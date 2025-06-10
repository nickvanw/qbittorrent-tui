# qBittorrent TUI - Claude Development Notes

## Project Overview

A fully-featured Terminal UI for qBittorrent with responsive layout, dynamic sorting, and configurable columns. All core features are complete and working.

## Key Technical Details

### qBittorrent Authentication
- **Test password**: `testpass123`
- **Config file**: Use `testdata/qBittorrent-working.conf`
- **Critical**: `bypass_local_auth` must be `false` in qBittorrent config

### Integration Tests
```bash
QBT_TEST_PASSWORD="testpass123" go test -v -tags=integration ./internal/api
```

### Build & Test
```bash
make validate  # Full validation suite
go test ./...  # Unit tests
```

## Architecture

### Core Components
- **API client** (`internal/api/`) - qBittorrent WebUI API integration
- **Filter system** (`internal/filter/`) - Multi-criteria torrent filtering
- **TUI components** (`internal/ui/components/`) - Bubble Tea UI components
- **Configuration** (`internal/config/`) - TOML config management

### Key Features Implemented
- ✅ **Responsive layout** - Dynamic column sizing based on terminal width
- ✅ **Torrent sorting** - Sort by any column with keyboard shortcuts (1-9)
- ✅ **Configurable columns** - Show/hide 14 available columns via 'C' key
- ✅ **Advanced filtering** - State, category, tracker, tag, and text search
- ✅ **Real-time updates** - Configurable refresh interval
- ✅ **Torrent details** - Drill-down view with Enter key

## Configuration

### CLI Usage
```bash
# Command line
qbt-tui --url http://localhost:8080 --username admin --password secret

# Environment variables  
QBT_SERVER_URL=http://localhost:8080 qbt-tui

# Config file (~/.config/qbt-tui/config.toml)
[server]
url = "http://localhost:8080"
username = "admin"
password = "secret"
```

## Test Coverage
- API package: 80.5%
- Config package: 89.3%  
- Filter package: 98.9%
# qBittorrent TUI

A terminal-based user interface for monitoring and managing qBittorrent. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![qBittorrent TUI Demo](demo.gif)

## Features

- **Real-time monitoring** - Live torrent status updates
- **Advanced filtering** - Filter by state, category, tracker, tags, or search
- **Intuitive navigation** - Vim-like keyboard shortcuts
- **Torrent details** - Drill down into individual torrent information
- **Global statistics** - Download/upload speeds, connection status, disk space
- **Flexible configuration** - TOML config files, environment variables, or CLI flags

## Installation

### From Source

```bash
git clone https://github.com/yourusername/qbittorrent-tui.git
cd qbittorrent-tui
make build
sudo cp bin/qbt-tui /usr/local/bin/
```

### Using Go

```bash
go install github.com/yourusername/qbittorrent-tui/cmd/qbt-tui@latest
```

## Quick Start

1. **Start qBittorrent** with Web UI enabled (default: http://localhost:8080)

2. **Run qbt-tui**:
   ```bash
   qbt-tui --url http://localhost:8080 --username admin --password yourpassword
   ```

3. **Navigate** with keyboard shortcuts (see below)

## Configuration

Configuration priority: **CLI flags** > **Environment variables** > **Config file** > **Defaults**

### Config File

Create `~/.config/qbt-tui/config.toml`:

```toml
[server]
url = "http://localhost:8080"
username = "admin"
password = "secret"

[ui]
refresh_interval = 3  # seconds
theme = "default"
```

### Environment Variables

```bash
export QBT_SERVER_URL="http://localhost:8080"
export QBT_SERVER_USERNAME="admin"
export QBT_SERVER_PASSWORD="secret"
export QBT_UI_REFRESH_INTERVAL=3
```

### CLI Options

```bash
qbt-tui --help  # See all available options
```

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| `↑/↓`, `j/k` | Navigate torrents |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | View torrent details |
| `Esc` | Return to main view |

### Filtering
| Key | Action |
|-----|--------|
| `f`, `/` | Search torrents |
| `s` | Filter by state |
| `c` | Filter by category |
| `t` | Filter by tracker |
| `a` | Filter by tag |
| `x` | Clear all filters |

### Sorting
| Key | Action |
|-----|--------|
| `1` | Sort by name |
| `2` | Sort by size |
| `3` | Sort by progress |
| `4` | Sort by state |
| `5` | Sort by download speed |
| `6` | Sort by upload speed |
| `Shift+[1-6]` | Reverse sort direction |

### Actions
| Key | Action |
|-----|--------|
| `r` | Refresh data |
| `?` | Show/hide help |
| `Ctrl+C` | Quit |

## Filter States

| State | Description |
|-------|-------------|
| `active` | Torrents actively transferring data |
| `downloading` | Currently downloading |
| `uploading` | Currently seeding/uploading |
| `completed` | Download finished |
| `paused` | Paused torrents |
| `queued` | Queued for download/upload |
| `stalled` | Stalled (no peers/seeds) |
| `checking` | Checking files |
| `error` | Error state |

## Development

### Requirements

- Go 1.19+
- Docker (for integration tests)

### Building

```bash
# Build the application
make build

# Run tests
make test

# Generate coverage report
make test-coverage

# Run with development config
make dev
```

### Testing

```bash
# Unit tests only
make test-unit

# Integration tests (requires Docker)
make test-integration

# Full validation (lint + test + coverage)
make validate
```

## Architecture

```
├── cmd/qbt-tui/          # Main application entry point
├── internal/
│   ├── api/              # qBittorrent API client
│   ├── config/           # Configuration management
│   ├── filter/           # Torrent filtering logic
│   └── ui/               # Terminal user interface
│       ├── components/   # Reusable UI components
│       ├── styles/       # Visual styling
│       └── views/        # Application views
└── testdata/             # Test fixtures and integration tests
```

## Troubleshooting

### Connection Issues

**"Authentication required"** - Check username/password
```bash
qbt-tui --url http://localhost:8080 --username admin --password yourpassword
```

**"Connection refused"** - Ensure qBittorrent Web UI is enabled
1. Open qBittorrent → Tools → Preferences
2. Go to "Web UI" tab
3. Check "Web User Interface (Remote control)"
4. Set port (default: 8080)

### Display Issues

**"No torrents visible"** - Check filters
- Press `x` to clear all filters
- Press `r` to refresh data

**"Free space shows 0 B"** - Update qBittorrent to latest version for full API support

### Performance

For 1000+ torrents, consider:
- Increase `refresh_interval` to reduce API calls
- Use specific filters to limit displayed torrents

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make validate`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [qBittorrent](https://www.qbittorrent.org/) - The excellent BitTorrent client
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Cobra](https://github.com/spf13/cobra) - CLI framework
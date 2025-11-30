# qBittorrent TUI

A terminal-based user interface for monitoring and managing qBittorrent. Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![qBittorrent TUI Demo](https://github.com/user-attachments/assets/54e2d185-4adb-400b-b682-6bca59c36cf7)

## Features

- **Real-time monitoring** - Live torrent status updates with configurable refresh intervals
- **Advanced filtering** - Filter by state, category, tracker, tags, or text search
- **Torrent management** - Add, pause, resume, and delete torrents
- **Detailed views** - Drill down into individual torrent information with tabs for general info, trackers, peers, and files
- **Intuitive navigation** - Vim-like keyboard shortcuts and responsive layout
- **Flexible configuration** - TOML config files, environment variables, or CLI flags
- **Column customization** - Sort by any column and show/hide 14+ available columns
- **Terminal title** - Customizable terminal window/tab title with dynamic stats

## Installation

### From Source

```bash
git clone https://github.com/nickvanw/qbittorrent-tui.git
cd qbittorrent-tui
make build
sudo cp bin/qbt-tui /usr/local/bin/
```

### Using Go

```bash
go install github.com/nickvanw/qbittorrent-tui/cmd/qbt-tui@latest
```

## Quick Start

1. **Start qBittorrent** with Web UI enabled (default: http://localhost:8080)

2. **Run qbt-tui**:
   ```bash
   qbt-tui --url http://localhost:8080 --username admin --password yourpassword
   ```

3. **Navigate** with keyboard shortcuts (see below)

## Configuration

**Priority**: CLI flags > Environment variables > Config file > Defaults

### Config File (`~/.config/qbt-tui/config.toml`)

```toml
[server]
url = "http://localhost:8080"
username = "admin"
password = "secret"
refresh_interval = 3
```

### Environment Variables / CLI Options

```bash
# Environment
export QBT_SERVER_URL="http://localhost:8080"
export QBT_SERVER_USERNAME="admin"
export QBT_SERVER_PASSWORD="secret"

# CLI
qbt-tui --url http://localhost:8080 --username admin --password secret
qbt-tui --help  # See all options
```

### Terminal Title

Customize your terminal window/tab title with dynamic information (disabled by default):

```toml
[ui.terminal_title]
enabled = true
template = "qbt-tui [{active_torrents}/{total_torrents}] ↓{dl_speed} ↑{up_speed}"
```

**Available Variables:**
- `{dl_speed}`, `{up_speed}` - Download/upload speeds
- `{session_downloaded}`, `{session_uploaded}` - Session totals
- `{active_torrents}`, `{total_torrents}` - Torrent counts
- `{dl_torrents}`, `{up_torrents}`, `{paused_torrents}` - By state
- `{server_url}` - Server URL

**Example Templates:**
```toml
# Minimal
template = "qbt: ↓{dl_speed} ↑{up_speed}"

# Detailed breakdown
template = "qbt [D:{dl_torrents} U:{up_torrents} P:{paused_torrents}] ↓{dl_speed} ↑{up_speed}"

# With session totals
template = "qbt - ↓{dl_speed} ↑{up_speed} | Session: ↓{session_downloaded} ↑{session_uploaded}"
```

**Environment Variables:**
```bash
export QBT_UI_TERMINAL_TITLE_ENABLED=true
export QBT_UI_TERMINAL_TITLE_TEMPLATE="qbt [{active_torrents}/{total_torrents}] ↓{dl_speed} ↑{up_speed}"
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
| `T` | Filter by tag |
| `x` | Clear all filters |

### Sorting
| Key | Action |
|-----|--------|
| `1-9` | Sort by visible column (1st-9th) |
| `Shift+[1-9]` | Reverse sort direction |

**Note**: Sorting keys dynamically map to visible columns. Press `1` to sort by the first visible column, `2` for the second, etc. The column headers show sort indicators (↑/↓).

### Columns
| Key | Action |
|-----|--------|
| `C` | Configure columns |

### Torrent Actions
| Key | Action |
|-----|--------|
| `a` | Add torrent |
| `p` | Pause torrent |
| `u` | Resume torrent |
| `d` | Delete torrent |

### General
| Key | Action |
|-----|--------|
| `r` | Refresh data |
| `?` | Show/hide help |
| `Ctrl+C` | Quit |

## Development

Requires Go 1.19+ and Docker (for integration tests).

```bash
make build test validate  # Build, test, and lint
```

## Troubleshooting

- **Connection issues**: Enable qBittorrent Web UI (Tools → Preferences → Web UI)
- **No torrents visible**: Press `x` to clear filters or `r` to refresh  
- **Performance**: Increase `refresh_interval` or use filters for many torrents

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

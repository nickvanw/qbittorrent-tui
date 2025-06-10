# qBittorrent TUI Project Notes

## Important: qBittorrent Authentication Configuration

The qBittorrent WebUI has specific authentication requirements when accessing from outside the Docker container:

1. **Password**: The test password is `testpass123`
2. **Config File**: Use `testdata/qBittorrent-working.conf` which has proper authentication settings
3. **Key Settings**:
   - `bypass_local_auth` must be `false` (not just `LocalHostAuth=false` in config)
   - CSRF protection is disabled for testing
   - Host header validation is disabled for testing

## Integration Tests

Run integration tests with:
```bash
QBT_TEST_PASSWORD="testpass123" go test -v -tags=integration ./internal/api
```

## Project Status

✅ Phase 1: Foundation (API client, config, types) - COMPLETE
✅ Phase 2: Filter system - COMPLETE (98.9% coverage)
✅ Phase 3: TUI implementation - COMPLETE
✅ Phase 4: API integration - COMPLETE  
✅ Phase 5: Details view - COMPLETE

## CLI Usage

The application now has a comprehensive command-line interface:

```bash
# Using command line flags
qbt-tui --url http://localhost:8080 --username admin --password secret

# Using environment variables
QBT_SERVER_URL=http://localhost:8080 QBT_SERVER_USERNAME=admin qbt-tui

# Using config file (~/.config/qbt-tui/config.toml)
[server]
url = "http://localhost:8080"
username = "admin" 
password = "secret"

[ui]
refresh_interval = 5
theme = "default"
```

Available flags:
- `--url`, `-u`: qBittorrent WebUI URL (required)
- `--username`: qBittorrent username
- `--password`, `-p`: qBittorrent password  
- `--refresh`, `-r`: Refresh interval in seconds (default: 3)
- `--theme`, `-t`: UI theme (default: default)
- `--config`: Custom config file path
- `--help`, `-h`: Show comprehensive help

## Test Coverage Goals

- API package: >80% ✅ (80.5%)
- Config package: >80% ✅ (89.3%)
- Filter package: >80% ✅ (98.9%)
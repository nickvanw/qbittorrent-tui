# Test Data Directory

This directory contains files needed for integration testing.

## Files

- `qBittorrent-working.conf` - Pre-configured qBittorrent configuration with:
  - Username: `admin`
  - Password: `testpass123`
  - Security features disabled for testing (CSRF, host validation)
  
- `run-integration-test.sh` - Simple script to test qBittorrent from inside the container

## Integration Testing

Integration tests are run using:
```bash
make test-integration
```

Or directly:
```bash
QBT_TEST_PASSWORD="testpass123" go test -v -tags=integration ./internal/api
```

## Note

Do not add debugging scripts or temporary files to this directory. Use the project root or /tmp for temporary test artifacts.

#!/bin/bash
set -e

echo "Running integration test with qBittorrent container..."

# Use the configured password
TEST_PASS="testpass123"
echo "Using configured password: $TEST_PASS"

# Run tests from inside the container network
docker compose -f docker-compose.test.yml exec -T qbittorrent bash -c "
    # Test login
    echo 'Testing login...'
    COOKIE_JAR=\$(mktemp)
    RESP=\$(curl -s -c \"\$COOKIE_JAR\" -X POST 'http://localhost:8080/api/v2/auth/login' \
        -d 'username=admin&password=$TEST_PASS' \
        -w '\nSTATUS:%{http_code}')
    
    if [[ \"\$RESP\" != *\"STATUS:200\"* ]]; then
        echo 'Login failed'
        echo \"Response: \$RESP\"
        rm -f \"\$COOKIE_JAR\"
        exit 1
    fi
    
    COOKIE=\$(grep SID \"\$COOKIE_JAR\" | awk '{print \$7}')
    echo \"Got cookie: [hidden]\"
    
    # Test getting torrents
    echo 'Testing get torrents...'
    TORRENTS=\$(curl -s 'http://localhost:8080/api/v2/torrents/info' \
        -H \"Cookie: SID=\$COOKIE\" \
        -w '\nSTATUS:%{http_code}')
    
    if [[ \"\$TORRENTS\" != *\"STATUS:200\"* ]]; then
        echo 'Get torrents failed'
        rm -f \"\$COOKIE_JAR\"
        exit 1
    fi
    
    # Clean up
    rm -f \"\$COOKIE_JAR\"
    
    echo 'Tests passed!'
"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Integration tests passed!"
    echo "  Note: Tests run from inside container due to localhost auth"
    echo ""
else
    echo "❌ Integration tests failed"
    exit 1
fi
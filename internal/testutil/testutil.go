package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func CreateTempConfig(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")
	err := os.WriteFile(configFile, []byte(content), 0644)
	require.NoError(t, err)
	return tmpDir
}

func GetTestConfig() string {
	return `[server]
url = "http://localhost:8080"
username = "admin"
password = "adminpass"

[ui]
refresh_interval = 2
theme = "default"`
}

func SetEnv(t *testing.T, key, value string) {
	oldValue := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err)

	t.Cleanup(func() {
		if oldValue != "" {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	})
}

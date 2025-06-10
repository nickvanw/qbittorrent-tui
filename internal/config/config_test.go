package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		envVars     map[string]string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config file",
			configData: `[server]
url = "http://localhost:8080"
username = "admin"
password = "pass123"

[ui]
refresh_interval = 5
theme = "dark"`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "http://localhost:8080", cfg.Server.URL)
				assert.Equal(t, "admin", cfg.Server.Username)
				assert.Equal(t, "pass123", cfg.Server.Password)
				assert.Equal(t, 5, cfg.UI.RefreshInterval)
				assert.Equal(t, "dark", cfg.UI.Theme)
			},
		},
		{
			name: "env vars override config",
			configData: `[server]
url = "http://localhost:8080"

[ui]
refresh_interval = 5`,
			envVars: map[string]string{
				"QBT_SERVER_URL":      "http://192.168.1.100:9090",
				"QBT_SERVER_USERNAME": "envuser",
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "http://192.168.1.100:9090", cfg.Server.URL)
				assert.Equal(t, "envuser", cfg.Server.Username)
			},
		},
		{
			name: "defaults applied",
			configData: `[server]
url = "http://localhost:8080"`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 3, cfg.UI.RefreshInterval)
				assert.Equal(t, "default", cfg.UI.Theme)
			},
		},
		{
			name: "missing required server url",
			configData: `[server]
username = "admin"`,
			wantErr:     true,
			errContains: "server.url is required",
		},
		{
			name: "invalid refresh interval",
			configData: `[server]
url = "http://localhost:8080"

[ui]
refresh_interval = 0`,
			wantErr:     true,
			errContains: "refresh_interval must be at least 1 second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			// Set env vars before Load() which sets up viper
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			if tt.configData != "" {
				tmpDir := t.TempDir()
				configFile := filepath.Join(tmpDir, "config.toml")
				err := os.WriteFile(configFile, []byte(tt.configData), 0644)
				require.NoError(t, err)

				// Change to temp dir to ensure config is found
				oldDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				defer os.Chdir(oldDir)
			}

			cfg, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Server: struct {
					URL      string `mapstructure:"url"`
					Username string `mapstructure:"username"`
					Password string `mapstructure:"password"`
				}{
					URL: "http://localhost:8080",
				},
				UI: struct {
					RefreshInterval int    `mapstructure:"refresh_interval"`
					Theme           string `mapstructure:"theme"`
				}{
					RefreshInterval: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "missing server URL",
			config: Config{
				UI: struct {
					RefreshInterval int    `mapstructure:"refresh_interval"`
					Theme           string `mapstructure:"theme"`
				}{
					RefreshInterval: 3,
				},
			},
			wantErr: true,
			errMsg:  "server.url is required",
		},
		{
			name: "invalid refresh interval",
			config: Config{
				Server: struct {
					URL      string `mapstructure:"url"`
					Username string `mapstructure:"username"`
					Password string `mapstructure:"password"`
				}{
					URL: "http://localhost:8080",
				},
				UI: struct {
					RefreshInterval int    `mapstructure:"refresh_interval"`
					Theme           string `mapstructure:"theme"`
				}{
					RefreshInterval: 0,
				},
			},
			wantErr: true,
			errMsg:  "refresh_interval must be at least 1 second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
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
			},
		},
		{
			name: "config with columns",
			configData: `[server]
url = "http://localhost:8080"

[ui]
columns = ["name", "size", "status", "down", "up"]`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, []string{"name", "size", "status", "down", "up"}, cfg.UI.Columns)
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

			cfg, err := Load(nil)

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
					RefreshInterval int      `mapstructure:"refresh_interval"`
					Columns         []string `mapstructure:"columns"`
					DefaultSort     struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					} `mapstructure:"default_sort"`
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
					RefreshInterval int      `mapstructure:"refresh_interval"`
					Columns         []string `mapstructure:"columns"`
					DefaultSort     struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					} `mapstructure:"default_sort"`
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
					RefreshInterval int      `mapstructure:"refresh_interval"`
					Columns         []string `mapstructure:"columns"`
					DefaultSort     struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					} `mapstructure:"default_sort"`
				}{
					RefreshInterval: 0,
				},
			},
			wantErr: true,
			errMsg:  "refresh_interval must be at least 1 second",
		},
		{
			name: "invalid default sort column",
			config: Config{
				Server: struct {
					URL      string `mapstructure:"url"`
					Username string `mapstructure:"username"`
					Password string `mapstructure:"password"`
				}{
					URL: "http://localhost:8080",
				},
				UI: struct {
					RefreshInterval int      `mapstructure:"refresh_interval"`
					Columns         []string `mapstructure:"columns"`
					DefaultSort     struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					} `mapstructure:"default_sort"`
				}{
					RefreshInterval: 3,
					DefaultSort: struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					}{
						Column: "invalid_column",
					},
				},
			},
			wantErr: true,
			errMsg:  "ui.default_sort.column must be one of:",
		},
		{
			name: "invalid default sort direction",
			config: Config{
				Server: struct {
					URL      string `mapstructure:"url"`
					Username string `mapstructure:"username"`
					Password string `mapstructure:"password"`
				}{
					URL: "http://localhost:8080",
				},
				UI: struct {
					RefreshInterval int      `mapstructure:"refresh_interval"`
					Columns         []string `mapstructure:"columns"`
					DefaultSort     struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					} `mapstructure:"default_sort"`
				}{
					RefreshInterval: 3,
					DefaultSort: struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					}{
						Column:    "name",
						Direction: "invalid",
					},
				},
			},
			wantErr: true,
			errMsg:  "ui.default_sort.direction must be either 'asc' or 'desc'",
		},
		{
			name: "valid default sort",
			config: Config{
				Server: struct {
					URL      string `mapstructure:"url"`
					Username string `mapstructure:"username"`
					Password string `mapstructure:"password"`
				}{
					URL: "http://localhost:8080",
				},
				UI: struct {
					RefreshInterval int      `mapstructure:"refresh_interval"`
					Columns         []string `mapstructure:"columns"`
					DefaultSort     struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					} `mapstructure:"default_sort"`
				}{
					RefreshInterval: 3,
					DefaultSort: struct {
						Column    string `mapstructure:"column"`
						Direction string `mapstructure:"direction"`
					}{
						Column:    "size",
						Direction: "desc",
					},
				},
			},
			wantErr: false,
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

func TestConfigPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		envVars     map[string]string
		flags       map[string]string
		expected    map[string]interface{}
		description string
	}{
		{
			name: "Flag > ENV > File > Default",
			configData: `[server]
url = "http://config:8080"
username = "configuser"

[ui]
refresh_interval = 10
theme = "config_theme"`,
			envVars: map[string]string{
				"QBT_SERVER_URL":          "http://env:8080",
				"QBT_SERVER_USERNAME":     "envuser",
				"QBT_UI_REFRESH_INTERVAL": "20",
			},
			flags: map[string]string{
				"url":      "http://flag:8080",
				"username": "flaguser",
			},
			expected: map[string]interface{}{
				"server.url":          "http://flag:8080", // Flag wins
				"server.username":     "flaguser",         // Flag wins
				"ui.refresh_interval": 20,                 // ENV wins (no flag)
				// Config wins (no flag, no env)
			},
			description: "Flags override env vars and config file",
		},
		{
			name: "ENV > File > Default (no flags)",
			configData: `[server]
url = "http://config:8080"
username = "configuser"
password = "configpass"

[ui]
refresh_interval = 10`,
			envVars: map[string]string{
				"QBT_SERVER_URL":      "http://env:8080",
				"QBT_SERVER_USERNAME": "envuser",
			},
			flags: map[string]string{},
			expected: map[string]interface{}{
				"server.url":          "http://env:8080", // ENV wins
				"server.username":     "envuser",         // ENV wins
				"server.password":     "configpass",      // Config wins (no env)
				"ui.refresh_interval": 10,                // Config wins (no env)
				"ui.theme":            "default",         // Default wins (no config, no env)
			},
			description: "ENV vars override config file when no flags present",
		},
		{
			name: "File > Default (no flags, no env)",
			configData: `[server]
url = "http://config:8080"

[ui]
refresh_interval = 15`,
			envVars: map[string]string{},
			flags:   map[string]string{},
			expected: map[string]interface{}{
				"server.url":          "http://config:8080", // Config wins
				"ui.refresh_interval": 15,                   // Config wins
			},
			description: "Config file values used when no flags or env vars",
		},
		{
			name:        "Defaults only",
			configData:  "",
			envVars:     map[string]string{},
			flags:       map[string]string{},
			expected:    map[string]interface{}{},
			description: "Should fail when no server URL provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Always create a temp directory to isolate from real config files
			tmpDir := t.TempDir()
			oldDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldDir)

			// Temporarily set HOME to tmpDir to avoid loading real config
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", tmpDir)
			defer os.Setenv("HOME", oldHome)

			// Create temp config file if needed
			if tt.configData != "" {
				configFile := filepath.Join(tmpDir, "config.toml")
				err := os.WriteFile(configFile, []byte(tt.configData), 0644)
				require.NoError(t, err)
			}

			// Create a mock command with flags if needed
			var cmd *cobra.Command
			if len(tt.flags) > 0 {
				cmd = &cobra.Command{}
				cmd.Flags().String("url", "", "")
				cmd.Flags().String("username", "", "")
				cmd.Flags().String("password", "", "")
				cmd.Flags().Int("refresh", 0, "")

				// Set flag values
				for flag, value := range tt.flags {
					cmd.Flags().Set(flag, value)
				}
			}

			// Load config
			cfg, err := Load(cmd)

			// Should only fail if missing required URL
			if _, hasURL := tt.expected["server.url"]; !hasURL {
				require.Error(t, err, "expected error when server.url is missing")
				assert.Contains(t, err.Error(), "server.url is required")
				return
			}

			require.NoError(t, err, tt.description)
			require.NotNil(t, cfg)

			// Check expected values
			for key, expectedValue := range tt.expected {
				switch key {
				case "server.url":
					assert.Equal(t, expectedValue, cfg.Server.URL, tt.description)
				case "server.username":
					assert.Equal(t, expectedValue, cfg.Server.Username, tt.description)
				case "server.password":
					assert.Equal(t, expectedValue, cfg.Server.Password, tt.description)
				case "ui.refresh_interval":
					assert.Equal(t, expectedValue, cfg.UI.RefreshInterval, tt.description)
				}
			}
		})
	}
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nickvanw/qbittorrent-tui/internal/ui/components"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		URL      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"server"`

	UI struct {
		RefreshInterval int      `mapstructure:"refresh_interval"`
		Columns         []string `mapstructure:"columns"`
		DefaultSort     struct {
			Column    string `mapstructure:"column"`
			Direction string `mapstructure:"direction"`
		} `mapstructure:"default_sort"`
	} `mapstructure:"ui"`
}

func Load(cmd *cobra.Command) (*Config, error) {
	// Set config file settings
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "qbt-tui"))

	// Set defaults
	viper.SetDefault("ui.refresh_interval", 3)

	// Set up environment variable binding
	viper.SetEnvPrefix("QBT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Explicitly bind environment variables
	viper.BindEnv("server.url", "QBT_SERVER_URL")
	viper.BindEnv("server.username", "QBT_SERVER_USERNAME")
	viper.BindEnv("server.password", "QBT_SERVER_PASSWORD")
	viper.BindEnv("ui.refresh_interval", "QBT_UI_REFRESH_INTERVAL")
	viper.BindEnv("ui.columns", "QBT_UI_COLUMNS")
	viper.BindEnv("ui.default_sort.column", "QBT_UI_DEFAULT_SORT_COLUMN")
	viper.BindEnv("ui.default_sort.direction", "QBT_UI_DEFAULT_SORT_DIRECTION")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Bind command line flags if provided
	// This must happen AFTER reading config file and env vars to ensure proper precedence
	if cmd != nil {
		if err := viper.BindPFlag("server.url", cmd.Flags().Lookup("url")); err != nil {
			return nil, fmt.Errorf("failed to bind server.url flag: %w", err)
		}
		if err := viper.BindPFlag("server.username", cmd.Flags().Lookup("username")); err != nil {
			return nil, fmt.Errorf("failed to bind server.username flag: %w", err)
		}
		if err := viper.BindPFlag("server.password", cmd.Flags().Lookup("password")); err != nil {
			return nil, fmt.Errorf("failed to bind server.password flag: %w", err)
		}
		if err := viper.BindPFlag("ui.refresh_interval", cmd.Flags().Lookup("refresh")); err != nil {
			return nil, fmt.Errorf("failed to bind ui.refresh_interval flag: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.URL == "" {
		return fmt.Errorf("server.url is required")
	}

	if c.UI.RefreshInterval < 1 {
		return fmt.Errorf("ui.refresh_interval must be at least 1 second")
	}

	// Validate default sort configuration if provided
	if c.UI.DefaultSort.Column != "" {
		validColumns := components.GetValidColumnKeys()
		isValid := slices.Contains(validColumns, c.UI.DefaultSort.Column)
		if !isValid {
			return fmt.Errorf("ui.default_sort.column must be one of: %v", validColumns)
		}
	}

	// Validate sort direction if provided
	if c.UI.DefaultSort.Direction != "" {
		if c.UI.DefaultSort.Direction != "asc" && c.UI.DefaultSort.Direction != "desc" {
			return fmt.Errorf("ui.default_sort.direction must be either 'asc' or 'desc'")
		}
	}

	return nil
}

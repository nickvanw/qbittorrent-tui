package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nickvanw/qbittorrent-tui/internal/ui/components"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/terminal"
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
		TerminalTitle struct {
			Enabled  bool   `mapstructure:"enabled"`
			Template string `mapstructure:"template"`
		} `mapstructure:"terminal_title"`
	} `mapstructure:"ui"`

	Debug struct {
		Enabled bool   `mapstructure:"enabled"`  // Enable debug logging
		LogFile string `mapstructure:"log_file"` // Path to log file (empty = auto-generate)
	} `mapstructure:"debug"`
}

func Load(cmd *cobra.Command) (*Config, error) {
	// Set config file settings
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "qbt-tui"))

	// Set defaults
	viper.SetDefault("ui.refresh_interval", 3)
	viper.SetDefault("ui.terminal_title.enabled", false)
	viper.SetDefault("ui.terminal_title.template", "qbt-tui [{active_torrents}/{total_torrents}] ↓{dl_speed} ↑{up_speed}")
	viper.SetDefault("debug.enabled", false)
	viper.SetDefault("debug.log_file", "") // Auto-generate if empty

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
	viper.BindEnv("ui.terminal_title.enabled", "QBT_UI_TERMINAL_TITLE_ENABLED")
	viper.BindEnv("ui.terminal_title.template", "QBT_UI_TERMINAL_TITLE_TEMPLATE")
	viper.BindEnv("debug.enabled", "QBT_DEBUG_ENABLED")
	viper.BindEnv("debug.log_file", "QBT_DEBUG_LOG_FILE")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Bind command line flags if provided
	// This must happen AFTER reading config file and env vars to ensure proper precedence
	if cmd != nil {
		// Helper to bind flag only if it exists
		bindFlag := func(key, flagName string) error {
			if flag := cmd.Flags().Lookup(flagName); flag != nil {
				if err := viper.BindPFlag(key, flag); err != nil {
					return fmt.Errorf("failed to bind %s flag: %w", key, err)
				}
			}
			return nil
		}

		if err := bindFlag("server.url", "url"); err != nil {
			return nil, err
		}
		if err := bindFlag("server.username", "username"); err != nil {
			return nil, err
		}
		if err := bindFlag("server.password", "password"); err != nil {
			return nil, err
		}
		if err := bindFlag("ui.refresh_interval", "refresh"); err != nil {
			return nil, err
		}
		if err := bindFlag("debug.enabled", "debug"); err != nil {
			return nil, err
		}
		if err := bindFlag("debug.log_file", "log-file"); err != nil {
			return nil, err
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

	// Validate terminal title template if provided
	if c.UI.TerminalTitle.Template != "" {
		if err := terminal.ValidateTemplate(c.UI.TerminalTitle.Template); err != nil {
			return fmt.Errorf("ui.terminal_title.template is invalid: %w", err)
		}
	}

	return nil
}

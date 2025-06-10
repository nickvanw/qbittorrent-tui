package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		URL      string `mapstructure:"url"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"server"`

	UI struct {
		RefreshInterval int    `mapstructure:"refresh_interval"`
		Theme           string `mapstructure:"theme"`
	} `mapstructure:"ui"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "qbt-tui"))

	viper.SetEnvPrefix("QBT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("ui.refresh_interval", 3)
	viper.SetDefault("ui.theme", "default")

	// Bind env vars to config struct
	viper.BindEnv("server.url")
	viper.BindEnv("server.username")
	viper.BindEnv("server.password")
	viper.BindEnv("ui.refresh_interval")
	viper.BindEnv("ui.theme")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
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

	return nil
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nickvanw/qbittorrent-tui/internal/api"
	"github.com/nickvanw/qbittorrent-tui/internal/config"
	"github.com/nickvanw/qbittorrent-tui/internal/ui/views"
)

var (
	configFile string
	serverURL  string
	username   string
	password   string
	refreshInt int
	theme      string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "qbt-tui",
	Short: "A terminal user interface for qBittorrent",
	Long: `qbt-tui is a terminal-based user interface for monitoring and managing qBittorrent.

It provides a real-time view of torrents with filtering capabilities and detailed information.
The application connects to qBittorrent's WebUI API to display torrent status, statistics,
and allows for easy navigation through your torrent collection.

CONFIGURATION:
  Configuration can be provided via command line flags, environment variables, or config file.
  Priority order: CLI flags > Environment variables > Config file > Defaults

  Config file locations (TOML format):
    - ./config.toml
    - $HOME/.config/qbt-tui/config.toml

  Environment variables (prefix QBT_):
    QBT_SERVER_URL           qBittorrent WebUI URL
    QBT_SERVER_USERNAME      qBittorrent username  
    QBT_SERVER_PASSWORD      qBittorrent password
    QBT_UI_REFRESH_INTERVAL  Refresh interval in seconds (default: 3)
    QBT_UI_THEME             UI theme (default: default)

EXAMPLES:
  Using command line flags:
    qbt-tui --url http://localhost:8080 --username admin --password secret

  Using environment variables:
    QBT_SERVER_URL=http://localhost:8080 QBT_SERVER_USERNAME=admin qbt-tui

  Using config file (~/.config/qbt-tui/config.toml):
    [server]
    url = "http://localhost:8080"
    username = "admin"
    password = "secret"
    
    [ui]
    refresh_interval = 5
    theme = "default"

KEYBOARD SHORTCUTS:
  Navigation:
    ↑/↓, j/k     Navigate torrents
    g            Go to top
    G            Go to bottom
    Enter        View torrent details
    Esc          Return to main view
  
  Filtering:
    f, /         Search torrents
    s            Filter by state
    c            Filter by category
    t            Filter by tracker
    a            Filter by tag
    x            Clear filters
  
  Actions:
    r            Refresh data
    ?            Show/hide help
    Ctrl+C       Quit
`,
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Configuration file flag
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.config/qbt-tui/config.toml)")

	// Server configuration flags
	rootCmd.Flags().StringVarP(&serverURL, "url", "u", "", "qBittorrent WebUI URL (required)")
	rootCmd.Flags().StringVar(&username, "username", "", "qBittorrent username")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "qBittorrent password")

	// UI configuration flags
	rootCmd.Flags().IntVarP(&refreshInt, "refresh", "r", 3, "refresh interval in seconds (default: 3)")
	rootCmd.Flags().StringVarP(&theme, "theme", "t", "default", "UI theme (default: default)")

	// Bind flags to viper
	if err := viper.BindPFlag("server.url", rootCmd.Flags().Lookup("url")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding server.url flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("server.username", rootCmd.Flags().Lookup("username")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding server.username flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("server.password", rootCmd.Flags().Lookup("password")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding server.password flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("ui.refresh_interval", rootCmd.Flags().Lookup("refresh")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding ui.refresh_interval flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("ui.theme", rootCmd.Flags().Lookup("theme")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding ui.theme flag: %v\n", err)
		os.Exit(1)
	}

	// Mark required flags
	if err := rootCmd.MarkFlagRequired("url"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking url flag as required: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create and connect API client
	client, err := api.NewClient(cfg.Server.URL)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	if err := client.Login(cfg.Server.Username, cfg.Server.Password); err != nil {
		return fmt.Errorf("failed to connect to qBittorrent API: %w", err)
	}

	// Initialize the main view with the API client
	model := views.NewMainView(cfg, client)

	// Create the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}

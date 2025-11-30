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

    [ui.terminal_title]
    enabled = true
    template = "qbt-tui [{active_torrents}/{total_torrents}] ↓{dl_speed} ↑{up_speed}"

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
	rootCmd.Flags().StringVarP(&serverURL, "url", "u", "", "qBittorrent WebUI URL")
	rootCmd.Flags().StringVar(&username, "username", "", "qBittorrent username")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "qBittorrent password")

	// UI configuration flags
	rootCmd.Flags().IntVarP(&refreshInt, "refresh", "r", 3, "refresh interval in seconds (default: 3)")

	// Note: Flag binding will be handled in config.Load() to ensure proper precedence
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cmd)
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

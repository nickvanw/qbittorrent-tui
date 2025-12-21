package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	defaultLogger *slog.Logger
	isEnabled     bool
	logFileHandle *os.File
)

// Setup initializes the global logger with the given configuration.
// If enabled is false, logging is disabled (no-op).
// If logFile is empty, a default path is generated in ~/.local/state/qbt-tui/
func Setup(enabled bool, logFile string) error {
	isEnabled = enabled

	if !enabled {
		// Create a no-op logger
		defaultLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
		return nil
	}

	// Determine log file path
	if logFile == "" {
		// Auto-generate path: ~/.local/state/qbt-tui/debug-YYYYMMDD-HHMMSS.log
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine user home directory: %w", err)
		}
		stateDir := filepath.Join(homeDir, ".local", "state", "qbt-tui")
		if err := os.MkdirAll(stateDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		timestamp := time.Now().Format("20060102-150405")
		logFile = filepath.Join(stateDir, fmt.Sprintf("debug-%s.log", timestamp))
	}

	// Create or open log file
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}
	logFileHandle = f

	// Configure slog handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewTextHandler(logFileHandle, opts)
	defaultLogger = slog.New(handler)

	// Log startup message
	defaultLogger.Info("Debug logging initialized", "log_file", logFile)

	return nil
}

// Debug logs a debug message (always logged if debug is enabled)
func Debug(msg string, args ...any) {
	if isEnabled && defaultLogger != nil {
		defaultLogger.Debug(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	if isEnabled && defaultLogger != nil {
		defaultLogger.Info(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if isEnabled && defaultLogger != nil {
		defaultLogger.Warn(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...any) {
	if isEnabled && defaultLogger != nil {
		defaultLogger.Error(msg, args...)
	}
}

// IsEnabled returns whether debug logging is enabled
func IsEnabled() bool {
	return isEnabled
}

// Close closes the log file handle if it's open.
// This should be called when the application exits.
func Close() error {
	if logFileHandle != nil {
		return logFileHandle.Close()
	}
	return nil
}

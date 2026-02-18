package dirs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const appName = "soloterm"

// ConfigDir returns the directory for configuration files (config.yaml).
// Priority: SOLOTERM_CONFIG_DIR > SOLOTERM_WORK_DIR > os.UserConfigDir()/soloterm
func ConfigDir() (string, error) {
	if dir := os.Getenv("SOLOTERM_CONFIG_DIR"); dir != "" {
		return ensureDir(dir)
	}
	if dir := os.Getenv("SOLOTERM_WORK_DIR"); dir != "" {
		return ensureDir(dir)
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine config directory: %w", err)
	}
	return ensureDir(filepath.Join(base, appName))
}

// DataDir returns the directory for data files (database, logs).
// Priority: SOLOTERM_DATA_DIR > SOLOTERM_WORK_DIR > XDG_DATA_HOME/soloterm (Linux) or UserConfigDir/soloterm (macOS)
func DataDir() (string, error) {
	if dir := os.Getenv("SOLOTERM_DATA_DIR"); dir != "" {
		return ensureDir(dir)
	}
	if dir := os.Getenv("SOLOTERM_WORK_DIR"); dir != "" {
		return ensureDir(dir)
	}

	base, err := userDataDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine data directory: %w", err)
	}
	return ensureDir(filepath.Join(base, appName))
}

// ExportDir returns the default directory for file import/export operations.
// Returns the user's home directory with a trailing path separator,
// or falls back to "/" if the home directory cannot be determined.
func ExportDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return string(os.PathSeparator)
	}
	return home + string(os.PathSeparator)
}

// userDataDir returns the base directory for user-specific data files.
// On Linux: XDG_DATA_HOME or ~/.local/share
// On macOS: ~/Library/Application Support (same as UserConfigDir)
func userDataDir() (string, error) {
	if runtime.GOOS == "linux" {
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return dir, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}

	// macOS and others: use the same directory as config
	return os.UserConfigDir()
}

// ensureDir creates the directory if it doesn't exist and returns the path.
func ensureDir(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return dir, nil
}

// Package config resolves runtime configuration and filesystem paths for the
// tcf-prep tool: the local data directory, cache/audio/database locations, and
// the Gemini credentials used for question generation.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// AppName is the directory name created under the user's data directory.
	AppName = "tcf-prep"

	// DefaultGeminiModel is used for question generation when GEMINI_MODEL is
	// unset.
	DefaultGeminiModel = "gemini-2.5-flash"
)

// Environment variables read by Load.
const (
	envAPIKey  = "GEMINI_API_KEY"
	envModel   = "GEMINI_MODEL"
	envDataDir = "TCF_PREP_DATA_DIR"
	envXDGData = "XDG_DATA_HOME"
)

// ErrMissingGeminiAPIKey is returned by RequireGeminiAPIKey when no Gemini
// credentials are configured.
var ErrMissingGeminiAPIKey = errors.New("GEMINI_API_KEY is not set")

// Config holds resolved paths and credentials for a single tcf-prep run.
type Config struct {
	// DataDir is the root directory for all local state.
	DataDir string
	// DBPath is the SQLite database holding cached episodes and history.
	DBPath string
	// CacheDir stores scraped HTML and generated quizzes.
	CacheDir string
	// AudioDir stores downloaded episode audio.
	AudioDir string
	// GeminiAPIKey is the Gemini credential, empty when unset.
	GeminiAPIKey string
	// GeminiModel is the model used for question generation.
	GeminiModel string
}

// Load resolves configuration from the environment, falling back to
// platform-appropriate defaults. It never fails on a missing Gemini key; use
// RequireGeminiAPIKey in commands that actually need Gemini.
func Load() (*Config, error) {
	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, err
	}

	model := os.Getenv(envModel)
	if model == "" {
		model = DefaultGeminiModel
	}

	return &Config{
		DataDir:      dataDir,
		DBPath:       filepath.Join(dataDir, "tcf.db"),
		CacheDir:     filepath.Join(dataDir, "cache"),
		AudioDir:     filepath.Join(dataDir, "audio"),
		GeminiAPIKey: os.Getenv(envAPIKey),
		GeminiModel:  model,
	}, nil
}

// resolveDataDir returns the application data directory. TCF_PREP_DATA_DIR
// takes precedence, then XDG_DATA_HOME, then ~/.local/share.
func resolveDataDir() (string, error) {
	if dir := os.Getenv(envDataDir); dir != "" {
		return filepath.Clean(dir), nil
	}

	if xdg := os.Getenv(envXDGData); xdg != "" {
		return filepath.Join(xdg, AppName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", AppName), nil
}

// EnsureDirs creates the data, cache, and audio directories when missing.
func (c *Config) EnsureDirs() error {
	for _, dir := range []string{c.DataDir, c.CacheDir, c.AudioDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}

// RequireGeminiAPIKey returns ErrMissingGeminiAPIKey when no key is configured.
func (c *Config) RequireGeminiAPIKey() error {
	if c.GeminiAPIKey == "" {
		return ErrMissingGeminiAPIKey
	}
	return nil
}

package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("TCF_PREP_DATA_DIR", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantDir := filepath.Join(home, ".local", "share", AppName)
	if cfg.DataDir != wantDir {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, wantDir)
	}
	if want := filepath.Join(wantDir, "tcf.db"); cfg.DBPath != want {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, want)
	}
	if want := filepath.Join(wantDir, "cache"); cfg.CacheDir != want {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, want)
	}
	if want := filepath.Join(wantDir, "audio"); cfg.AudioDir != want {
		t.Errorf("AudioDir = %q, want %q", cfg.AudioDir, want)
	}
	if cfg.GeminiModel != DefaultGeminiModel {
		t.Errorf("GeminiModel = %q, want %q", cfg.GeminiModel, DefaultGeminiModel)
	}
	if err := cfg.RequireGeminiAPIKey(); !errors.Is(err, ErrMissingGeminiAPIKey) {
		t.Errorf("RequireGeminiAPIKey() = %v, want ErrMissingGeminiAPIKey", err)
	}
}

func TestLoadUsesXDGDataHome(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdg)
	t.Setenv("TCF_PREP_DATA_DIR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if want := filepath.Join(xdg, AppName); cfg.DataDir != want {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, want)
	}
}

func TestLoadDataDirOverrideWins(t *testing.T) {
	override := filepath.Join(t.TempDir(), "custom")
	t.Setenv("TCF_PREP_DATA_DIR", override)
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DataDir != override {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, override)
	}
}

func TestLoadModelOverride(t *testing.T) {
	t.Setenv("GEMINI_MODEL", "gemini-2.5-pro")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GeminiModel != "gemini-2.5-pro" {
		t.Errorf("GeminiModel = %q, want %q", cfg.GeminiModel, "gemini-2.5-pro")
	}
}

func TestRequireGeminiAPIKeyWhenSet(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "abc123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := cfg.RequireGeminiAPIKey(); err != nil {
		t.Errorf("RequireGeminiAPIKey() = %v, want nil", err)
	}
}

func TestEnsureDirsCreatesDirectories(t *testing.T) {
	t.Setenv("TCF_PREP_DATA_DIR", filepath.Join(t.TempDir(), "nested", AppName))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := cfg.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() error = %v", err)
	}

	for _, dir := range []string{cfg.DataDir, cfg.CacheDir, cfg.AudioDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("stat %q: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
	}
}

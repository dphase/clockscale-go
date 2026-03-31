package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func useTempDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	old := DirOverride
	DirOverride = tmp
	t.Cleanup(func() { DirOverride = old })
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Timezones) == 0 {
		t.Fatal("default config has no timezones")
	}

	// Exactly one timezone should be marked local
	localCount := 0
	for _, tz := range cfg.Timezones {
		if tz.Local {
			localCount++
		}
	}
	if localCount != 1 {
		t.Errorf("expected exactly 1 local timezone, got %d", localCount)
	}

	// Colors should be populated
	if cfg.Colors.DefaultCell.EvenBg == "" {
		t.Error("default cell even background is empty")
	}
	if cfg.Colors.DefaultCell.OddBg == "" {
		t.Error("default cell odd background is empty")
	}
	if cfg.Colors.CurrentTimeCells.Local.Bg == "" {
		t.Error("local current time cell background is empty")
	}
}

func TestDefaultConfigTimezonesAreValid(t *testing.T) {
	cfg := DefaultConfig()

	for _, tz := range cfg.Timezones {
		if tz.Timezone == "" {
			t.Error("timezone has empty name")
		}
		if tz.Label == "" {
			t.Errorf("timezone %q has empty label", tz.Timezone)
		}
	}
}

func TestDefaultConfigRoundTripsJSON(t *testing.T) {
	cfg := DefaultConfig()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var roundtripped Config
	if err := json.Unmarshal(data, &roundtripped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(roundtripped.Timezones) != len(cfg.Timezones) {
		t.Fatalf("timezone count mismatch: got %d, want %d",
			len(roundtripped.Timezones), len(cfg.Timezones))
	}

	for i, tz := range cfg.Timezones {
		rt := roundtripped.Timezones[i]
		if rt.Timezone != tz.Timezone || rt.Label != tz.Label || rt.Local != tz.Local {
			t.Errorf("timezone %d mismatch: got %+v, want %+v", i, rt, tz)
		}
	}
}

func TestBootstrapCreatesFileAndDirs(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "config.json")

	cfg := DefaultConfig()
	if err := bootstrap(path, cfg); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read bootstrapped file: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal bootstrapped file: %v", err)
	}

	if len(loaded.Timezones) != len(cfg.Timezones) {
		t.Errorf("timezone count mismatch after bootstrap")
	}
}

func TestLoadCreatesDefaultWhenMissing(t *testing.T) {
	useTempDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Timezones) == 0 {
		t.Error("loaded config has no timezones")
	}

	// Verify the file was actually created on disk
	path, _ := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created on disk")
	}
}

func TestSaveOverwritesExistingFile(t *testing.T) {
	useTempDir(t)

	// Create initial config via Load (bootstraps default)
	original, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Modify and save
	modified := *original
	modified.Timezones = append(modified.Timezones, TimezoneConfig{
		Timezone: "Europe/London",
		Label:    "GMT",
	})
	if err := Save(&modified); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Re-load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}

	if len(loaded.Timezones) != len(modified.Timezones) {
		t.Errorf("expected %d timezones after save, got %d",
			len(modified.Timezones), len(loaded.Timezones))
	}
}

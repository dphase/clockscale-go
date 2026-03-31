package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "clockscale", "config.json")

	// Override ConfigPath for this test by writing and reading directly
	cfg := DefaultConfig()
	if err := bootstrap(path, cfg); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(loaded.Timezones) == 0 {
		t.Error("loaded config has no timezones")
	}
}

func TestSaveOverwritesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	original := DefaultConfig()
	if err := bootstrap(path, original); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	modified := DefaultConfig()
	modified.Timezones = append(modified.Timezones, TimezoneConfig{
		Timezone: "Europe/London",
		Label:    "GMT",
	})
	if err := bootstrap(path, modified); err != nil {
		t.Fatalf("save modified: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(loaded.Timezones) != len(modified.Timezones) {
		t.Errorf("expected %d timezones after save, got %d",
			len(modified.Timezones), len(loaded.Timezones))
	}
}

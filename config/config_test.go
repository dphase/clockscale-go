package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestDefaultConfigRoundTripsYAML(t *testing.T) {
	cfg := DefaultConfig()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var roundtripped Config
	if err := yaml.Unmarshal(data, &roundtripped); err != nil {
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
	path := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := DefaultConfig()
	if err := bootstrap(path, cfg); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read bootstrapped file: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
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

func TestMigrateJSONToYAML(t *testing.T) {
	useTempDir(t)

	// Write a legacy config.json in the temp dir
	cfg := DefaultConfig()
	cfg.Timezones = cfg.Timezones[:2] // Use fewer timezones to distinguish from default
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}

	jsonPath := filepath.Join(DirOverride, "config.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	// Load should detect the JSON, migrate it, and return the config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Timezones) != 2 {
		t.Errorf("expected 2 timezones from migrated config, got %d", len(loaded.Timezones))
	}

	// config.yaml should now exist
	yamlPath := filepath.Join(DirOverride, "config.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Error("config.yaml was not created during migration")
	}

	// config.json should have been renamed to config.json.bak
	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Error("config.json should have been renamed to .bak")
	}
	bakPath := jsonPath + ".bak"
	if _, err := os.Stat(bakPath); os.IsNotExist(err) {
		t.Error("config.json.bak was not created")
	}
}

func TestLoadCorruptYAML(t *testing.T) {
	useTempDir(t)

	// Bootstrap a valid file first, then overwrite with garbage
	_, err := Load()
	if err != nil {
		t.Fatalf("initial Load: %v", err)
	}

	path, _ := ConfigPath()
	if err := os.WriteFile(path, []byte("{{{{not valid yaml!!!!"), 0644); err != nil {
		t.Fatalf("write corrupt yaml: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Error("expected error loading corrupt YAML, got nil")
	}
}

func TestLoadUnreadableFile(t *testing.T) {
	useTempDir(t)

	// Bootstrap a valid file, then make it unreadable
	_, err := Load()
	if err != nil {
		t.Fatalf("initial Load: %v", err)
	}

	path, _ := ConfigPath()
	if err := os.Chmod(path, 0000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(path, 0644) })

	_, err = Load()
	if err == nil {
		t.Error("expected error loading unreadable file, got nil")
	}
}

func TestMigrateCorruptJSON(t *testing.T) {
	useTempDir(t)

	// Write corrupt JSON where migration would look for it
	jsonPath := filepath.Join(DirOverride, "config.json")
	if err := os.WriteFile(jsonPath, []byte("{not json}"), 0644); err != nil {
		t.Fatalf("write corrupt json: %v", err)
	}

	// Load should fall through migration failure and bootstrap default
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Should get default config since migration failed
	if len(cfg.Timezones) == 0 {
		t.Error("expected default config timezones after failed migration")
	}
}

func TestConfigPathDefault(t *testing.T) {
	// Clear both overrides
	oldPath := PathOverride
	oldDir := DirOverride
	PathOverride = ""
	DirOverride = ""
	t.Cleanup(func() {
		PathOverride = oldPath
		DirOverride = oldDir
	})

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "clockscale", "config.yaml")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestSaveAndReload(t *testing.T) {
	useTempDir(t)

	cfg := &Config{
		Timezones: []TimezoneConfig{
			{Timezone: "Europe/London", Label: "GMT"},
		},
		Colors: DefaultConfig().Colors,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Timezones) != 1 {
		t.Fatalf("expected 1 timezone, got %d", len(loaded.Timezones))
	}
	if loaded.Timezones[0].Timezone != "Europe/London" {
		t.Errorf("expected Europe/London, got %s", loaded.Timezones[0].Timezone)
	}
}

func TestPathOverride(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "custom", "my-config.yaml")

	old := PathOverride
	PathOverride = customPath
	t.Cleanup(func() { PathOverride = old })

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath: %v", err)
	}
	if path != customPath {
		t.Errorf("expected %q, got %q", customPath, path)
	}
}

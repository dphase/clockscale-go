package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type TimezoneConfig struct {
	Timezone string `yaml:"timezone"`
	Label    string `yaml:"label"`
	Local    bool   `yaml:"local,omitempty"`
}

type ColorPair struct {
	Bg string `yaml:"bg"`
	Fg string `yaml:"fg"`
}

type DefaultCellColors struct {
	EvenBg string `yaml:"evenBg"`
	OddBg  string `yaml:"oddBg"`
	Fg     string `yaml:"fg"`
}

type CurrentTimeCellColors struct {
	Default ColorPair `yaml:"default"`
	Local   ColorPair `yaml:"local"`
}

type ColorsConfig struct {
	DefaultTimezoneLabel ColorPair             `yaml:"defaultTimezoneLabel"`
	LocalTimezoneLabel   ColorPair             `yaml:"localTimezoneLabel"`
	DefaultCell          DefaultCellColors     `yaml:"defaultCell"`
	CurrentTimeCells     CurrentTimeCellColors `yaml:"currentTimeCells"`
}

type Config struct {
	Timezones []TimezoneConfig `yaml:"timezones"`
	Colors    ColorsConfig     `yaml:"colors"`
}

func DefaultConfig() *Config {
	return &Config{
		Timezones: []TimezoneConfig{
			{Timezone: "Israel", Label: "IDT"},
			{Timezone: "Zulu", Label: "Z"},
			{Timezone: "US/Central", Label: "Local", Local: true},
			{Timezone: "US/Pacific", Label: "PDT"},
			{Timezone: "US/Eastern", Label: "EDT"},
		},
		Colors: ColorsConfig{
			DefaultTimezoneLabel: ColorPair{Bg: "default", Fg: "#02ffff"},
			LocalTimezoneLabel:   ColorPair{Bg: "default", Fg: "#ffff00"},
			DefaultCell: DefaultCellColors{
				EvenBg: "#1c1c1c",
				OddBg:  "#2d2e2e",
				Fg:     "#dadada",
			},
			CurrentTimeCells: CurrentTimeCellColors{
				Default: ColorPair{Bg: "#5e5e86", Fg: "#90ee90"},
				Local:   ColorPair{Bg: "#b4420a", Fg: "#ffff00"},
			},
		},
	}
}

// PathOverride, when non-empty, is used as the exact config file path.
// Set via the --config CLI flag.
var PathOverride string

// DirOverride, when non-empty, replaces the default config directory.
// This exists so tests can redirect all config I/O to a temp directory.
var DirOverride string

func ConfigPath() (string, error) {
	if PathOverride != "" {
		return PathOverride, nil
	}
	if DirOverride != "" {
		return filepath.Join(DirOverride, "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "clockscale", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Check for legacy config.json and migrate if found
		if migrated, mErr := migrateJSON(path); mErr == nil && migrated {
			return Load()
		}
		cfg := DefaultConfig()
		if err := bootstrap(path, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// migrateJSON looks for a config.json next to the expected config.yaml path,
// converts it to YAML, writes config.yaml, and renames the old file to config.json.bak.
// Returns true if migration occurred.
func migrateJSON(yamlPath string) (bool, error) {
	jsonPath := strings.TrimSuffix(yamlPath, ".yaml") + ".json"
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return false, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false, err
	}

	if err := bootstrap(yamlPath, &cfg); err != nil {
		return false, err
	}

	// Rename old file so it's not picked up again
	_ = os.Rename(jsonPath, jsonPath+".bak")
	return true, nil
}

func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	return bootstrap(path, cfg)
}

func bootstrap(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

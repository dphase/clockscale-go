package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TimezoneConfig struct {
	Timezone string `json:"timezone"`
	Label    string `json:"label"`
	Local    bool   `json:"local,omitempty"`
}

type ColorPair struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

type DefaultCellColors struct {
	EvenBg string `json:"evenBg"`
	OddBg  string `json:"oddBg"`
	Fg     string `json:"fg"`
}

type CurrentTimeCellColors struct {
	Default ColorPair `json:"default"`
	Local   ColorPair `json:"local"`
}

type ColorsConfig struct {
	DefaultTimezoneLabel ColorPair             `json:"defaultTimezoneLabel"`
	LocalTimezoneLabel   ColorPair             `json:"localTimezoneLabel"`
	DefaultCell          DefaultCellColors     `json:"defaultCell"`
	CurrentTimeCells     CurrentTimeCellColors `json:"currentTimeCells"`
}

type Config struct {
	Timezones []TimezoneConfig `json:"timezones"`
	Colors    ColorsConfig     `json:"colors"`
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

// DirOverride, when non-empty, replaces the default config directory.
// This exists so tests can redirect all config I/O to a temp directory.
var DirOverride string

func ConfigPath() (string, error) {
	if DirOverride != "" {
		return filepath.Join(DirOverride, "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "clockscale", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
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
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

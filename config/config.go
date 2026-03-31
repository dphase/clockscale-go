package config

import (
	"os"
	"path/filepath"

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

// DirOverride, when non-empty, replaces the default config directory.
// This exists so tests can redirect all config I/O to a temp directory.
var DirOverride string

func ConfigPath() (string, error) {
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

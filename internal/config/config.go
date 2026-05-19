package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	FontName         string `json:"font_name"`
	FontSize         int    `json:"font_size"`
	Shell            string `json:"shell"`
	CursorBlinkMode  int    `json:"cursor_blink_mode"` // 0 = system, 1 = blink on, 2 = blink off
	CursorShape      int    `json:"cursor_shape"`      // 0 = block, 1 = i-beam, 2 = underline
	ScrollbackLines  int    `json:"scrollback_lines"`
}

func DefaultConfig() *Config {
	return &Config{
		FontName:        "monospace",
		FontSize:        12,
		Shell:           "/bin/bash",
		CursorBlinkMode: 1, // On
		CursorShape:     0, // Block
		ScrollbackLines: 10000,
	}
}

func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "zet-terminal")
}

func LoadConfig() *Config {
	dir := GetConfigDir()
	path := filepath.Join(dir, "config.json")
	
	file, err := os.Open(path)
	if err != nil {
		// Doesn't exist, create default
		cfg := DefaultConfig()
		_ = SaveConfig(cfg)
		return cfg
	}
	defer file.Close()
	
	cfg := DefaultConfig()
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return DefaultConfig()
	}
	return cfg
}

func SaveConfig(cfg *Config) error {
	dir := GetConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	path := filepath.Join(dir, "config.json")
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

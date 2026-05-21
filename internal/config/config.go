package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type TabConfig struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomName bool   `json:"custom_name"`
	Pinned     bool   `json:"pinned"`
}

type GroupConfig struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Collapsed bool        `json:"collapsed"`
	Tabs      []TabConfig `json:"tabs"`
}

type KeybindingsConfig struct {
	NewTab        string `json:"new_tab"`
	NewWindow     string `json:"new_window"`
	ToggleSidebar string `json:"toggle_sidebar"`
	Copy          string `json:"copy"`
	Paste         string `json:"paste"`
	ZoomIn        string `json:"zoom_in"`
	ZoomOut       string `json:"zoom_out"`
	CloseTab      string `json:"close_tab"`
	NextTab       string `json:"next_tab"`
	PrevTab       string `json:"prev_tab"`
}

type Config struct {
	FontName           string            `json:"font_name"`
	FontSize           int               `json:"font_size"`
	Shell              string            `json:"shell"`
	CursorBlinkMode    int               `json:"cursor_blink_mode"` // 0 = system, 1 = blink on, 2 = blink off
	CursorShape        int               `json:"cursor_shape"`      // 0 = block, 1 = i-beam, 2 = underline
	ScrollbackLines    int               `json:"scrollback_lines"`
	CommandHistory     []string          `json:"command_history"`
	TabGroups          []GroupConfig     `json:"tab_groups"`
	UIThemeAccent      string            `json:"ui_theme_accent"`      // "cyan", "purple", "emerald", "amber", "crimson", "steel", "custom"
	CustomAccentColor  string            `json:"custom_accent_color"`  // Hex, default: #38bdf8
	CustomGlowColor    string            `json:"custom_glow_color"`    // rgba or hex, default: rgba(56, 189, 248, 0.4)
	UIThemeGlowOpacity float64           `json:"ui_theme_glow_opacity"` // 0.0 to 1.0, default: 0.4
	TermThemePreset    string            `json:"term_theme_preset"`    // "default", "nord", "gruvbox", "solarized", "monokai", "onehalf", "custom"
	TermBackground     string            `json:"term_background"`      // Hex, default: #121212
	TermForeground     string            `json:"term_foreground"`      // Hex, default: #e0e0e0
	TermPalette        []string          `json:"term_palette"`         // 16 ANSI colors
	Keybindings        KeybindingsConfig `json:"keybindings"`
}

func DefaultKeybindings() KeybindingsConfig {
	return KeybindingsConfig{
		NewTab:        "ctrl+t",
		NewWindow:     "ctrl+shift+n",
		ToggleSidebar: "ctrl+b",
		Copy:          "ctrl+shift+c",
		Paste:         "ctrl+shift+v",
		ZoomIn:        "ctrl+shift+plus",
		ZoomOut:       "ctrl+shift+minus",
		CloseTab:      "ctrl+shift+w",
		NextTab:       "ctrl+Page_Down",
		PrevTab:       "ctrl+Page_Up",
	}
}

func DefaultConfig() *Config {
	return &Config{
		FontName:        "monospace",
		FontSize:        11,
		Shell:           "/bin/bash",
		CursorBlinkMode: 1, // On
		CursorShape:     0, // Block
		ScrollbackLines: 10000,
		CommandHistory:  []string{},
		TabGroups: []GroupConfig{
			{
				ID:        "group-general",
				Name:      "General Workspace",
				Collapsed: false,
				Tabs: []TabConfig{
					{
						ID:   "tab-1",
						Name: "Primary Console",
					},
				},
			},
		},
		UIThemeAccent:      "cyan",
		CustomAccentColor:  "#38bdf8",
		CustomGlowColor:    "rgba(56, 189, 248, 0.4)",
		UIThemeGlowOpacity: 0.4,
		TermThemePreset:    "default",
		TermBackground:     "#121212",
		TermForeground:     "#e0e0e0",
		TermPalette: []string{
			"#1c1c1c", "#d32f2f", "#388e3c", "#fbc02d",
			"#1976d2", "#7b1fa2", "#0097a7", "#bdbdbd",
			"#424242", "#ef5350", "#66bb6a", "#fff59d",
			"#42a5f5", "#ab47bc", "#26c6da", "#ffffff",
		},
		Keybindings: DefaultKeybindings(),
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
	if cfg.FontName == "" {
		cfg.FontName = "monospace"
		_ = SaveConfig(cfg)
	}
	
	// Ensure defaults for new fields
	needsSave := false
	if data, err := os.ReadFile(path); err == nil {
		if !strings.Contains(string(data), "keybindings") {
			needsSave = true
		}
	}
	if cfg.UIThemeAccent == "" {
		cfg.UIThemeAccent = "cyan"
		needsSave = true
	}
	if cfg.CustomAccentColor == "" {
		cfg.CustomAccentColor = "#38bdf8"
		needsSave = true
	}
	if cfg.CustomGlowColor == "" {
		cfg.CustomGlowColor = "rgba(56, 189, 248, 0.4)"
		needsSave = true
	}
	if cfg.UIThemeGlowOpacity <= 0 {
		cfg.UIThemeGlowOpacity = 0.4
		needsSave = true
	}
	if cfg.TermThemePreset == "" {
		cfg.TermThemePreset = "default"
		needsSave = true
	}
	if cfg.TermBackground == "" {
		cfg.TermBackground = "#121212"
		needsSave = true
	}
	if cfg.TermForeground == "" {
		cfg.TermForeground = "#e0e0e0"
		needsSave = true
	}
	if len(cfg.TermPalette) == 0 {
		cfg.TermPalette = []string{
			"#1c1c1c", "#d32f2f", "#388e3c", "#fbc02d",
			"#1976d2", "#7b1fa2", "#0097a7", "#bdbdbd",
			"#424242", "#ef5350", "#66bb6a", "#fff59d",
			"#42a5f5", "#ab47bc", "#26c6da", "#ffffff",
		}
		needsSave = true
	}

	needsKbdSave := false
	if cfg.Keybindings.NewTab == "" {
		cfg.Keybindings.NewTab = "ctrl+t"
		needsKbdSave = true
	}
	if cfg.Keybindings.NewWindow == "" {
		cfg.Keybindings.NewWindow = "ctrl+shift+n"
		needsKbdSave = true
	}
	if cfg.Keybindings.ToggleSidebar == "" {
		cfg.Keybindings.ToggleSidebar = "ctrl+b"
		needsKbdSave = true
	}
	if cfg.Keybindings.Copy == "" {
		cfg.Keybindings.Copy = "ctrl+shift+c"
		needsKbdSave = true
	}
	if cfg.Keybindings.Paste == "" {
		cfg.Keybindings.Paste = "ctrl+shift+v"
		needsKbdSave = true
	}
	if cfg.Keybindings.ZoomIn == "" {
		cfg.Keybindings.ZoomIn = "ctrl+shift+plus"
		needsKbdSave = true
	}
	if cfg.Keybindings.ZoomOut == "" {
		cfg.Keybindings.ZoomOut = "ctrl+shift+minus"
		needsKbdSave = true
	}
	if cfg.Keybindings.CloseTab == "" {
		cfg.Keybindings.CloseTab = "ctrl+shift+w"
		needsKbdSave = true
	}
	if cfg.Keybindings.NextTab == "" {
		cfg.Keybindings.NextTab = "ctrl+Page_Down"
		needsKbdSave = true
	}
	if cfg.Keybindings.PrevTab == "" {
		cfg.Keybindings.PrevTab = "ctrl+Page_Up"
		needsKbdSave = true
	}
	if needsKbdSave {
		needsSave = true
	}

	if needsSave {
		_ = SaveConfig(cfg)
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

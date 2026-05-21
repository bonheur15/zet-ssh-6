package theme

import (
	"fmt"
	"strconv"
	"strings"

	"zet-terminal/internal/config"
)

type ColorPalette struct {
	Foreground string
	Background string
	Palette    []string // 16 ANSI colors
}

var TerminalPresets = map[string]*ColorPalette{
	"default": {
		Foreground: "#e0e0e0",
		Background: "#121212",
		Palette: []string{
			"#1c1c1c", "#d32f2f", "#388e3c", "#fbc02d",
			"#1976d2", "#7b1fa2", "#0097a7", "#bdbdbd",
			"#424242", "#ef5350", "#66bb6a", "#fff59d",
			"#42a5f5", "#ab47bc", "#26c6da", "#ffffff",
		},
	},
	"nord": {
		Foreground: "#d8dee9",
		Background: "#2e3440",
		Palette: []string{
			"#3b4252", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#88c0d0", "#e5e9f0",
			"#4c566a", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#8fbcbb", "#eceff4",
		},
	},
	"gruvbox": {
		Foreground: "#ebdbb2",
		Background: "#282828",
		Palette: []string{
			"#282828", "#cc241d", "#98971a", "#d79921",
			"#458588", "#b16286", "#689d6a", "#a89984",
			"#928374", "#fb4934", "#b8bb26", "#fabd2f",
			"#83a598", "#d3869b", "#8ec07c", "#ebdbb2",
		},
	},
	"solarized": {
		Foreground: "#839496",
		Background: "#002b36",
		Palette: []string{
			"#073642", "#dc322f", "#859900", "#b58900",
			"#268bd2", "#d33682", "#2aa198", "#eee8d5",
			"#002b36", "#cb4b16", "#586e75", "#657b83",
			"#839496", "#6c71c4", "#93a1a1", "#fdf6e3",
		},
	},
	"monokai": {
		Foreground: "#f8f8f2",
		Background: "#272822",
		Palette: []string{
			"#272822", "#f92672", "#a6e22e", "#f4bf75",
			"#66d9ef", "#ae81ff", "#a1efe4", "#f8f8f2",
			"#75715e", "#f92672", "#a6e22e", "#f4bf75",
			"#66d9ef", "#ae81ff", "#a1efe4", "#f9f8f5",
		},
	},
	"onehalf": {
		Foreground: "#abb2bf",
		Background: "#282c34",
		Palette: []string{
			"#282c34", "#e06c75", "#98c379", "#d19a66",
			"#61afef", "#c678dd", "#56b6c2", "#abb2bf",
			"#5c6370", "#e06c75", "#98c379", "#d19a66",
			"#61afef", "#c678dd", "#56b6c2", "#ffffff",
		},
	},
}

var AccentPresets = map[string]string{
	"cyan":    "#38bdf8",
	"purple":  "#c084fc",
	"emerald": "#34d399",
	"amber":   "#fbbf24",
	"crimson": "#f43f5e",
	"steel":   "#94a3b8",
}

func GetTerminalPalette(cfg *config.Config) *ColorPalette {
	if cfg.TermThemePreset == "custom" {
		return &ColorPalette{
			Foreground: cfg.TermForeground,
			Background: cfg.TermBackground,
			Palette:    cfg.TermPalette,
		}
	}
	preset, ok := TerminalPresets[cfg.TermThemePreset]
	if ok {
		return preset
	}
	return TerminalPresets["default"]
}

func hexToRGBA(hex string, alpha float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return fmt.Sprintf("rgba(56, 189, 248, %g)", alpha)
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return fmt.Sprintf("rgba(%d, %d, %d, %g)", r, g, b, alpha)
}

func GetCSS(cfg *config.Config) string {
	accent := AccentPresets[cfg.UIThemeAccent]
	if cfg.UIThemeAccent == "custom" {
		accent = cfg.CustomAccentColor
	}
	if accent == "" {
		accent = "#38bdf8"
	}

	opacity := cfg.UIThemeGlowOpacity
	if opacity <= 0 {
		opacity = 0.4
	}

	glow := hexToRGBA(accent, opacity)
	if cfg.UIThemeAccent == "custom" && cfg.CustomGlowColor != "" && !strings.Contains(cfg.CustomGlowColor, "rgba") {
		// If custom glow is hex, parse it with current opacity
		glow = hexToRGBA(cfg.CustomGlowColor, opacity)
	} else if cfg.UIThemeAccent == "custom" && cfg.CustomGlowColor != "" {
		glow = cfg.CustomGlowColor
	}

	bgGlow := hexToRGBA(accent, opacity*0.2)
	bgGlowActive := hexToRGBA(accent, opacity*0.375)
	borderDim := hexToRGBA(accent, opacity*0.375)
	borderHover := hexToRGBA(accent, opacity*0.875)
	shadowGlow := hexToRGBA(accent, opacity*0.375)

	css := strings.ReplaceAll(TerminalCSS, "__ACCENT_COLOR__", accent)
	css = strings.ReplaceAll(css, "__ACCENT_GLOW__", glow)
	css = strings.ReplaceAll(css, "__ACCENT_BG_GLOW__", bgGlow)
	css = strings.ReplaceAll(css, "__ACCENT_BG_ACTIVE__", bgGlowActive)
	css = strings.ReplaceAll(css, "__ACCENT_BORDER_DIM__", borderDim)
	css = strings.ReplaceAll(css, "__ACCENT_BORDER_HOVER__", borderHover)
	css = strings.ReplaceAll(css, "__ACCENT_SHADOW_GLOW__", shadowGlow)

	return css
}

const TerminalCSS = `
/* ─── Modern Sleek Dark Terminal Theme ─── */
window.terminal-window {
	background: #0e1012;
	color: #f1f5f9;
}

headerbar {
	background: #111317;
	border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	border-top: none;
	border-left: none;
	border-right: none;
	min-height: 28px;
	padding: 0px 6px;
	margin: 0;
	outline: none;
	box-shadow: none;
}

headerbar windowhandle,
headerbar windowhandle > box,
headerbar box {
	min-height: 28px;
	padding: 0;
	margin: 0;
}

headerbar windowcontrols {
	min-height: 28px;
	min-width: 0px;
	padding: 0;
	margin: 0;
}

headerbar button,
headerbar button > contents,
headerbar windowcontrols button,
headerbar windowcontrols button > contents {
	padding: 0px;
	margin: 0px 2px;
	min-height: 22px;
	min-width: 22px;
}

headerbar .header-btn,
headerbar .header-btn > contents {
	background: transparent;
	border: none;
	box-shadow: none;
}

headerbar .header-btn image {
	-gtk-icon-size: 14px;
	min-width: 14px;
	min-height: 14px;
	padding: 0;
	margin: 0;
}

headerbar .header-btn {
	border-radius: 4px;
	transition: background-color 0.2s ease;
}

headerbar .header-btn:hover {
	background: rgba(255, 255, 255, 0.08);
}

headerbar .title-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 13px;
	font-weight: 700;
	color: #cbd5e1;
	padding: 0px;
	margin: 0;
}

/* ─── Settings Dialog Styles ─── */
window.settings-dialog {
	background: #111317;
	color: #f1f5f9;
}

.settings-box {
	padding: 16px 20px;
}

.settings-section-title {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
	font-size: 10px;
	font-weight: 800;
	color: __ACCENT_COLOR__;
	margin-top: 8px;
	margin-bottom: 12px;
	letter-spacing: 1px;
	text-transform: uppercase;
}

.settings-row {
	margin-bottom: 10px;
}

.settings-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, sans-serif;
	font-size: 11px;
	font-weight: 600;
	color: #cbd5e1;
}

.settings-entry {
	background: rgba(255, 255, 255, 0.04);
	border: 1px solid rgba(255, 255, 255, 0.1);
	border-radius: 6px;
	color: #f1f5f9;
	padding: 5px 8px;
	font-size: 11px;
}

.settings-entry:focus {
	border-color: __ACCENT_COLOR__;
	outline: none;
}

.settings-dropdown {
	background: rgba(255, 255, 255, 0.04);
	border: 1px solid rgba(255, 255, 255, 0.1);
	border-radius: 6px;
	color: #f1f5f9;
	padding: 4px;
	font-size: 11px;
}

/* ─── VTE Terminal Widget Container ─── */
.terminal-container {
	background: #0e1012;
	border: none;
	padding: 4px;
	margin: 0;
}

/* ─── Side Dock & Panel ─── */
.sidebar-dock {
	background: transparent;
}

.sidebar-panel {
	background-image: linear-gradient(to bottom, #16181c, #0e1012);
	border-right: 1px solid rgba(255, 255, 255, 0.05);
	padding: 14px 10px;
	box-shadow: 4px 0 16px rgba(0, 0, 0, 0.5);
}

.sidebar-title {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 10px;
	font-weight: 800;
	color: #64748b;
	margin-bottom: 12px;
	letter-spacing: 1.2px;
	text-transform: uppercase;
}

.sidebar-btn {
	background: transparent;
	border: none;
	border-radius: 6px;
	padding: 6px 10px;
	transition: background-color 0.2s ease;
}

.sidebar-btn:hover {
	background: rgba(255, 255, 255, 0.05);
}

.sidebar-btn-label {
	font-family: "JetBrains Mono", "Fira Code", "Hack", monospace;
	font-size: 10px;
	color: #94a3b8;
	transition: color 0.2s ease;
}

.sidebar-row:hover .sidebar-btn-label {
	color: #f1f5f9;
}

.sidebar-trigger {
	background-color: rgba(255, 255, 255, 0.001);
	border-right: 1px solid rgba(255, 255, 255, 0.02);
	transition: background-color 0.2s ease;
}

.sidebar-trigger:hover {
	background-color: rgba(255, 255, 255, 0.04);
}

.sidebar-row {
	margin: 3px 4px;
	border-radius: 6px;
	background: rgba(255, 255, 255, 0.02);
	border: 1px solid rgba(255, 255, 255, 0.04);
	transition: all 0.2s ease;
}

.sidebar-row:hover {
	background: rgba(255, 255, 255, 0.05);
	border-color: rgba(255, 255, 255, 0.08);
}

.sidebar-arrow-btn {
	background: transparent;
	border: none;
	padding: 6px 8px;
	opacity: 0;
	transition: opacity 0.2s ease, background-color 0.2s ease;
}

.sidebar-row:hover .sidebar-arrow-btn {
	opacity: 0.6;
}

.sidebar-arrow-btn:hover {
	opacity: 1.0;
	background: __ACCENT_BG_GLOW__;
	border-radius: 4px;
}

.sidebar-arrow-label {
	font-size: 11px;
	color: __ACCENT_COLOR__;
	font-weight: 700;
}

/* Sleek Context Menu Popover */
popover > contents {
	background: #16181c;
	border: 1px solid rgba(255, 255, 255, 0.08);
	border-radius: 8px;
	padding: 4px;
	box-shadow: 0 4px 20px rgba(0, 0, 0, 0.6);
}

.menu-item-btn {
	background: transparent;
	border: none;
	border-radius: 6px;
	padding: 6px 12px;
	min-width: 100px;
	transition: background-color 0.2s ease;
}

.menu-item-btn:hover {
	background: rgba(255, 255, 255, 0.06);
}

.menu-item-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 11px;
	font-weight: 600;
	color: #cbd5e1;
}

/* Premium Workspace Panel Styling */
.workspace-header {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 10px;
	font-weight: 800;
	color: #64748b;
	margin-top: 14px;
	margin-bottom: 10px;
	letter-spacing: 1.2px;
	text-transform: uppercase;
}

.workspace-action-btn {
	background: __ACCENT_BG_GLOW__;
	border: 1px solid __ACCENT_BORDER_DIM__;
	border-radius: 6px;
	padding: 6px 12px;
	margin-bottom: 14px;
	transition: all 0.2s ease-in-out;
}

.workspace-action-btn:hover {
	background: __ACCENT_BG_ACTIVE__;
	border-color: __ACCENT_BORDER_HOVER__;
	box-shadow: 0 2px 8px __ACCENT_SHADOW_GLOW__;
}

.workspace-action-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 11px;
	color: __ACCENT_COLOR__;
	font-weight: 700;
	letter-spacing: 0.5px;
}

/* Tab Group Header Styling */
.group-header-row {
	background: transparent;
	padding: 6px 4px;
	margin-top: 8px;
	border-bottom: 1px solid rgba(255, 255, 255, 0.03);
	border-radius: 4px;
	transition: background-color 0.2s ease;
}

.group-header-row:hover {
	background: rgba(255, 255, 255, 0.02);
}

.group-toggle-btn {
	background: transparent;
	border: none;
	padding: 2px;
}

.group-toggle-label {
	font-size: 10px;
	color: #64748b;
	transition: color 0.2s ease;
}

.group-header-row:hover .group-toggle-label {
	color: #94a3b8;
}

.group-title-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 11px;
	font-weight: 600;
	color: #cbd5e1;
	margin-left: 6px;
}

.group-header-row:hover .group-title-label {
	color: #f1f5f9;
}

.group-action-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 4px 6px;
	margin-left: 2px;
	opacity: 0;
	transition: opacity 0.2s ease, background-color 0.2s ease;
}

.group-header-row:hover .group-action-btn {
	opacity: 0.5;
}

.group-action-btn:hover {
	opacity: 1.0;
	background: rgba(255, 255, 255, 0.08);
}

.group-action-btn image {
	color: #94a3b8;
}

.group-action-btn:hover image {
	color: #ffffff;
}

/* Tab Row Styling */
.tab-row {
	margin-left: 8px;
	margin-top: 3px;
	margin-bottom: 3px;
	border-radius: 4px;
	background: transparent;
	transition: background-color 0.2s ease;
}

.tab-row:hover:not(.active) {
	background: rgba(255, 255, 255, 0.03);
}

.tab-row.active {
	background: linear-gradient(to right, __ACCENT_BG_ACTIVE__, rgba(255, 255, 255, 0.02));
	border-left: 3px solid __ACCENT_COLOR__;
	border-radius: 0 4px 4px 0;
}

.tab-select-btn {
	background: transparent;
	border: none;
	padding: 6px 8px;
}

.tab-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 11px;
	color: #94a3b8;
	transition: color 0.2s ease;
}

.tab-row:hover .tab-label {
	color: #cbd5e1;
}

.tab-row.active .tab-label {
	color: __ACCENT_COLOR__;
	font-weight: 700;
}

.tab-action-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 4px;
	margin-right: 4px;
	opacity: 0;
	transition: opacity 0.2s ease, background-color 0.2s ease;
}

.tab-row:hover .tab-action-btn {
	opacity: 0.5;
}

.tab-action-btn:hover:not(.tab-close-btn) {
	opacity: 1.0;
	background: rgba(255, 255, 255, 0.08);
}

.tab-action-btn:hover:not(.tab-close-btn) image {
	color: #ffffff;
}

.tab-action-btn.tab-close-btn:hover {
	opacity: 1.0;
	background: rgba(239, 68, 68, 0.15);
}

.tab-action-btn.tab-close-btn:hover image {
	color: #ef4444;
}

.tab-action-btn image {
	color: #64748b;
	transition: color 0.2s ease;
}

.tab-row:hover .tab-action-btn image {
	color: #94a3b8;
}

/* Collapsible Section for Past Commands */
.history-section-header {
	background: transparent;
	border: none;
	padding: 10px 4px;
	margin-top: 20px;
	border-top: 1px solid rgba(255, 255, 255, 0.05);
	transition: background-color 0.2s ease;
}

.history-section-header:hover {
	background: rgba(255, 255, 255, 0.02);
}

.history-header-arrow {
	font-size: 9px;
	color: #64748b;
	margin-right: 8px;
}

.history-header-title {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 10px;
	font-weight: 800;
	color: #64748b;
	letter-spacing: 1.2px;
	text-transform: uppercase;
}

.history-section-header:hover .history-header-title {
	color: #94a3b8;
}

.history-section-header:hover .history-header-arrow {
	color: #94a3b8;
}

/* Custom Scrollbar Styles for modern GTK4 */
scrollbar {
	background-color: transparent;
	background-image: none;
	border: none;
}
scrollbar.vertical {
	min-width: 6px;
}
scrollbar.horizontal {
	min-height: 6px;
}
scrollbar button {
	min-width: 0px;
	min-height: 0px;
	opacity: 0;
	padding: 0;
	margin: 0;
}
scrollbar slider {
	background-color: rgba(255, 255, 255, 0.08);
	border-radius: 4px;
	border: none;
	min-width: 6px;
	min-height: 6px;
	margin: 0;
	transition: background-color 0.2s ease;
}
scrollbar slider:hover {
	background-color: __ACCENT_GLOW__;
}

/* Styled sidebar popover text input fields */
.sidebar-entry {
	background: rgba(255, 255, 255, 0.04);
	border: 1px solid rgba(255, 255, 255, 0.1);
	border-radius: 6px;
	color: #f1f5f9;
	padding: 6px 10px;
	margin-bottom: 8px;
	transition: all 0.2s ease;
}

.sidebar-inline-edit-row {
	background: rgba(255, 255, 255, 0.03);
	border: 1px solid rgba(255, 255, 255, 0.08);
	border-radius: 6px;
	padding: 4px;
	margin-top: 6px;
	margin-bottom: 6px;
}

.sidebar-inline-entry {
	background: rgba(0, 0, 0, 0.2);
	border: 1px solid rgba(255, 255, 255, 0.1);
	border-radius: 4px;
	color: #f1f5f9;
	padding: 3px 6px;
	font-size: 11px;
}

.sidebar-inline-entry:focus {
	border-color: __ACCENT_COLOR__;
}

.sidebar-inline-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 4px 6px;
	transition: background-color 0.2s ease;
}

.sidebar-inline-btn:hover {
	background: rgba(255, 255, 255, 0.08);
}

.sidebar-inline-btn image {
	color: #cbd5e1;
}

.sidebar-inline-btn:hover image {
	color: #ffffff;
}
`

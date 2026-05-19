package theme

type ColorPalette struct {
	Foreground string
	Background string
	Palette    []string // 16 ANSI colors
}

var DefaultPalette = &ColorPalette{
	Foreground: "#e0e0e0",
	Background: "#121212",
	Palette: []string{
		"#1c1c1c", "#d32f2f", "#388e3c", "#fbc02d", // 0-3
		"#1976d2", "#7b1fa2", "#0097a7", "#bdbdbd", // 4-7
		"#424242", "#ef5350", "#66bb6a", "#fff59d", // 8-11
		"#42a5f5", "#ab47bc", "#26c6da", "#ffffff", // 12-15
	},
}

const TerminalCSS = `
/* ─── Minimalistic Solid Dark Terminal Style ─── */
window.terminal-window {
	background: #121212;
	color: #e0e0e0;
}

headerbar {
	background: #181818;
	border-bottom: 1px solid #262626;
	min-height: 38px;
	padding: 0 12px;
}

headerbar label {
	font-family: inherit;
	font-size: 13px;
	font-weight: 600;
	color: #e0e0e0;
}

/* ─── Tab Bar & Tabs ─── */
notebook {
	background: #121212;
	border: none;
}

notebook header {
	background: #1a1a1a;
	border-bottom: 1px solid #262626;
	padding: 2px 4px 0 4px;
}

notebook tab {
	background: #242424;
	border: 1px solid #2c2c2c;
	border-bottom: none;
	border-radius: 4px 4px 0 0;
	padding: 6px 12px;
	margin-right: 2px;
	color: #888888;
	font-size: 12px;
	transition: all 100ms ease;
}

notebook tab:hover {
	background: #2c2c2c;
	color: #cccccc;
}

notebook tab:checked {
	background: #121212;
	border-top: 2px solid #777777;
	border-left: 1px solid #2c2c2c;
	border-right: 1px solid #2c2c2c;
	color: #ffffff;
	font-weight: 600;
}

/* Tab Close Button */
notebook tab button {
	background: transparent;
	border: none;
	border-radius: 4px;
	color: #666666;
	margin-left: 8px;
	padding: 2px;
}

notebook tab button:hover {
	background: #333333;
	color: #ffffff;
}

/* ─── VTE Terminal Widget Container ─── */
.terminal-container {
	background: #121212;
	border: none;
	padding: 4px;
	margin: 0;
}

/* Header Buttons styling */
.header-btn {
	background: #242424;
	border: 1px solid #2c2c2c;
	border-radius: 4px;
	color: #cccccc;
	padding: 4px 10px;
	margin: 0 2px;
	font-size: 12px;
}

.header-btn:hover {
	background: #2c2c2c;
	color: #ffffff;
	border-color: #444444;
}
`

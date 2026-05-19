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
	min-height: 34px;
	padding: 0 12px;
}

headerbar label {
	font-family: inherit;
	font-size: 13px;
	font-weight: 600;
	color: #e0e0e0;
}

/* ─── VTE Terminal Widget Container ─── */
.terminal-container {
	background: #121212;
	border: none;
	padding: 4px;
	margin: 0;
}

.vte-terminal-widget {
	font-family: "Hack", "Liberation Mono", "DejaVu Sans Mono", monospace;
}
`

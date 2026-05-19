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
	border-top: none;
	border-left: none;
	border-right: none;
	min-height: 24px;
	padding: 0;
	margin: 0;
	align-items: center;
	outline: none;
	box-shadow: none;
}

headerbar windowhandle {
	min-height: 24px;
	padding: 0 8px;
	margin: 0;
	align-items: center;
	border: none;
	outline: none;
	box-shadow: none;
}

headerbar button {
	padding: 0;
	margin: 0 2px;
	min-height: 18px;
	min-width: 18px;
	align-self: center;
}

headerbar label {
	font-family: inherit;
	font-size: 11px;
	font-weight: 600;
	color: #e0e0e0;
	padding: 0;
	margin: 0;
	align-self: center;
}

/* ─── VTE Terminal Widget Container ─── */
.terminal-container {
	background: #121212;
	border: none;
	padding: 4px;
	margin: 0;
}
`

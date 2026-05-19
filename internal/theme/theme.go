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

/* ─── Side Dock & Panel ─── */
.sidebar-dock {
	background: transparent;
}

.sidebar-panel {
	background: #151515;
	border-right: 1px solid #242424;
	padding: 12px 8px;
	box-shadow: 4px 0 12px rgba(0, 0, 0, 0.4);
}

.sidebar-title {
	font-family: inherit;
	font-size: 11px;
	font-weight: 700;
	color: #888888;
	margin-bottom: 12px;
	letter-spacing: 1px;
	text-transform: uppercase;
}

.sidebar-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 6px 8px;
	transition: background 0.2s ease;
}

.sidebar-btn:hover {
	background: #202020;
}

.sidebar-btn-label {
	font-family: "Hack", monospace;
	font-size: 10px;
	color: #d0d0d0;
}

.sidebar-btn:hover .sidebar-btn-label {
	color: #ffffff;
}

.sidebar-trigger {
	background: transparent;
	border-right: 1px solid #202020;
	transition: background 0.2s ease;
}

.sidebar-trigger:hover {
	background: rgba(255, 255, 255, 0.02);
}

.sidebar-row {
	margin-bottom: 4px;
	border: none;
	background: transparent;
}

.sidebar-arrow-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 6px;
	margin-left: 2px;
	transition: background 0.2s ease;
}

.sidebar-arrow-btn:hover {
	background: #252525;
}

.sidebar-arrow-label {
	font-size: 11px;
	color: #888888;
}

.sidebar-arrow-btn:hover .sidebar-arrow-label {
	color: #38bdf8;
}
`

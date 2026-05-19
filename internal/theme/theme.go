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

/* Sleek Context Menu Popover */
popover > contents {
	background: #181818 !important;
	border: 1px solid #282828 !important;
	border-radius: 6px !important;
	padding: 4px !important;
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5) !important;
}

.menu-item-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 6px 12px;
	min-width: 100px;
	transition: background 0.2s ease;
}

.menu-item-btn:hover {
	background: #242424;
}

.menu-item-label {
	font-family: inherit;
	font-size: 11px;
	font-weight: 500;
	color: #e0e0e0;
}

/* Premium Workspace Panel Styling */
.workspace-header {
	font-family: inherit;
	font-size: 11px;
	font-weight: 700;
	color: #888888;
	margin-top: 12px;
	margin-bottom: 8px;
	letter-spacing: 1px;
	text-transform: uppercase;
}

.workspace-action-btn {
	background: #181818;
	border: 1px solid #282828;
	border-radius: 4px;
	padding: 4px 8px;
	margin-bottom: 12px;
	transition: all 0.2s ease;
}

.workspace-action-btn:hover {
	background: #242424;
	border-color: #444444;
}

.workspace-action-label {
	font-size: 10px;
	color: #d0d0d0;
	font-weight: 600;
}

/* Tab Group Header Styling */
.group-header-row {
	background: transparent;
	padding: 4px 2px;
	margin-top: 6px;
	border-bottom: 1px solid rgba(255, 255, 255, 0.03);
}

.group-toggle-btn {
	background: transparent;
	border: none;
	padding: 2px 4px;
}

.group-toggle-label {
	font-size: 10px;
	color: #888888;
}

.group-title-label {
	font-family: inherit;
	font-size: 11px;
	font-weight: 600;
	color: #e0e0e0;
	margin-left: 4px;
}

.group-action-btn {
	background: transparent;
	border: none;
	border-radius: 4px;
	padding: 2px 6px;
	margin-left: 2px;
	transition: background 0.2s ease;
}

.group-action-btn:hover {
	background: #202020;
}

.group-action-label {
	font-size: 10px;
	color: #888888;
}

.group-action-btn:hover .group-action-label {
	color: #ffffff;
}

/* Tab Row Styling */
.tab-row {
	margin-left: 12px;
	margin-top: 2px;
	margin-bottom: 2px;
	border-radius: 4px;
	transition: background 0.2s ease;
}

.tab-row.active {
	background: #1e293b; /* Premium slate-blue active state color */
}

.tab-select-btn {
	background: transparent;
	border: none;
	padding: 4px 6px;
	text-align: left;
}

.tab-label {
	font-family: inherit;
	font-size: 10px;
	color: #a0a0a0;
}

.tab-row.active .tab-label {
	color: #38bdf8; /* modern neon blue glow for active tab text */
	font-weight: 600;
}

.tab-action-btn {
	background: transparent;
	border: none;
	border-radius: 3px;
	padding: 2px 4px;
	transition: background 0.2s ease;
}

.tab-action-btn:hover {
	background: rgba(255, 255, 255, 0.05);
}

.tab-action-label {
	font-size: 9px;
	color: #666666;
}

.tab-action-btn:hover .tab-action-label {
	color: #ef4444; /* Close button glows red on hover */
}

/* Collapsible Section for Past Commands */
.history-section-header {
	background: transparent;
	border: none;
	padding: 8px 4px;
	margin-top: 16px;
	border-top: 1px solid #202020;
	transition: background 0.2s ease;
}

.history-section-header:hover {
	background: rgba(255, 255, 255, 0.01);
}

.history-header-arrow {
	font-size: 10px;
	color: #666666;
	margin-right: 6px;
}

.history-header-title {
	font-family: inherit;
	font-size: 10px;
	font-weight: 700;
	color: #888888;
	letter-spacing: 0.5px;
	text-transform: uppercase;
}
`

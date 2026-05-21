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
	color: #38bdf8;
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
	border-color: #38bdf8;
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
	background: rgba(56, 189, 248, 0.1);
	border-radius: 4px;
}

.sidebar-arrow-label {
	font-size: 11px;
	color: #38bdf8;
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
	background: rgba(56, 189, 248, 0.08);
	border: 1px solid rgba(56, 189, 248, 0.15);
	border-radius: 6px;
	padding: 6px 12px;
	margin-bottom: 14px;
	transition: all 0.2s ease-in-out;
}

.workspace-action-btn:hover {
	background: rgba(56, 189, 248, 0.16);
	border-color: rgba(56, 189, 248, 0.35);
	box-shadow: 0 2px 8px rgba(56, 189, 248, 0.15);
}

.workspace-action-label {
	font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
	font-size: 11px;
	color: #38bdf8;
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
	background: linear-gradient(to right, rgba(56, 189, 248, 0.15), rgba(56, 189, 248, 0.02));
	border-left: 3px solid #38bdf8;
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
	color: #38bdf8;
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

.tab-action-btn:hover {
	opacity: 1.0;
	background: rgba(239, 68, 68, 0.15);
}

.tab-action-btn image {
	color: #64748b;
	transition: color 0.2s ease;
}

.tab-row:hover .tab-action-btn image {
	color: #94a3b8;
}

.tab-action-btn:hover image {
	color: #ef4444;
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
	background-color: rgba(56, 189, 248, 0.4);
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
	border-color: #38bdf8;
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


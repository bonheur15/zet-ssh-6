package theme

type ColorPalette struct {
	Name       string
	Foreground string
	Background string
	Palette    []string // 16 ANSI colors
}

var Themes = map[string]*ColorPalette{
	"tokyonight": {
		Name:       "Tokyo Night",
		Foreground: "#a9b1d6",
		Background: "#1a1b26",
		Palette: []string{
			"#15161e", "#f7768e", "#9ece6a", "#e0af68", // 0-3
			"#7aa2f7", "#bb9af7", "#7dcfff", "#a9b1d6", // 4-7
			"#414868", "#f7768e", "#9ece6a", "#e0af68", // 8-11
			"#7aa2f7", "#bb9af7", "#7dcfff", "#c0caf5", // 12-15
		},
	},
	"dracula": {
		Name:       "Dracula",
		Foreground: "#f8f8f2",
		Background: "#282a36",
		Palette: []string{
			"#000000", "#ff5555", "#50fa7b", "#f1fa8c",
			"#bd93f9", "#ff79c6", "#8be9fd", "#bbbbbb",
			"#555555", "#ff5555", "#50fa7b", "#f1fa8c",
			"#bd93f9", "#ff79c6", "#8be9fd", "#ffffff",
		},
	},
	"cyberpunk": {
		Name:       "Cyberpunk Neon",
		Foreground: "#00ffcc",
		Background: "#0a0a14",
		Palette: []string{
			"#05050f", "#ff0055", "#00ffcc", "#ffea00",
			"#00baff", "#d600ff", "#00ffff", "#e2e2f0",
			"#1e1e3f", "#ff3377", "#33ffdd", "#ffee33",
			"#33cbff", "#dd33ff", "#33ffff", "#ffffff",
		},
	},
	"nord": {
		Name:       "Nordic Frost",
		Foreground: "#d8dee9",
		Background: "#2e3440",
		Palette: []string{
			"#3b4252", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#88c0d0", "#e5e9f0",
			"#4c566a", "#bf616a", "#a3be8c", "#ebcb8b",
			"#81a1c1", "#b48ead", "#8fbcbb", "#eceff4",
		},
	},
}

const TerminalCSS = `
/* ─── Premium Glassmorphism Terminal Window Style ─── */
window.terminal-window {
	background: linear-gradient(135deg, #06060c 0%, #101020 50%, #080d1a 100%);
	color: #e0e0ea;
}

headerbar {
	background: linear-gradient(90deg, rgba(12, 12, 28, 0.95), rgba(18, 18, 38, 0.95));
	border-bottom: 1px solid rgba(0, 255, 204, 0.15);
	box-shadow: 0 4px 30px rgba(0, 0, 0, 0.7);
	min-height: 48px;
	padding: 0 16px;
}

headerbar title {
	font-family: 'Outfit', 'Inter', 'Cantarell', sans-serif;
	font-size: 14px;
	font-weight: 700;
	color: #00ffcc;
	letter-spacing: 2px;
	text-transform: uppercase;
	text-shadow: 0 0 10px rgba(0, 255, 204, 0.4);
}

/* ─── Tab Bar & Tabs ─── */
notebook {
	background: transparent;
	border: none;
}

notebook header {
	background: rgba(10, 10, 22, 0.8);
	border-bottom: 1px solid rgba(0, 255, 204, 0.1);
	padding: 4px 8px 0 8px;
}

notebook tab {
	background: rgba(20, 20, 40, 0.5);
	border: 1px solid rgba(255, 255, 255, 0.05);
	border-bottom: none;
	border-radius: 12px 12px 0 0;
	padding: 8px 18px;
	margin-right: 4px;
	color: rgba(220, 220, 240, 0.6);
	font-family: 'Outfit', 'Inter', sans-serif;
	font-weight: 500;
	font-size: 13px;
	transition: all 200ms ease;
}

notebook tab:hover {
	background: rgba(30, 30, 60, 0.7);
	color: #ffffff;
}

notebook tab:checked {
	background: linear-gradient(180deg, rgba(30, 40, 80, 0.7) 0%, rgba(15, 20, 45, 0.8) 100%);
	border-top: 2px solid #00ffcc;
	border-left: 1px solid rgba(0, 255, 204, 0.2);
	border-right: 1px solid rgba(0, 255, 204, 0.2);
	color: #00ffcc;
	font-weight: 700;
	box-shadow: 0 -4px 15px rgba(0, 255, 204, 0.1);
}

/* Tab Close Button */
notebook tab button {
	background: transparent;
	border: none;
	border-radius: 50%;
	color: rgba(255, 255, 255, 0.4);
	margin-left: 8px;
	padding: 2px;
	transition: all 150ms ease;
}

notebook tab button:hover {
	background: rgba(255, 50, 100, 0.2);
	color: #ff3377;
}

/* ─── VTE Terminal Widget Container ─── */
.terminal-container {
	background: rgba(8, 8, 16, 0.75);
	border: 1px solid rgba(0, 255, 204, 0.08);
	border-radius: 16px;
	box-shadow: inset 0 0 40px rgba(0, 0, 0, 0.9);
	padding: 8px;
	margin: 12px;
}

/* ─── Premium Glassmorphism Right Sidebar ─── */
.sidebar {
	background: linear-gradient(180deg, rgba(12, 12, 28, 0.75) 0%, rgba(8, 8, 18, 0.75) 100%);
	border-left: 1px solid rgba(0, 255, 204, 0.12);
	box-shadow: -10px 0 30px rgba(0, 0, 0, 0.6);
	backdrop-filter: blur(20px);
	padding: 20px;
	min-width: 260px;
}

.sidebar-title {
	font-family: 'Outfit', 'Inter', sans-serif;
	font-size: 12px;
	font-weight: 800;
	color: #00ffcc;
	letter-spacing: 2px;
	text-transform: uppercase;
	margin-bottom: 20px;
	text-shadow: 0 0 8px rgba(0, 255, 204, 0.3);
}

.monitor-card {
	background: linear-gradient(135deg, rgba(25, 25, 55, 0.4), rgba(15, 15, 35, 0.4));
	border: 1px solid rgba(0, 255, 204, 0.1);
	border-radius: 14px;
	padding: 16px;
	margin-bottom: 14px;
	box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
}

.monitor-value {
	font-family: 'JetBrains Mono', monospace;
	font-size: 20px;
	font-weight: 700;
	color: #ffffff;
}

.monitor-label {
	font-family: 'Outfit', 'Inter', sans-serif;
	font-size: 11px;
	font-weight: 600;
	color: rgba(180, 180, 220, 0.6);
	letter-spacing: 1.5px;
	text-transform: uppercase;
	margin-top: 4px;
}

/* ─── ProgressBar inside sidebar ─── */
.sidebar-progress {
	background: rgba(30, 30, 60, 0.5);
	border-radius: 6px;
	min-height: 6px;
}

.sidebar-progress trough {
	background: rgba(30, 30, 60, 0.5);
	border-radius: 6px;
	min-height: 6px;
}

.sidebar-progress progress {
	background: linear-gradient(90deg, #00ffcc, #d600ff);
	border-radius: 6px;
	min-height: 6px;
}

/* ─── Quick action buttons ─── */
.sidebar-btn {
	background: linear-gradient(135deg, rgba(0, 255, 204, 0.1), rgba(214, 0, 255, 0.05));
	border: 1px solid rgba(0, 255, 204, 0.2);
	border-radius: 12px;
	color: #00ffcc;
	font-family: 'Outfit', 'Inter', sans-serif;
	font-size: 12px;
	font-weight: 600;
	padding: 10px;
	margin-bottom: 8px;
	transition: all 200ms ease;
}

.sidebar-btn:hover {
	background: linear-gradient(135deg, rgba(0, 255, 204, 0.25), rgba(214, 0, 255, 0.2));
	border-color: #00ffcc;
	box-shadow: 0 0 15px rgba(0, 255, 204, 0.3);
	color: #ffffff;
}

/* ─── Search / Action bar in Title Bar ─── */
.header-btn {
	background: rgba(255, 255, 255, 0.05);
	border: 1px solid rgba(255, 255, 255, 0.1);
	border-radius: 10px;
	color: #ffffff;
	padding: 6px 12px;
	margin: 0 4px;
	transition: all 150ms ease;
}

.header-btn:hover {
	background: rgba(0, 255, 204, 0.15);
	border-color: rgba(0, 255, 204, 0.3);
	color: #00ffcc;
}

/* ─── Status pill in sidebar ─── */
.sidebar-status {
	background: rgba(0, 255, 204, 0.08);
	border: 1px solid rgba(0, 255, 204, 0.25);
	border-radius: 100px;
	color: #00ffcc;
	font-size: 11px;
	font-weight: 700;
	padding: 4px 12px;
	letter-spacing: 1px;
	text-transform: uppercase;
}
`

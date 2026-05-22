package window

import (
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/config"
	"zet-terminal/internal/theme"
)

type TerminalWindow struct {
	Win            *gtk.ApplicationWindow
	App            *gtk.Application
	Cfg            *config.Config
	Container      *gtk.Box
	Overlay        *gtk.Overlay
	DockBox        *gtk.Box
	Revealer       *gtk.Revealer
	HistoryListBox *gtk.Box
	SidebarPinned  bool

	// Tab Management Fields
	Stack              *gtk.Stack
	TabInstances       map[string]*TabInstance
	ActiveTabID        string
	WorkspaceBox       *gtk.Box
	HistoryRevealer    *gtk.Revealer
	HistoryHeaderBtn   *gtk.Button
	HistoryArrowLbl    *gtk.Label
	HistorySearchEntry *gtk.Entry
	HistoryCommands    []string
	isCreatingGroup    bool
	renamingGroupIndex int    // -1 if not renaming
	renamingTabID      string // empty if not renaming
	sidebarTimeoutID   glib.SourceHandle
	titleUpdateID      glib.SourceHandle
}

var (
	activeWindows        = make(map[*TerminalWindow]bool)
	GlobalCSSProvider    *gtk.CSSProvider
	backgroundPinnedTabs = make(map[string]*TabInstance)
	appIsHeld            bool
)

func PutBackgroundPinnedTab(id string, tab *TabInstance) {
	backgroundPinnedTabs[id] = tab
}

func GetBackgroundPinnedTab(id string) (*TabInstance, bool) {
	tab, exists := backgroundPinnedTabs[id]
	if exists {
		delete(backgroundPinnedTabs, id)
		return tab, true
	}
	return nil, false
}

func UpdateAppHoldStatus(app *gtk.Application, cfg *config.Config) {
	hasPinned := false
	for _, group := range cfg.TabGroups {
		for _, tab := range group.Tabs {
			if tab.Pinned {
				hasPinned = true
				break
			}
		}
	}

	if hasPinned && !appIsHeld {
		app.Hold()
		appIsHeld = true
	} else if !hasPinned && appIsHeld {
		app.Release()
		appIsHeld = false
	}
}

func InitGlobalCSS(cfg *config.Config) {
	GlobalCSSProvider = gtk.NewCSSProvider()
	GlobalCSSProvider.LoadFromData(theme.GetCSS(cfg))
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		GlobalCSSProvider,
		gtk.STYLE_PROVIDER_PRIORITY_USER,
	)
}

func cloneConfig(cfg *config.Config) *config.Config {
	if cfg == nil {
		return config.DefaultConfig()
	}
	dup := *cfg
	dup.CommandHistory = append([]string(nil), cfg.CommandHistory...)
	dup.TermPalette = append([]string(nil), cfg.TermPalette...)
	dup.TabGroups = make([]config.GroupConfig, len(cfg.TabGroups))
	for i, group := range cfg.TabGroups {
		dup.TabGroups[i] = group
		dup.TabGroups[i].Tabs = append([]config.TabConfig(nil), group.Tabs...)
	}
	return &dup
}

func broadcastConfig(cfg *config.Config) {
	for w := range activeWindows {
		w.Cfg = cloneConfig(cfg)
		w.renderWorkspace()
		if w.ActiveTabID != "" {
			if tab, ok := w.TabInstances[w.ActiveTabID]; ok {
				w.Win.SetTitle(tab.Name + " - Terminal")
			}
		}
	}
}

func syncWindowConfigsFromDisk() *config.Config {
	cfg := config.LoadConfig()
	broadcastConfig(cfg)
	return cfg
}

func NewTerminalWindow(app *gtk.Application, cfg *config.Config, initialActiveTabID, initialActiveTabDir string) *TerminalWindow {
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("Terminal")
	if cfg.WindowWidth > 0 && cfg.WindowHeight > 0 {
		win.SetDefaultSize(cfg.WindowWidth, cfg.WindowHeight)
	} else {
		win.SetDefaultSize(1100, 720)
	}
	win.AddCSSClass("terminal-window")

	tw := &TerminalWindow{
		Win:                win,
		App:                app,
		Cfg:                cfg,
		TabInstances:       make(map[string]*TabInstance),
		renamingGroupIndex: -1,
	}

	activeWindows[tw] = true
	win.ConnectCloseRequest(func() bool {
		isLastWindow := len(activeWindows) == 1

		for id, tab := range tw.TabInstances {
			pinned := tw.isTabPinned(id)
			if pinned && isLastWindow {
				tw.Stack.Remove(tab.TermInst.Widget)
				PutBackgroundPinnedTab(id, tab)

				// Re-bind callbacks to be safe in background
				tab.TermInst.OnChildExited(func(status int) {
					tab.TermInst.Destroy()
					delete(backgroundPinnedTabs, id)
					tw.removeTabFromConfig(id)
					UpdateAppHoldStatus(tw.App, tw.Cfg)
				})
				tab.TermInst.OnWindowTitleChanged(nil)
			} else {
				tab.TermInst.Destroy()
				tw.Stack.Remove(tab.TermInst.Widget)
			}
		}
		tw.TabInstances = make(map[string]*TabInstance)
		if tw.sidebarTimeoutID != 0 {
			glib.SourceRemove(tw.sidebarTimeoutID)
			tw.sidebarTimeoutID = 0
		}
		if tw.titleUpdateID != 0 {
			glib.SourceRemove(tw.titleUpdateID)
			tw.titleUpdateID = 0
		}

		delete(activeWindows, tw)
		UpdateAppHoldStatus(tw.App, tw.Cfg)

		win.Destroy()
		return true
	})

	win.ConnectDestroy(func() {
		if tw.sidebarTimeoutID != 0 {
			glib.SourceRemove(tw.sidebarTimeoutID)
			tw.sidebarTimeoutID = 0
		}
		if tw.titleUpdateID != 0 {
			glib.SourceRemove(tw.titleUpdateID)
			tw.titleUpdateID = 0
		}
		delete(activeWindows, tw)
		UpdateAppHoldStatus(tw.App, tw.Cfg)
	})

	UpdateAppHoldStatus(app, cfg)

	// If no initial active tab ID is provided, but a directory is provided,
	// resolve the first tab to be the target for the directory.
	if initialActiveTabID == "" && initialActiveTabDir != "" {
		for _, group := range cfg.TabGroups {
			if len(group.Tabs) > 0 {
				initialActiveTabID = group.Tabs[0].ID
				break
			}
		}
	}

	tw.setupUI(initialActiveTabID, initialActiveTabDir)
	tw.setupShortcuts()

	// Show and present window to grab focus at OS/WM level
	win.Present()

	// Pick the first loaded tab to activate, or the requested initial active tab
	activeID := initialActiveTabID
	if activeID == "" {
		for _, group := range tw.Cfg.TabGroups {
			if len(group.Tabs) > 0 {
				activeID = group.Tabs[0].ID
				break
			}
		}
	}
	if activeID != "" {
		tw.ActivateTab(activeID)
	}

	return tw
}

func (tw *TerminalWindow) setupUI(initialActiveTabID, initialActiveTabDir string) {
	// Minimalistic Title Bar
	header := gtk.NewHeaderBar()
	titleLabel := gtk.NewLabel("Terminal")
	titleLabel.AddCSSClass("title-label")
	titleLabel.SetHAlign(gtk.AlignCenter)
	titleLabel.SetVAlign(gtk.AlignCenter)
	header.SetTitleWidget(titleLabel)
	tw.Win.SetTitlebar(header)

	// Add Sidebar Toggle button to start of header
	btnSidebar := gtk.NewButton()
	btnSidebar.AddCSSClass("header-btn")
	btnSidebar.SetTooltipText("Toggle Sidebar (Ctrl+B)")
	imgSidebar := gtk.NewImageFromIconName("view-sidebar-symbolic")
	imgSidebar.SetPixelSize(14)
	btnSidebar.SetChild(imgSidebar)
	btnSidebar.ConnectClicked(func() {
		tw.SidebarPinned = !tw.SidebarPinned
		tw.Revealer.SetRevealChild(tw.SidebarPinned)
		tw.Cfg.SidebarPinned = tw.SidebarPinned
		_ = config.SaveConfig(tw.Cfg)
		broadcastConfig(tw.Cfg)
		if tw.SidebarPinned {
			tw.renderWorkspace()
		}
	})
	header.PackStart(btnSidebar)

	// Add New Tab button to start of header
	btnNewTab := gtk.NewButton()
	btnNewTab.AddCSSClass("header-btn")
	btnNewTab.SetTooltipText("Create New Tab")
	imgNewTab := gtk.NewImageFromIconName("tab-new-symbolic")
	imgNewTab.SetPixelSize(14)
	btnNewTab.SetChild(imgNewTab)
	btnNewTab.ConnectClicked(func() {
		if len(tw.Cfg.TabGroups) > 0 {
			tw.createTabInGroup(tw.Cfg.TabGroups[0].ID)
		}
	})
	header.PackStart(btnNewTab)

	// Add Settings button to end of header
	btnSettings := gtk.NewButton()
	btnSettings.AddCSSClass("header-btn")
	btnSettings.SetTooltipText("Settings")
	imgSettings := gtk.NewImageFromIconName("preferences-system-symbolic")
	imgSettings.SetPixelSize(14)
	btnSettings.SetChild(imgSettings)
	btnSettings.ConnectClicked(func() {
		tw.openSettingsDialog()
	})
	header.PackEnd(btnSettings)

	// Stack for Switchable VTE widgets
	tw.Stack = gtk.NewStack()
	tw.Stack.SetTransitionType(gtk.StackTransitionTypeNone)
	tw.Stack.SetHExpand(true)
	tw.Stack.SetVExpand(true)

	// Container box
	tw.Container = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.Container.AddCSSClass("terminal-container")
	tw.Container.SetHExpand(true)
	tw.Container.SetVExpand(true)
	tw.Container.Append(tw.Stack)

	// Wrap in Overlay for side dock
	tw.Overlay = gtk.NewOverlay()
	tw.Overlay.SetChild(tw.Container)
	tw.Win.SetChild(tw.Overlay)

	// Load tabs from persistent config groups
	hasTabsLoaded := false
	for _, group := range tw.Cfg.TabGroups {
		for _, tab := range group.Tabs {
			if bgTab, exists := GetBackgroundPinnedTab(tab.ID); exists {
				tw.RestoreBackgroundTab(bgTab, group.ID)
			} else {
				dir := ""
				if tab.ID == initialActiveTabID {
					dir = initialActiveTabDir
				}
				tw.CreateTab(tab.ID, tab.Name, group.ID, dir, false)
			}
			hasTabsLoaded = true
		}
	}

	// Default fallback tab structure if no tabs loaded
	if !hasTabsLoaded {
		tw.createDefaultTabStructure()
	}

	// Set up the left hover dock (workspace tabs & past commands)
	tw.setupSideDock()

	// Periodic title auto-updater (every 1 second)
	tw.titleUpdateID = glib.TimeoutAdd(1000, func() bool {
		if _, exists := activeWindows[tw]; !exists {
			tw.titleUpdateID = 0
			return false
		}
		tw.updateTabAutoTitles()
		return true
	})
}

func (tw *TerminalWindow) setupShortcuts() {
	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		// New Tab
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.NewTab) {
			tw.createTabInGroup(tw.GetActiveTabGroupID())
			return true
		}

		// New Window
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.NewWindow) {
			dir := tw.getActiveTabDir()
			groupID := tw.GetActiveTabGroupID()
			tabID := tw.CreateNewTabInGroup(groupID, dir, nil)
			NewTerminalWindow(tw.App, tw.Cfg, tabID, dir)
			return true
		}

		// Close Tab
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.CloseTab) {
			if tw.ActiveTabID != "" {
				tw.CloseTab(tw.ActiveTabID)
			}
			return true
		}

		// Next Tab
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.NextTab) {
			tw.CycleTab(true)
			return true
		}

		// Previous Tab
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.PrevTab) {
			tw.CycleTab(false)
			return true
		}

		// Copy
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.Copy) {
			if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
				activeTab.TermInst.Copy()
			}
			return true
		}

		// Paste
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.Paste) {
			if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
				activeTab.TermInst.Paste()
			}
			return true
		}

		// Toggle Sidebar
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.ToggleSidebar) {
			tw.SidebarPinned = !tw.SidebarPinned
			tw.Revealer.SetRevealChild(tw.SidebarPinned)
			if tw.SidebarPinned {
				tw.renderWorkspace()
			}
			return true
		}

		// Zoom In
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.ZoomIn) {
			tw.Cfg.FontSize++
			if tw.Cfg.FontSize > 72 {
				tw.Cfg.FontSize = 72
			}
			_ = config.SaveConfig(tw.Cfg)
			for _, tab := range tw.TabInstances {
				tw.applyConfigToInstance(tab.TermInst)
			}
			return true
		}

		// Zoom Out
		if MatchShortcut(keyval, state, tw.Cfg.Keybindings.ZoomOut) {
			tw.Cfg.FontSize--
			if tw.Cfg.FontSize < 4 {
				tw.Cfg.FontSize = 4
			}
			_ = config.SaveConfig(tw.Cfg)
			for _, tab := range tw.TabInstances {
				tw.applyConfigToInstance(tab.TermInst)
			}
			return true
		}

		return false
	})
	tw.Win.AddController(keyCtrl)
}

func ActiveWindowsCount() int {
	return len(activeWindows)
}

func FocusActiveWindow() {
	for tw := range activeWindows {
		tw.Win.Present()
		break
	}
}

func GetFirstActiveWindow() *TerminalWindow {
	for tw := range activeWindows {
		return tw
	}
	return nil
}

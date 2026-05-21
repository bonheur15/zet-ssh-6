package window

import (
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/config"
)

type TerminalWindow struct {
	Win              *gtk.ApplicationWindow
	App              *gtk.Application
	Cfg              *config.Config
	Container        *gtk.Box
	Overlay          *gtk.Overlay
	DockBox          *gtk.Box
	Revealer         *gtk.Revealer
	HistoryListBox   *gtk.Box
	SidebarPinned    bool

	// Tab Management Fields
	Stack            *gtk.Stack
	TabInstances     map[string]*TabInstance
	ActiveTabID      string
	WorkspaceBox     *gtk.Box
	HistoryRevealer  *gtk.Revealer
	HistoryHeaderBtn *gtk.Button
	HistoryArrowLbl  *gtk.Label
	PopoverActive    bool
	sidebarTimeoutID glib.SourceHandle
}

func NewTerminalWindow(app *gtk.Application, cfg *config.Config) *TerminalWindow {
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("Terminal")
	win.SetDefaultSize(850, 550)
	win.AddCSSClass("terminal-window")

	tw := &TerminalWindow{
		Win:          win,
		App:          app,
		Cfg:          cfg,
		TabInstances: make(map[string]*TabInstance),
	}

	tw.setupUI()
	tw.setupShortcuts()

	// Show and present window to grab focus at OS/WM level
	win.Present()

	// Pick the first loaded tab to activate
	var firstTabID string
	for _, group := range tw.Cfg.TabGroups {
		if len(group.Tabs) > 0 {
			firstTabID = group.Tabs[0].ID
			break
		}
	}
	if firstTabID != "" {
		tw.ActivateTab(firstTabID)
	}

	return tw
}

func (tw *TerminalWindow) setupUI() {
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
			tw.CreateTab(tab.ID, tab.Name, group.ID, false)
			hasTabsLoaded = true
		}
	}

	// Default fallback tab structure if no tabs loaded
	if !hasTabsLoaded {
		tw.createDefaultTabStructure()
	}

	// Set up the left hover dock (workspace tabs & past commands)
	tw.setupSideDock()
}

func (tw *TerminalWindow) setupShortcuts() {
	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		isCtrl := (state & gdk.ControlMask) != 0
		isShift := (state & gdk.ShiftMask) != 0
		isAlt := (state & gdk.AltMask) != 0

		if isCtrl && isShift {
			switch keyval {
			case 'N', 'n': // Ctrl+Shift+N -> New Window
				NewTerminalWindow(tw.App, tw.Cfg)
				return true
			case 'C', 'c': // Ctrl+Shift+C -> Copy active selection
				if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
					activeTab.TermInst.Copy()
				}
				return true
			case 'V', 'v': // Ctrl+Shift+V -> Paste system clipboard
				if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
					activeTab.TermInst.Paste()
				}
				return true
			case '+', '=': // Ctrl+Shift++ or Ctrl+Shift+= -> Zoom In
				tw.Cfg.FontSize++
				if tw.Cfg.FontSize > 72 {
					tw.Cfg.FontSize = 72
				}
				_ = config.SaveConfig(tw.Cfg)
				for _, tab := range tw.TabInstances {
					tw.applyConfigToInstance(tab.TermInst)
				}
				return true
			case '-': // Ctrl+Shift+- -> Zoom Out
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
		}

		if isCtrl && !isShift && !isAlt {
			switch keyval {
			case 'B', 'b': // Ctrl+B -> Toggle Sidebar
				tw.SidebarPinned = !tw.SidebarPinned
				tw.Revealer.SetRevealChild(tw.SidebarPinned)
				if tw.SidebarPinned {
					tw.renderWorkspace()
				}
				return true
			}
		}

		return false
	})
	tw.Win.AddController(keyCtrl)
}

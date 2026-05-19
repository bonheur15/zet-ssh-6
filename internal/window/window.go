package window

import (
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/config"
	"zet-terminal/internal/terminal"
	"zet-terminal/internal/theme"
)

type TabInstance struct {
	ID       string
	Name     string
	TermInst *terminal.VteTerminalInstance
	GroupID  string
}

type TerminalWindow struct {
	Win              *gtk.ApplicationWindow
	App              *gtk.Application
	Cfg              *config.Config
	Container        *gtk.Box
	Overlay          *gtk.Overlay
	DockBox          *gtk.Box
	Revealer         *gtk.Revealer
	HistoryListBox   *gtk.Box
	currentCmdBuffer string

	// Tab Management Fields
	Stack            *gtk.Stack
	TabInstances     map[string]*TabInstance
	ActiveTabID      string
	WorkspaceBox     *gtk.Box
	HistoryRevealer  *gtk.Revealer
	HistoryHeaderBtn *gtk.Button
	HistoryArrowLbl  *gtk.Label
	PopoverActive    bool
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
	titleLabel.SetHAlign(gtk.AlignCenter)
	titleLabel.SetVAlign(gtk.AlignCenter)
	header.SetTitleWidget(titleLabel)
	tw.Win.SetTitlebar(header)

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

func (tw *TerminalWindow) applyConfigToInstance(inst *terminal.VteTerminalInstance) {
	inst.SetFont(tw.Cfg.FontName, tw.Cfg.FontSize)
	inst.SetColors(theme.DefaultPalette.Foreground, theme.DefaultPalette.Background, theme.DefaultPalette.Palette)
}

func (tw *TerminalWindow) CreateTab(id, name, groupID string, saveToConfig bool) *TabInstance {
	inst := terminal.NewVteTerminal()
	inst.Widget.AddCSSClass("vte-terminal-widget")
	inst.Widget.SetHExpand(true)
	inst.Widget.SetVExpand(true)
	inst.SetScrollbackLines(tw.Cfg.ScrollbackLines)
	inst.SetCursorBlinkMode(tw.Cfg.CursorBlinkMode)
	inst.SetCursorShape(tw.Cfg.CursorShape)

	tw.applyConfigToInstance(inst)

	tab := &TabInstance{
		ID:       id,
		Name:     name,
		TermInst: inst,
		GroupID:  groupID,
	}

	tw.TabInstances[id] = tab

	// Add widget to the Stack
	tw.Stack.AddChild(inst.Widget)

	// Spawn shell
	inst.SpawnShell(tw.Cfg.Shell, "")

	// Signal Handlers
	inst.OnChildExited(func(status int) {
		inst.Destroy()
		tw.CloseTab(id)
	})

	inst.OnWindowTitleChanged(func(title string) {
		title = strings.TrimSpace(title)
		if title == "" {
			title = "Console"
		}
		// Automatically derive tab name from the active window title
		tab.Name = title
		tw.updateTabNameInConfig(id, title)
		_ = config.SaveConfig(tw.Cfg)

		if tw.ActiveTabID == id {
			tw.Win.SetTitle(title + " - Terminal")
		}
		tw.renderWorkspace()
	})

	// Setup context menu popovers (Copy & Paste, no emojis)
	tw.setupContextMenuForTab(tab)

	if saveToConfig {
		tw.saveTabsToConfig()
		tw.renderWorkspace()
	}

	return tab
}

func (tw *TerminalWindow) ActivateTab(id string) {
	tab, ok := tw.TabInstances[id]
	if !ok {
		return
	}

	tw.ActiveTabID = id
	tw.Stack.SetVisibleChild(tab.TermInst.Widget)

	// Update window titlebar
	tw.Win.SetTitle(tab.Name + " - Terminal")

	tw.renderWorkspace()

	// Grab input focus directly on active terminal instance
	tab.TermInst.Widget.GrabFocus()
}

func (tw *TerminalWindow) CloseTab(id string) {
	tab, ok := tw.TabInstances[id]
	if !ok {
		return
	}

	// Destroy process/bindings
	tab.TermInst.Destroy()

	tw.Stack.Remove(tab.TermInst.Widget)
	delete(tw.TabInstances, id)

	tw.removeTabFromConfig(id)

	if len(tw.TabInstances) == 0 {
		tw.Win.Close()
		return
	}

	if tw.ActiveTabID == id {
		// Pick first available tab
		for nextID := range tw.TabInstances {
			tw.ActivateTab(nextID)
			break
		}
	} else {
		tw.renderWorkspace()
	}
}

func (tw *TerminalWindow) CloseTabSilently(id string) {
	tab, ok := tw.TabInstances[id]
	if !ok {
		return
	}
	tab.TermInst.Destroy()
	tw.Stack.Remove(tab.TermInst.Widget)
	delete(tw.TabInstances, id)
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

		// Track command history typing silently in the background
		if !isCtrl && !isAlt {
			switch keyval {
			case 0xff0d, 0xff8d: // Enter key
				tw.AddCommandToHistory(tw.currentCmdBuffer)
				tw.currentCmdBuffer = ""
			case 0xff08: // Backspace key
				if len(tw.currentCmdBuffer) > 0 {
					runes := []rune(tw.currentCmdBuffer)
					tw.currentCmdBuffer = string(runes[:len(runes)-1])
				}
			default:
				if keyval >= 32 && keyval <= 126 {
					tw.currentCmdBuffer += string(rune(keyval))
				}
			}
		} else if isCtrl && (keyval == 'C' || keyval == 'c' || keyval == 'D' || keyval == 'd') {
			tw.currentCmdBuffer = ""
		}

		return false
	})
	tw.Win.AddController(keyCtrl)
}

func (tw *TerminalWindow) setupSideDock() {
	tw.Revealer = gtk.NewRevealer()
	tw.Revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideRight)
	tw.Revealer.SetTransitionDuration(250)
	tw.Revealer.SetRevealChild(false)

	// Panel Container Box
	panelBox := gtk.NewBox(gtk.OrientationVertical, 0)
	panelBox.AddCSSClass("sidebar-panel")
	panelBox.SetSizeRequest(220, -1)
	panelBox.SetVExpand(true)

	// Workspace Header
	wsHeader := gtk.NewLabel("Workspace")
	wsHeader.AddCSSClass("workspace-header")
	wsHeader.SetHAlign(gtk.AlignStart)
	wsHeader.SetMarginStart(4)
	panelBox.Append(wsHeader)

	// Scrolled window for workspace explorer
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetHExpand(true)
	panelBox.Append(scrolled)

	// Workspace Box inside scrolled
	tw.WorkspaceBox = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.WorkspaceBox.SetVExpand(true)
	tw.WorkspaceBox.SetHExpand(true)
	scrolled.SetChild(tw.WorkspaceBox)

	// Collapsible Section for Past Commands
	tw.setupHistorySection(panelBox)

	tw.Revealer.SetChild(panelBox)

	// Thin vertical hover trigger bar
	triggerBar := gtk.NewBox(gtk.OrientationVertical, 0)
	triggerBar.AddCSSClass("sidebar-trigger")
	triggerBar.SetSizeRequest(6, -1)
	triggerBar.SetVExpand(true)

	// Main Dock Box holding [ Revealer | TriggerBar ]
	tw.DockBox = gtk.NewBox(gtk.OrientationHorizontal, 0)
	tw.DockBox.AddCSSClass("sidebar-dock")
	tw.DockBox.SetHAlign(gtk.AlignStart)
	tw.DockBox.SetVAlign(gtk.AlignFill)
	tw.DockBox.Append(tw.Revealer)
	tw.DockBox.Append(triggerBar)

	tw.Overlay.AddOverlay(tw.DockBox)

	// Single unified Hover Event Controller on the parent DockBox
	// Eliminates flicker and race conditions when crossing boundaries
	dockMotionCtrl := gtk.NewEventControllerMotion()
	dockMotionCtrl.ConnectEnter(func(x float64, y float64) {
		tw.renderWorkspace()
		tw.Revealer.SetRevealChild(true)
	})
	dockMotionCtrl.ConnectLeave(func() {
		if !tw.PopoverActive {
			tw.Revealer.SetRevealChild(false)
		}
	})
	tw.DockBox.AddController(dockMotionCtrl)
}

func (tw *TerminalWindow) setupHistorySection(parentBox *gtk.Box) {
	tw.HistoryHeaderBtn = gtk.NewButton()
	tw.HistoryHeaderBtn.AddCSSClass("history-section-header")
	tw.HistoryHeaderBtn.SetHAlign(gtk.AlignFill)

	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	headerBox.SetHExpand(true)

	tw.HistoryArrowLbl = gtk.NewLabel("▶")
	tw.HistoryArrowLbl.AddCSSClass("history-header-arrow")
	headerBox.Append(tw.HistoryArrowLbl)

	title := gtk.NewLabel("Past Commands")
	title.AddCSSClass("history-header-title")
	headerBox.Append(title)

	tw.HistoryHeaderBtn.SetChild(headerBox)
	parentBox.Append(tw.HistoryHeaderBtn)

	tw.HistoryRevealer = gtk.NewRevealer()
	tw.HistoryRevealer.SetTransitionType(gtk.RevealerTransitionTypeSlideDown)
	tw.HistoryRevealer.SetTransitionDuration(200)
	tw.HistoryRevealer.SetRevealChild(false)

	tw.HistoryListBox = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.HistoryListBox.SetHExpand(true)
	tw.HistoryRevealer.SetChild(tw.HistoryListBox)
	parentBox.Append(tw.HistoryRevealer)

	tw.HistoryHeaderBtn.ConnectClicked(func() {
		isExpanded := tw.HistoryRevealer.RevealChild()
		tw.HistoryRevealer.SetRevealChild(!isExpanded)
		if !isExpanded {
			tw.HistoryArrowLbl.SetLabel("▼")
			tw.updateHistoryUI()
		} else {
			tw.HistoryArrowLbl.SetLabel("▶")
		}
	})
}

func (tw *TerminalWindow) renderWorkspace() {
	for child := tw.WorkspaceBox.FirstChild(); child != nil; child = tw.WorkspaceBox.FirstChild() {
		tw.WorkspaceBox.Remove(child)
	}

	// Create Group Button
	btnNewGroup := gtk.NewButton()
	btnNewGroup.AddCSSClass("workspace-action-btn")
	lblNewGroup := gtk.NewLabel("+ New Group")
	lblNewGroup.AddCSSClass("workspace-action-label")
	btnNewGroup.SetChild(lblNewGroup)
	btnNewGroup.ConnectClicked(func() {
		tw.promptCreateGroup()
	})
	tw.WorkspaceBox.Append(btnNewGroup)

	// Render Groups and Tabs
	for gIdx, group := range tw.Cfg.TabGroups {
		groupConfig := group
		groupIndex := gIdx

		groupContainer := gtk.NewBox(gtk.OrientationVertical, 0)

		headerRow := gtk.NewBox(gtk.OrientationHorizontal, 0)
		headerRow.AddCSSClass("group-header-row")
		headerRow.SetHExpand(true)

		btnToggle := gtk.NewButton()
		btnToggle.AddCSSClass("group-toggle-btn")
		arrowStr := "▼"
		if groupConfig.Collapsed {
			arrowStr = "▶"
		}
		lblToggle := gtk.NewLabel(arrowStr)
		lblToggle.AddCSSClass("group-toggle-label")
		btnToggle.SetChild(lblToggle)
		headerRow.Append(btnToggle)

		lblTitle := gtk.NewLabel(groupConfig.Name)
		lblTitle.AddCSSClass("group-title-label")
		lblTitle.SetHAlign(gtk.AlignStart)
		lblTitle.SetHExpand(true)
		headerRow.Append(lblTitle)

		// Create New Tab inside this group (+ Tab)
		btnNewTab := gtk.NewButton()
		btnNewTab.AddCSSClass("group-action-btn")
		btnNewTab.SetTooltipText("Add tab")
		imgNewTab := gtk.NewImageFromIconName("list-add-symbolic")
		btnNewTab.SetChild(imgNewTab)
		btnNewTab.ConnectClicked(func() {
			tw.createTabInGroup(groupConfig.ID)
		})
		headerRow.Append(btnNewTab)

		// Rename Group button (Rename)
		btnRenameGroup := gtk.NewButton()
		btnRenameGroup.AddCSSClass("group-action-btn")
		btnRenameGroup.SetTooltipText("Rename group")
		imgRenameGroup := gtk.NewImageFromIconName("document-edit-symbolic")
		btnRenameGroup.SetChild(imgRenameGroup)
		btnRenameGroup.ConnectClicked(func() {
			tw.promptRenameGroup(groupIndex)
		})
		headerRow.Append(btnRenameGroup)

		// Delete Group button (Delete)
		btnDelGroup := gtk.NewButton()
		btnDelGroup.AddCSSClass("group-action-btn")
		btnDelGroup.SetTooltipText("Delete group")
		imgDelGroup := gtk.NewImageFromIconName("user-trash-symbolic")
		btnDelGroup.SetChild(imgDelGroup)
		btnDelGroup.ConnectClicked(func() {
			tw.deleteGroup(groupIndex)
		})
		headerRow.Append(btnDelGroup)

		groupContainer.Append(headerRow)

		childrenBox := gtk.NewBox(gtk.OrientationVertical, 2)
		childrenBox.SetVisible(!groupConfig.Collapsed)

		btnToggle.ConnectClicked(func() {
			tw.Cfg.TabGroups[groupIndex].Collapsed = !tw.Cfg.TabGroups[groupIndex].Collapsed
			_ = config.SaveConfig(tw.Cfg)
			tw.renderWorkspace()
		})

		for _, tabCfg := range groupConfig.Tabs {
			tabConfig := tabCfg

			tabRow := gtk.NewBox(gtk.OrientationHorizontal, 4)
			tabRow.AddCSSClass("tab-row")
			tabRow.SetHExpand(true)

			if tabConfig.ID == tw.ActiveTabID {
				tabRow.AddCSSClass("active")
			}

			// Main Select Tab Row Button
			btnSelectTab := gtk.NewButton()
			btnSelectTab.AddCSSClass("tab-select-btn")
			btnSelectTab.SetHExpand(true)
			btnSelectTab.SetHAlign(gtk.AlignFill)

			lblTab := gtk.NewLabel(tabConfig.Name)
			lblTab.AddCSSClass("tab-label")
			lblTab.SetHAlign(gtk.AlignStart)
			lblTab.SetXAlign(0.0)
			btnSelectTab.SetChild(lblTab)

			btnSelectTab.ConnectClicked(func() {
				tw.ActivateTab(tabConfig.ID)
			})
			tabRow.Append(btnSelectTab)

			// Fully working, comfortably padded Close button using symbolic cross icon
			btnCloseTab := gtk.NewButton()
			btnCloseTab.AddCSSClass("tab-action-btn")
			btnCloseTab.SetTooltipText("Close tab")
			imgCloseTab := gtk.NewImageFromIconName("window-close-symbolic")
			btnCloseTab.SetChild(imgCloseTab)
			btnCloseTab.ConnectClicked(func() {
				tw.CloseTab(tabConfig.ID)
			})
			tabRow.Append(btnCloseTab)

			childrenBox.Append(tabRow)
		}

		groupContainer.Append(childrenBox)
		tw.WorkspaceBox.Append(groupContainer)
	}
}

func (tw *TerminalWindow) promptCreateGroup() {
	tw.PopoverActive = true

	popover := gtk.NewPopover()
	popover.SetParent(tw.WorkspaceBox)
	popover.SetHasArrow(true)

	popover.ConnectClosed(func() {
		tw.PopoverActive = false
		tw.Revealer.SetRevealChild(false)
	})

	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetMarginBottom(4)
	box.SetMarginTop(4)
	box.SetMarginStart(4)
	box.SetMarginEnd(4)

	lbl := gtk.NewLabel("Group Name:")
	lbl.AddCSSClass("menu-item-label")
	lbl.SetHAlign(gtk.AlignStart)
	box.Append(lbl)

	entry := gtk.NewEntry()
	entry.AddCSSClass("sidebar-entry")
	entry.SetPlaceholderText("SSH, Dev, etc.")
	box.Append(entry)

	btnCreate := gtk.NewButton()
	btnCreate.AddCSSClass("workspace-action-btn")
	lblCreate := gtk.NewLabel("Create")
	lblCreate.AddCSSClass("workspace-action-label")
	btnCreate.SetChild(lblCreate)

	btnCreate.ConnectClicked(func() {
		name := entry.Text()
		if name == "" {
			name = "New Group"
		}
		groupID := fmt.Sprintf("group-%d", time.Now().UnixNano())
		newGroup := config.GroupConfig{
			ID:        groupID,
			Name:      name,
			Collapsed: false,
			Tabs:      []config.TabConfig{},
		}
		tw.Cfg.TabGroups = append(tw.Cfg.TabGroups, newGroup)
		_ = config.SaveConfig(tw.Cfg)
		tw.renderWorkspace()
		popover.Popdown()
	})
	box.Append(btnCreate)

	popover.SetChild(box)
	popover.Popup()
}

func (tw *TerminalWindow) createTabInGroup(groupID string) {
	tabID := fmt.Sprintf("tab-%d", time.Now().UnixNano())
	tabName := "Console"

	for i, group := range tw.Cfg.TabGroups {
		if group.ID == groupID {
			tw.Cfg.TabGroups[i].Tabs = append(tw.Cfg.TabGroups[i].Tabs, config.TabConfig{
				ID:   tabID,
				Name: tabName,
			})
			_ = config.SaveConfig(tw.Cfg)

			tw.CreateTab(tabID, tabName, groupID, false)
			tw.ActivateTab(tabID)
			break
		}
	}
}

func (tw *TerminalWindow) promptRenameGroup(groupIndex int) {
	tw.PopoverActive = true

	popover := gtk.NewPopover()
	popover.SetParent(tw.WorkspaceBox)
	popover.SetHasArrow(true)

	popover.ConnectClosed(func() {
		tw.PopoverActive = false
		tw.Revealer.SetRevealChild(false)
	})

	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetMarginBottom(4)
	box.SetMarginTop(4)
	box.SetMarginStart(4)
	box.SetMarginEnd(4)

	lbl := gtk.NewLabel("Rename Group:")
	lbl.AddCSSClass("menu-item-label")
	lbl.SetHAlign(gtk.AlignStart)
	box.Append(lbl)

	entry := gtk.NewEntry()
	entry.AddCSSClass("sidebar-entry")
	entry.SetText(tw.Cfg.TabGroups[groupIndex].Name)
	box.Append(entry)

	btnSave := gtk.NewButton()
	btnSave.AddCSSClass("workspace-action-btn")
	lblSave := gtk.NewLabel("Save")
	lblSave.AddCSSClass("workspace-action-label")
	btnSave.SetChild(lblSave)

	btnSave.ConnectClicked(func() {
		name := entry.Text()
		if name != "" {
			tw.Cfg.TabGroups[groupIndex].Name = name
			_ = config.SaveConfig(tw.Cfg)
			tw.renderWorkspace()
		}
		popover.Popdown()
	})
	box.Append(btnSave)

	popover.SetChild(box)
	popover.Popup()
}

func (tw *TerminalWindow) deleteGroup(groupIndex int) {
	group := tw.Cfg.TabGroups[groupIndex]

	for _, tabCfg := range group.Tabs {
		tw.CloseTabSilently(tabCfg.ID)
	}

	tw.Cfg.TabGroups = append(tw.Cfg.TabGroups[:groupIndex], tw.Cfg.TabGroups[groupIndex+1:]...)
	_ = config.SaveConfig(tw.Cfg)

	if len(tw.TabInstances) == 0 {
		tw.createDefaultTabStructure()
	} else {
		tw.renderWorkspace()
	}
}

func (tw *TerminalWindow) updateTabNameInConfig(tabID, name string) {
	for gIdx, group := range tw.Cfg.TabGroups {
		for tIdx, tab := range group.Tabs {
			if tab.ID == tabID {
				tw.Cfg.TabGroups[gIdx].Tabs[tIdx].Name = name
				return
			}
		}
	}
}

func (tw *TerminalWindow) removeTabFromConfig(tabID string) {
	for gIdx, group := range tw.Cfg.TabGroups {
		for tIdx, tab := range group.Tabs {
			if tab.ID == tabID {
				tw.Cfg.TabGroups[gIdx].Tabs = append(tw.Cfg.TabGroups[gIdx].Tabs[:tIdx], tw.Cfg.TabGroups[gIdx].Tabs[tIdx+1:]...)
				_ = config.SaveConfig(tw.Cfg)
				return
			}
		}
	}
}

func (tw *TerminalWindow) saveTabsToConfig() {
	// Sync changes
}

func (tw *TerminalWindow) createDefaultTabStructure() {
	tw.Cfg.TabGroups = []config.GroupConfig{
		{
			ID:        "group-general",
			Name:      "General Workspace",
			Collapsed: false,
			Tabs: []config.TabConfig{
				{
					ID:   "tab-1",
					Name: "Primary Console",
				},
			},
		},
	}
	_ = config.SaveConfig(tw.Cfg)

	tw.CreateTab("tab-1", "Primary Console", "group-general", false)
	tw.ActivateTab("tab-1")
}

func (tw *TerminalWindow) updateHistoryUI() {
	for child := tw.HistoryListBox.FirstChild(); child != nil; child = tw.HistoryListBox.FirstChild() {
		tw.HistoryListBox.Remove(child)
	}

	if len(tw.Cfg.CommandHistory) == 0 {
		emptyLabel := gtk.NewLabel("No commands run yet.")
		emptyLabel.AddCSSClass("sidebar-btn-label")
		emptyLabel.SetMarginTop(20)
		emptyLabel.SetHAlign(gtk.AlignCenter)
		tw.HistoryListBox.Append(emptyLabel)
		return
	}

	for _, cmd := range tw.Cfg.CommandHistory {
		cmdStr := cmd

		row := gtk.NewBox(gtk.OrientationHorizontal, 0)
		row.AddCSSClass("sidebar-row")
		row.SetHExpand(true)

		btnText := gtk.NewButton()
		btnText.AddCSSClass("sidebar-btn")
		btnText.SetHExpand(true)
		btnText.SetHAlign(gtk.AlignFill)
		btnText.SetTooltipText("Click to copy to clipboard")

		lbl := gtk.NewLabel(cmdStr)
		lbl.AddCSSClass("sidebar-btn-label")
		lbl.SetHAlign(gtk.AlignStart)
		lbl.SetXAlign(0.0)
		btnText.SetChild(lbl)

		btnText.ConnectClicked(func() {
			display := gdk.DisplayGetDefault()
			if display != nil {
				clipboard := display.Clipboard()
				if clipboard != nil {
					clipboard.SetText(cmdStr)
				}
			}
		})
		row.Append(btnText)

		btnArrow := gtk.NewButton()
		btnArrow.AddCSSClass("sidebar-arrow-btn")
		btnArrow.SetTooltipText("Execute command")

		lblArrow := gtk.NewLabel("→")
		lblArrow.AddCSSClass("sidebar-arrow-label")
		btnArrow.SetChild(lblArrow)

		btnArrow.ConnectClicked(func() {
			if tw.ActiveTabID != "" {
				if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
					activeTab.TermInst.FeedChild(cmdStr + "\n")
					tw.Revealer.SetRevealChild(false)
					activeTab.TermInst.Widget.GrabFocus()
				}
			}
		})
		row.Append(btnArrow)

		tw.HistoryListBox.Append(row)
	}
}

func (tw *TerminalWindow) AddCommandToHistory(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) < 2 || len(cmd) > 60 {
		return
	}

	cmdLower := strings.ToLower(cmd)
	if cmdLower == "ls" || cmdLower == "clear" || cmdLower == "top" {
		return
	}

	if len(tw.Cfg.CommandHistory) > 0 && tw.Cfg.CommandHistory[0] == cmd {
		return
	}

	tw.Cfg.CommandHistory = append([]string{cmd}, tw.Cfg.CommandHistory...)

	if len(tw.Cfg.CommandHistory) > 10 {
		tw.Cfg.CommandHistory = tw.Cfg.CommandHistory[:10]
	}

	_ = config.SaveConfig(tw.Cfg)
}

func (tw *TerminalWindow) setupContextMenuForTab(tab *TabInstance) {
	popover := gtk.NewPopover()
	popover.SetParent(tab.TermInst.Widget)
	popover.SetHasArrow(true)

	box := gtk.NewBox(gtk.OrientationVertical, 2)

	btnCopy := gtk.NewButton()
	btnCopy.AddCSSClass("menu-item-btn")
	lblCopy := gtk.NewLabel("Copy")
	lblCopy.AddCSSClass("menu-item-label")
	lblCopy.SetHAlign(gtk.AlignStart)
	btnCopy.SetChild(lblCopy)
	btnCopy.ConnectClicked(func() {
		tab.TermInst.Copy()
		popover.Popdown()
	})
	box.Append(btnCopy)

	btnPaste := gtk.NewButton()
	btnPaste.AddCSSClass("menu-item-btn")
	lblPaste := gtk.NewLabel("Paste")
	lblPaste.AddCSSClass("menu-item-label")
	lblPaste.SetHAlign(gtk.AlignStart)
	btnPaste.SetChild(lblPaste)
	btnPaste.ConnectClicked(func() {
		tab.TermInst.Paste()
		popover.Popdown()
	})
	box.Append(btnPaste)

	popover.SetChild(box)

	clickGesture := gtk.NewGestureClick()
	clickGesture.SetButton(0)
	clickGesture.ConnectReleased(func(nPress int, x float64, y float64) {
		button := clickGesture.CurrentButton()
		hasSel := tab.TermInst.HasSelection()

		if button == 3 { // Right click
			btnCopy.SetVisible(hasSel)
			btnPaste.SetVisible(!hasSel)

			rect := gdk.NewRectangle(int(x), int(y), 1, 1)
			popover.SetPointingTo(&rect)
			popover.Popup()
		} else if button == 1 { // Left click
			if hasSel {
				btnCopy.SetVisible(true)
				btnPaste.SetVisible(false)

				rect := gdk.NewRectangle(int(x), int(y), 1, 1)
				popover.SetPointingTo(&rect)
				popover.Popup()
			} else {
				popover.Popdown()
			}
		}
	})
	tab.TermInst.Widget.AddController(clickGesture)
}

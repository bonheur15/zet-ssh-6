package window

import (
	"strings"

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

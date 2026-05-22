package window

import (
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"zet-terminal/internal/config"
	"zet-terminal/internal/terminal"
)

func (tw *TerminalWindow) setupSideDock() {
	tw.Revealer = gtk.NewRevealer()
	tw.Revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideRight)
	tw.Revealer.SetTransitionDuration(250)
	tw.SidebarPinned = tw.Cfg.SidebarPinned
	tw.Revealer.SetRevealChild(tw.SidebarPinned)

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
	scrolled.SetMinContentWidth(220)
	scrolled.SetMaxContentWidth(220)
	panelBox.Append(scrolled)

	// Workspace Box inside scrolled
	tw.WorkspaceBox = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.WorkspaceBox.SetVExpand(true)
	tw.WorkspaceBox.SetHExpand(true)
	scrolled.SetChild(tw.WorkspaceBox)

	// Collapsible Section for Past Commands
	tw.setupHistorySection(panelBox)

	tw.Revealer.SetChild(panelBox)

	// Thin vertical hover trigger bar (widened to 10px for better accessibility)
	triggerBar := gtk.NewBox(gtk.OrientationVertical, 0)
	triggerBar.AddCSSClass("sidebar-trigger")
	triggerBar.SetSizeRequest(10, -1)
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
	dockMotionCtrl := gtk.NewEventControllerMotion()

	dockMotionCtrl.ConnectEnter(func(x float64, y float64) {
		if tw.sidebarTimeoutID != 0 {
			glib.SourceRemove(tw.sidebarTimeoutID)
			tw.sidebarTimeoutID = 0
		}
		tw.renderWorkspace()
		tw.Revealer.SetRevealChild(true)
	})

	dockMotionCtrl.ConnectMotion(func(x float64, y float64) {
		if tw.sidebarTimeoutID != 0 {
			glib.SourceRemove(tw.sidebarTimeoutID)
			tw.sidebarTimeoutID = 0
		}
	})

	dockMotionCtrl.ConnectLeave(func() {
		if tw.sidebarTimeoutID != 0 {
			glib.SourceRemove(tw.sidebarTimeoutID)
		}
		tw.sidebarTimeoutID = glib.TimeoutAdd(250, func() bool {
			tw.sidebarTimeoutID = 0
			if !tw.SidebarPinned && !tw.isCreatingGroup && tw.renamingGroupIndex == -1 && tw.renamingTabID == "" {
				tw.Revealer.SetRevealChild(false)
			}
			return false
		})
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

	historyBox := gtk.NewBox(gtk.OrientationVertical, 8)
	historyBox.AddCSSClass("history-panel")

	tw.HistorySearchEntry = gtk.NewEntry()
	tw.HistorySearchEntry.AddCSSClass("history-search-entry")
	tw.HistorySearchEntry.SetPlaceholderText("Search commands")
	tw.HistorySearchEntry.ConnectChanged(func() {
		tw.renderHistoryList(tw.HistorySearchEntry.Text())
	})
	historyBox.Append(tw.HistorySearchEntry)

	historyScrolled := gtk.NewScrolledWindow()
	historyScrolled.AddCSSClass("history-scroller")
	historyScrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	historyScrolled.SetVExpand(true)
	historyScrolled.SetMinContentHeight(180)
	historyScrolled.SetMaxContentHeight(320)

	tw.HistoryListBox = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.HistoryListBox.SetHExpand(true)
	historyScrolled.SetChild(tw.HistoryListBox)
	historyBox.Append(historyScrolled)
	tw.HistoryRevealer.SetChild(historyBox)
	parentBox.Append(tw.HistoryRevealer)

	tw.HistoryHeaderBtn.ConnectClicked(func() {
		isExpanded := tw.HistoryRevealer.RevealChild()
		tw.HistoryRevealer.SetRevealChild(!isExpanded)
		if !isExpanded {
			tw.HistoryArrowLbl.SetLabel("▼")
			tw.updateHistoryUI()
			tw.HistorySearchEntry.GrabFocus()
		} else {
			tw.HistoryArrowLbl.SetLabel("▶")
		}
	})
}

func (tw *TerminalWindow) renderWorkspace() {
	for child := tw.WorkspaceBox.FirstChild(); child != nil; child = tw.WorkspaceBox.FirstChild() {
		tw.WorkspaceBox.Remove(child)
	}

	// Create Group Button or Inline Entry
	if tw.isCreatingGroup {
		createRow := gtk.NewBox(gtk.OrientationHorizontal, 4)
		createRow.AddCSSClass("sidebar-inline-edit-row")
		createRow.SetHExpand(true)

		entry := gtk.NewEntry()
		entry.AddCSSClass("sidebar-inline-entry")
		entry.SetPlaceholderText("Group Name")
		entry.SetHExpand(true)

		glib.IdleAdd(func() {
			entry.GrabFocus()
		})

		btnSave := gtk.NewButton()
		btnSave.AddCSSClass("sidebar-inline-btn")
		btnSave.SetTooltipText("Save Group")
		imgSave := gtk.NewImageFromIconName("emblem-ok-symbolic")
		btnSave.SetChild(imgSave)

		btnCancel := gtk.NewButton()
		btnCancel.AddCSSClass("sidebar-inline-btn")
		btnCancel.SetTooltipText("Cancel")
		imgCancel := gtk.NewImageFromIconName("window-close-symbolic")
		btnCancel.SetChild(imgCancel)

		createRow.Append(entry)
		createRow.Append(btnSave)
		createRow.Append(btnCancel)

		saveFunc := func() {
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
			broadcastConfig(tw.Cfg)
			tw.isCreatingGroup = false
			tw.renderWorkspace()
		}

		cancelFunc := func() {
			tw.isCreatingGroup = false
			tw.renderWorkspace()
		}

		entry.ConnectActivate(saveFunc)

		keyCtrl := gtk.NewEventControllerKey()
		keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
			if keyval == gdk.KEY_Escape {
				cancelFunc()
				return true
			}
			return false
		})
		entry.AddController(keyCtrl)

		btnSave.ConnectClicked(saveFunc)
		btnCancel.ConnectClicked(cancelFunc)

		tw.WorkspaceBox.Append(createRow)
	} else {
		btnNewGroup := gtk.NewButton()
		btnNewGroup.AddCSSClass("workspace-action-btn")
		lblNewGroup := gtk.NewLabel("+ New Group")
		lblNewGroup.AddCSSClass("workspace-action-label")
		btnNewGroup.SetChild(lblNewGroup)
		btnNewGroup.ConnectClicked(func() {
			tw.isCreatingGroup = true
			tw.renderWorkspace()
		})
		tw.WorkspaceBox.Append(btnNewGroup)
	}

	// Render Groups and Tabs
	for gIdx, group := range tw.Cfg.TabGroups {
		groupConfig := group
		groupIndex := gIdx

		groupContainer := gtk.NewBox(gtk.OrientationVertical, 0)

		var toggleBtn *gtk.Button

		if groupIndex == tw.renamingGroupIndex {
			renameRow := gtk.NewBox(gtk.OrientationHorizontal, 4)
			renameRow.AddCSSClass("sidebar-inline-edit-row")
			renameRow.SetHExpand(true)

			entry := gtk.NewEntry()
			entry.AddCSSClass("sidebar-inline-entry")
			entry.SetText(groupConfig.Name)
			entry.SetHExpand(true)

			glib.IdleAdd(func() {
				entry.GrabFocus()
			})

			btnSave := gtk.NewButton()
			btnSave.AddCSSClass("sidebar-inline-btn")
			btnSave.SetTooltipText("Save Name")
			imgSave := gtk.NewImageFromIconName("emblem-ok-symbolic")
			btnSave.SetChild(imgSave)

			btnCancel := gtk.NewButton()
			btnCancel.AddCSSClass("sidebar-inline-btn")
			btnCancel.SetTooltipText("Cancel")
			imgCancel := gtk.NewImageFromIconName("window-close-symbolic")
			btnCancel.SetChild(imgCancel)

			renameRow.Append(entry)
			renameRow.Append(btnSave)
			renameRow.Append(btnCancel)

			saveFunc := func() {
				name := entry.Text()
				if name != "" {
					tw.Cfg.TabGroups[groupIndex].Name = name
					_ = config.SaveConfig(tw.Cfg)
					broadcastConfig(tw.Cfg)
				}
				tw.renamingGroupIndex = -1
				tw.renderWorkspace()
			}

			cancelFunc := func() {
				tw.renamingGroupIndex = -1
				tw.renderWorkspace()
			}

			entry.ConnectActivate(saveFunc)

			keyCtrl := gtk.NewEventControllerKey()
			keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
				if keyval == gdk.KEY_Escape {
					cancelFunc()
					return true
				}
				return false
			})
			entry.AddController(keyCtrl)

			btnSave.ConnectClicked(saveFunc)
			btnCancel.ConnectClicked(cancelFunc)

			groupContainer.Append(renameRow)
		} else {
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
			toggleBtn = btnToggle

			lblTitle := gtk.NewLabel(groupConfig.Name)
			lblTitle.AddCSSClass("group-title-label")
			lblTitle.SetHAlign(gtk.AlignStart)
			lblTitle.SetHExpand(true)
			lblTitle.SetEllipsize(pango.EllipsizeEnd)
			lblTitle.SetMaxWidthChars(18)
			lblTitle.SetSingleLineMode(true)
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
				tw.renamingGroupIndex = groupIndex
				tw.renderWorkspace()
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
		}

		childrenBox := gtk.NewBox(gtk.OrientationVertical, 2)
		childrenBox.SetVisible(!groupConfig.Collapsed)

		if toggleBtn != nil {
			toggleBtn.ConnectClicked(func() {
				tw.Cfg.TabGroups[groupIndex].Collapsed = !tw.Cfg.TabGroups[groupIndex].Collapsed
				_ = config.SaveConfig(tw.Cfg)
				broadcastConfig(tw.Cfg)
			})
		}

		for _, tabCfg := range groupConfig.Tabs {
			tabConfig := tabCfg

			if tabConfig.ID == tw.renamingTabID {
				renameRow := gtk.NewBox(gtk.OrientationHorizontal, 4)
				renameRow.AddCSSClass("sidebar-inline-edit-row")
				renameRow.SetHExpand(true)
				renameRow.SetMarginStart(8)

				entry := gtk.NewEntry()
				entry.AddCSSClass("sidebar-inline-entry")
				entry.SetText(tabConfig.Name)
				entry.SetHExpand(true)

				glib.IdleAdd(func() {
					entry.GrabFocus()
				})

				btnSave := gtk.NewButton()
				btnSave.AddCSSClass("sidebar-inline-btn")
				btnSave.SetTooltipText("Save Name")
				imgSave := gtk.NewImageFromIconName("emblem-ok-symbolic")
				btnSave.SetChild(imgSave)

				btnCancel := gtk.NewButton()
				btnCancel.AddCSSClass("sidebar-inline-btn")
				btnCancel.SetTooltipText("Cancel")
				imgCancel := gtk.NewImageFromIconName("window-close-symbolic")
				btnCancel.SetChild(imgCancel)

				renameRow.Append(entry)
				renameRow.Append(btnSave)
				renameRow.Append(btnCancel)

				saveFunc := func() {
					name := entry.Text()
					if name != "" {
						tw.setTabCustomNameInConfig(tabConfig.ID, name, true)
						if inst, ok := tw.TabInstances[tabConfig.ID]; ok {
							inst.Name = name
						}
						if tw.ActiveTabID == tabConfig.ID {
							tw.Win.SetTitle(name + " - Terminal")
						}
						_ = config.SaveConfig(tw.Cfg)
						broadcastConfig(tw.Cfg)
					}
					tw.renamingTabID = ""
					tw.renderWorkspace()
				}

				cancelFunc := func() {
					tw.renamingTabID = ""
					tw.renderWorkspace()
				}

				entry.ConnectActivate(saveFunc)

				keyCtrl := gtk.NewEventControllerKey()
				keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
					if keyval == gdk.KEY_Escape {
						cancelFunc()
						return true
					}
					return false
				})
				entry.AddController(keyCtrl)

				btnSave.ConnectClicked(saveFunc)
				btnCancel.ConnectClicked(cancelFunc)

				childrenBox.Append(renameRow)
			} else {
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

				displayName := tabConfig.Name
				if inst, ok := tw.TabInstances[tabConfig.ID]; ok && inst.Name != "" {
					displayName = inst.Name
				}
				if tw.isTabPinned(tabConfig.ID) {
					displayName = "📌 " + displayName
				}

				lblTab := gtk.NewLabel(displayName)
				lblTab.AddCSSClass("tab-label")
				lblTab.SetHAlign(gtk.AlignStart)
				lblTab.SetXAlign(0.0)
				lblTab.SetEllipsize(pango.EllipsizeEnd)
				lblTab.SetMaxWidthChars(20)
				lblTab.SetSingleLineMode(true)
				btnSelectTab.SetChild(lblTab)

				btnSelectTab.ConnectClicked(func() {
					tw.ActivateTab(tabConfig.ID)
				})
				tabRow.Append(btnSelectTab)

				// Pin tab button
				btnPinTab := gtk.NewButton()
				btnPinTab.AddCSSClass("tab-action-btn")
				isPinned := tw.isTabPinned(tabConfig.ID)
				if isPinned {
					btnPinTab.SetTooltipText("Unpin tab")
					imgPinTab := gtk.NewImageFromIconName("bookmark-symbolic")
					btnPinTab.SetChild(imgPinTab)
				} else {
					btnPinTab.SetTooltipText("Pin tab")
					imgPinPin := gtk.NewImageFromIconName("bookmark-new-symbolic")
					btnPinTab.SetChild(imgPinPin)
				}
				btnPinTab.ConnectClicked(func() {
					tw.toggleTabPinned(tabConfig.ID)
				})
				tabRow.Append(btnPinTab)

				// Rename tab button
				btnRenameTab := gtk.NewButton()
				btnRenameTab.AddCSSClass("tab-action-btn")
				btnRenameTab.SetTooltipText("Rename tab")
				imgRenameTab := gtk.NewImageFromIconName("document-edit-symbolic")
				btnRenameTab.SetChild(imgRenameTab)
				btnRenameTab.ConnectClicked(func() {
					tw.renamingTabID = tabConfig.ID
					tw.renderWorkspace()
				})
				tabRow.Append(btnRenameTab)

				// Fully working, comfortably padded Close button using symbolic cross icon
				btnCloseTab := gtk.NewButton()
				btnCloseTab.AddCSSClass("tab-action-btn")
				btnCloseTab.AddCSSClass("tab-close-btn") // Differentiate for red hover
				btnCloseTab.SetTooltipText("Close tab")
				imgCloseTab := gtk.NewImageFromIconName("window-close-symbolic")
				btnCloseTab.SetChild(imgCloseTab)
				btnCloseTab.ConnectClicked(func() {
					tw.CloseTab(tabConfig.ID)
				})
				tabRow.Append(btnCloseTab)

				childrenBox.Append(tabRow)
			}
		}

		groupContainer.Append(childrenBox)
		tw.WorkspaceBox.Append(groupContainer)
	}
}

func (tw *TerminalWindow) createTabInGroup(groupID string) {
	dir := tw.getActiveTabDir()
	tw.CreateNewTabInGroup(groupID, dir, tw)
}

func (tw *TerminalWindow) deleteGroup(groupIndex int) {
	group := tw.Cfg.TabGroups[groupIndex]

	tw.Cfg.TabGroups = append(tw.Cfg.TabGroups[:groupIndex], tw.Cfg.TabGroups[groupIndex+1:]...)
	if len(tw.Cfg.TabGroups) == 0 {
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
	}
	_ = config.SaveConfig(tw.Cfg)

	for w := range activeWindows {
		for _, tabCfg := range group.Tabs {
			w.closeTabSilently(tabCfg.ID, false)
		}
	}
	broadcastConfig(tw.Cfg)
	if len(group.Tabs) > 0 && tw.TabInstances["tab-1"] == nil && len(tw.Cfg.TabGroups) == 1 && len(tw.Cfg.TabGroups[0].Tabs) == 1 && tw.Cfg.TabGroups[0].Tabs[0].ID == "tab-1" {
		for w := range activeWindows {
			if len(w.TabInstances) == 0 {
				w.CreateTab("tab-1", "Primary Console", "group-general", "", false)
				w.ActivateTab("tab-1")
			}
		}
	}
}

func (tw *TerminalWindow) updateHistoryUI() {
	tw.HistoryCommands = terminal.ReadShellHistory(tw.Cfg.Shell, tw.Cfg.CommandHistoryLimit)
	tw.renderHistoryList("")
	if tw.HistorySearchEntry != nil {
		tw.HistorySearchEntry.SetText("")
	}
}

func (tw *TerminalWindow) renderHistoryList(query string) {
	for child := tw.HistoryListBox.FirstChild(); child != nil; child = tw.HistoryListBox.FirstChild() {
		tw.HistoryListBox.Remove(child)
	}

	query = strings.TrimSpace(strings.ToLower(query))
	history := tw.HistoryCommands
	if len(history) == 0 {
		emptyLabel := gtk.NewLabel("No commands run yet.")
		emptyLabel.AddCSSClass("sidebar-btn-label")
		emptyLabel.SetMarginTop(20)
		emptyLabel.SetHAlign(gtk.AlignCenter)
		tw.HistoryListBox.Append(emptyLabel)
		return
	}

	matchCount := 0
	for _, cmd := range history {
		cmdStr := cmd
		if query != "" && !strings.Contains(strings.ToLower(cmdStr), query) {
			continue
		}
		matchCount++

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
		lbl.SetEllipsize(pango.EllipsizeEnd)
		lbl.SetMaxWidthChars(24)
		lbl.SetSingleLineMode(true)
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

	if matchCount == 0 {
		emptyLabel := gtk.NewLabel("No matching commands.")
		emptyLabel.AddCSSClass("sidebar-btn-label")
		emptyLabel.SetMarginTop(10)
		emptyLabel.SetHAlign(gtk.AlignCenter)
		tw.HistoryListBox.Append(emptyLabel)
	}
}

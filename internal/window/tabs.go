package window

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

func (tw *TerminalWindow) applyConfigToInstance(inst *terminal.VteTerminalInstance) {
	inst.SetFont(tw.Cfg.FontName, tw.Cfg.FontSize)
	pal := theme.GetTerminalPalette(tw.Cfg)
	inst.SetColors(pal.Foreground, pal.Background, pal.Palette)
}

func (tw *TerminalWindow) CreateTab(id, name, groupID, workingDir string, saveToConfig bool) *TabInstance {
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

	// Spawn shell in directory
	inst.SpawnShell(tw.Cfg.Shell, workingDir)

	// Signal Handlers
	inst.OnChildExited(func(status int) {
		inst.Destroy()
		tw.CloseTab(id)
	})

	inst.OnWindowTitleChanged(func(title string) {
		if tw.isTabCustomNamed(id) {
			if tw.ActiveTabID == id {
				tw.Win.SetTitle(title + " - Terminal")
			}
			return
		}

		autoTitle := tw.getTabAutoTitle(id, title)
		tab.Name = autoTitle

		if tw.ActiveTabID == id {
			tw.Win.SetTitle(autoTitle + " - Terminal")
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
	_, ok := tw.TabInstances[id]
	if !ok {
		return
	}

	// Remove from config first
	tw.removeTabFromConfig(id)

	// Close on all windows
	for w := range activeWindows {
		w.Cfg = tw.Cfg
		w.CloseTabSilently(id)
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

	tw.CreateTab("tab-1", "Primary Console", "group-general", "", false)
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

func expandPath(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func getGitBranch(dir string) string {
	current := dir
	for {
		headPath := filepath.Join(current, ".git", "HEAD")
		data, err := os.ReadFile(headPath)
		if err != nil {
			dotGitPath := filepath.Join(current, ".git")
			st, err := os.Stat(dotGitPath)
			if err == nil && !st.IsDir() {
				gitDirData, err := os.ReadFile(dotGitPath)
				if err == nil {
					content := strings.TrimSpace(string(gitDirData))
					if strings.HasPrefix(content, "gitdir: ") {
						realGitDir := strings.TrimPrefix(content, "gitdir: ")
						if !filepath.IsAbs(realGitDir) {
							realGitDir = filepath.Clean(filepath.Join(current, realGitDir))
						}
						headPath = filepath.Join(realGitDir, "HEAD")
						data, err = os.ReadFile(headPath)
					}
				}
			}
		}
		if err == nil {
			content := strings.TrimSpace(string(data))
			if strings.HasPrefix(content, "ref: refs/heads/") {
				return strings.TrimPrefix(content, "ref: refs/heads/")
			}
			if len(content) > 7 {
				return content[:7]
			}
			return content
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

func getForegroundCommand(shellPID int) string {
	if shellPID <= 0 {
		return ""
	}
	statPath := fmt.Sprintf("/proc/%d/stat", shellPID)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return ""
	}
	content := string(data)
	lastParen := strings.LastIndex(content, ")")
	if lastParen == -1 || lastParen+2 >= len(content) {
		return ""
	}
	fields := strings.Fields(content[lastParen+2:])
	if len(fields) < 6 {
		return ""
	}
	pgrp, err := strconv.Atoi(fields[2])
	if err != nil {
		return ""
	}
	tpgid, err := strconv.Atoi(fields[5])
	if err != nil {
		return ""
	}

	if tpgid <= 0 || tpgid == pgrp {
		return ""
	}

	// Try reading tpgid cmdline first
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", tpgid)
	cmdData, err := os.ReadFile(cmdlinePath)
	if err == nil && len(cmdData) > 0 {
		cmdStr := string(cmdData)
		if idx := strings.IndexByte(cmdStr, 0); idx != -1 {
			cmdStr = cmdStr[:idx]
		}
		parts := strings.Fields(cmdStr)
		if len(parts) > 0 {
			return filepath.Base(parts[0])
		}
		return filepath.Base(cmdStr)
	}

	// If the group leader is dead, find any child of shellPID that belongs to the tpgid group
	childrenPath := fmt.Sprintf("/proc/%d/task/%d/children", shellPID, shellPID)
	childrenData, err := os.ReadFile(childrenPath)
	if err == nil {
		childrenFields := strings.Fields(string(childrenData))
		for _, childStr := range childrenFields {
			childPID, err := strconv.Atoi(childStr)
			if err != nil {
				continue
			}
			childStatPath := fmt.Sprintf("/proc/%d/stat", childPID)
			cStatData, err := os.ReadFile(childStatPath)
			if err == nil {
				cContent := string(cStatData)
				cLastParen := strings.LastIndex(cContent, ")")
				if cLastParen != -1 && cLastParen+2 < len(cContent) {
					cFields := strings.Fields(cContent[cLastParen+2:])
					if len(cFields) >= 3 {
						cPgrp, _ := strconv.Atoi(cFields[2])
						if cPgrp == tpgid {
							cCmdlinePath := fmt.Sprintf("/proc/%d/cmdline", childPID)
							cCmdData, err := os.ReadFile(cCmdlinePath)
							if err == nil && len(cCmdData) > 0 {
								cCmdStr := string(cCmdData)
								if idx := strings.IndexByte(cCmdStr, 0); idx != -1 {
									cCmdStr = cCmdStr[:idx]
								}
								parts := strings.Fields(cCmdStr)
								if len(parts) > 0 {
									return filepath.Base(parts[0])
								}
								return filepath.Base(cCmdStr)
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func formatMinimalPath(path string) string {
	path = expandPath(path)
	home, err := os.UserHomeDir()
	if err == nil {
		if path == home {
			return "~"
		}
		if strings.HasPrefix(path, home+"/") {
			path = "~/" + strings.TrimPrefix(path, home+"/")
		}
	}

	parts := strings.Split(path, "/")
	var cleanParts []string
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}

	if len(cleanParts) == 0 {
		return "/"
	}

	if len(cleanParts) >= 2 {
		return cleanParts[len(cleanParts)-2] + "/" + cleanParts[len(cleanParts)-1]
	}
	return cleanParts[len(cleanParts)-1]
}

func (tw *TerminalWindow) getTabAutoTitle(tabID string, rawTitle string) string {
	rawTitle = strings.TrimSpace(rawTitle)

	// 1. Try to get the active foreground command running in the terminal
	if inst, ok := tw.TabInstances[tabID]; ok {
		shellPID := inst.TermInst.GetChildPID()
		if shellPID > 0 {
			if fgCmd := getForegroundCommand(shellPID); fgCmd != "" {
				// Ignore basic shells so we can display directory / branch details
				if fgCmd != "bash" && fgCmd != "zsh" && fgCmd != "fish" && fgCmd != "sh" {
					return fgCmd
				}
			}
		}
	}

	// 2. Handle remote sessions or standard prompt title
	var dirPath string
	isRemote := false
	if strings.Contains(rawTitle, "@") && strings.Contains(rawTitle, ":") {
		parts := strings.SplitN(rawTitle, ":", 2)
		if len(parts) == 2 {
			userHost := parts[0]
			dirPath = strings.TrimSpace(parts[1])

			hostname, _ := os.Hostname()
			currentUser := os.Getenv("USER")

			isLocal := false
			if currentUser != "" && hostname != "" {
				expectedLocal := currentUser + "@" + hostname
				if userHost == expectedLocal || strings.HasPrefix(userHost, currentUser+"@localhost") {
					isLocal = true
				}
			} else {
				isLocal = true
			}

			if !isLocal {
				isRemote = true
				hostParts := strings.Split(userHost, "@")
				host := userHost
				if len(hostParts) == 2 {
					host = hostParts[1]
				}
				shortDir := dirPath
				if !strings.HasPrefix(dirPath, "~/") {
					dirParts := strings.Split(dirPath, "/")
					if len(dirParts) > 0 {
						shortDir = dirParts[len(dirParts)-1]
					}
				}
				return fmt.Sprintf("%s:%s", host, shortDir)
			}
		}
	}

	// 3. If the user or app set a custom window title using OSC 2 (e.g. "Gemini Chat", "ssh node1"), use it!
	if rawTitle != "" && !isRemote {
		isDefaultShellPrompt := strings.Contains(rawTitle, "@") && strings.Contains(rawTitle, ":")
		isShellName := rawTitle == "bash" || rawTitle == "zsh" || rawTitle == "fish" || rawTitle == "sh" || rawTitle == "tmux"
		if !isDefaultShellPrompt && !isShellName {
			return rawTitle
		}
	}

	// 4. Local directory path fallback
	if dirPath == "" {
		if inst, ok := tw.TabInstances[tabID]; ok {
			dirPath = inst.TermInst.GetCurrentDirectory()
		}
	}

	if dirPath != "" {
		expanded := expandPath(dirPath)
		shortPath := formatMinimalPath(expanded)
		branch := getGitBranch(expanded)
		if branch != "" {
			return fmt.Sprintf("%s  %s", shortPath, branch)
		}
		return shortPath
	}

	if rawTitle != "" {
		return rawTitle
	}

	return "Console"
}

func (tw *TerminalWindow) isTabCustomNamed(tabID string) bool {
	for _, group := range tw.Cfg.TabGroups {
		for _, tab := range group.Tabs {
			if tab.ID == tabID {
				return tab.CustomName
			}
		}
	}
	return false
}

func (tw *TerminalWindow) setTabCustomNameInConfig(tabID string, name string, custom bool) {
	for gIdx, group := range tw.Cfg.TabGroups {
		for tIdx, tab := range group.Tabs {
			if tab.ID == tabID {
				tw.Cfg.TabGroups[gIdx].Tabs[tIdx].Name = name
				tw.Cfg.TabGroups[gIdx].Tabs[tIdx].CustomName = custom
				return
			}
		}
	}
}

func (tw *TerminalWindow) updateTabAutoTitles() {
	needRender := false
	for id, tab := range tw.TabInstances {
		if tw.isTabCustomNamed(id) {
			continue
		}

		rawTitle := tab.TermInst.GetWindowTitle()
		autoTitle := tw.getTabAutoTitle(id, rawTitle)
		if tab.Name != autoTitle {
			tab.Name = autoTitle
			if tw.ActiveTabID == id {
				tw.Win.SetTitle(autoTitle + " - Terminal")
			}
			needRender = true
		}
	}
	if needRender {
		tw.renderWorkspace()
	}
}

func (tw *TerminalWindow) getActiveTabDir() string {
	if tw.ActiveTabID == "" {
		return ""
	}
	activeTab, ok := tw.TabInstances[tw.ActiveTabID]
	if !ok {
		return ""
	}
	return activeTab.TermInst.GetCurrentDirectory()
}

func (tw *TerminalWindow) GetActiveTabGroupID() string {
	if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
		return activeTab.GroupID
	}
	if len(tw.Cfg.TabGroups) > 0 {
		return tw.Cfg.TabGroups[0].ID
	}
	return ""
}

func (tw *TerminalWindow) CreateNewTabInGroup(groupID string, workingDir string, activateOnWindow *TerminalWindow) string {
	tabID := fmt.Sprintf("tab-%d", time.Now().UnixNano())
	tabName := "Console"

	// Add to configuration
	for i, group := range tw.Cfg.TabGroups {
		if group.ID == groupID {
			tw.Cfg.TabGroups[i].Tabs = append(tw.Cfg.TabGroups[i].Tabs, config.TabConfig{
				ID:   tabID,
				Name: tabName,
			})
			_ = config.SaveConfig(tw.Cfg)
			break
		}
	}

	// Create tab on all active windows
	for w := range activeWindows {
		w.Cfg = tw.Cfg
		w.CreateTab(tabID, tabName, groupID, workingDir, false)
		if w == activateOnWindow {
			w.ActivateTab(tabID)
		} else {
			w.renderWorkspace()
		}
	}

	return tabID
}

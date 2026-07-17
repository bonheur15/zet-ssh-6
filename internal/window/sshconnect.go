package window

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/gotk4/pkg/core/glib"

	"zet-terminal/internal/cmdlog"
	"zet-terminal/internal/config"
	"zet-terminal/internal/sshmgr"
	"zet-terminal/internal/vault"
)

func cfgDir() string { return config.GetConfigDir() }

func profileStore() *sshmgr.ProfileStore {
	return sshmgr.Profiles(config.GetConfigDir())
}

var tunnelWatchOnce sync.Once

func tunnelStore() *sshmgr.TunnelStore {
	store := sshmgr.Tunnels(config.GetConfigDir())
	// Re-render open sidebars when a tunnel starts/stops/dies. The
	// callback fires from a background goroutine, so hop to the UI loop.
	tunnelWatchOnce.Do(func() {
		store.OnChange(func() {
			glib.IdleAdd(func() {
				for w := range activeWindows {
					if w.sidebarSection == "tunnels" {
						w.renderWorkspace()
					}
				}
			})
		})
	})
	return store
}

// profileIDForTab returns the SSH profile bound to a tab, if any.
func (tw *TerminalWindow) profileIDForTab(tabID string) string {
	for _, group := range tw.Cfg.TabGroups {
		for _, tab := range group.Tabs {
			if tab.ID == tabID {
				return tab.ProfileID
			}
		}
	}
	return ""
}

// CreateCommandTab opens a new tab running argv instead of a shell.
// profileID (optional) binds the tab to a profile so it reconnects on
// restart; label is used as the custom tab name.
func (tw *TerminalWindow) CreateCommandTab(label string, argv []string, profileID string) *TabInstance {
	tabID := fmt.Sprintf("tab-%d", time.Now().UnixNano())
	groupID := tw.GetActiveTabGroupID()

	for i, group := range tw.Cfg.TabGroups {
		if group.ID == groupID {
			tw.Cfg.TabGroups[i].Tabs = append(tw.Cfg.TabGroups[i].Tabs, config.TabConfig{
				ID:         tabID,
				Name:       label,
				CustomName: true,
				ProfileID:  profileID,
			})
			_ = config.SaveConfig(tw.Cfg)
			break
		}
	}

	tab := tw.createTabCore(tabID, label, groupID, "", argv)
	tw.ActivateTab(tabID)
	return tab
}

// ConnectProfile opens an SSH session for the profile in a new tab,
// logging the exact generated command. Password-auth profiles unlock
// the vault first and auto-fill the password prompt.
func (tw *TerminalWindow) ConnectProfile(p *sshmgr.Profile) {
	if p == nil {
		return
	}
	if p.Auth == sshmgr.AuthPassword {
		tw.WithVault(func(v *vault.Vault) {
			tw.connectProfileNow(p)
		})
		return
	}
	tw.connectProfileNow(p)
}

func (tw *TerminalWindow) connectProfileNow(p *sshmgr.Profile) {
	profileStore().MarkUsed(p.ID)
	argv := p.SSHArgv()
	cmdlog.Add("Connect "+p.Label(), sshmgr.CommandString(argv))

	tab := tw.CreateCommandTab(p.Label(), argv, p.ID)
	tw.maybeAutofillPassword(tab, p)
}

// OpenSFTPTab opens an interactive sftp session for the profile.
func (tw *TerminalWindow) OpenSFTPTab(p *sshmgr.Profile) {
	argv := p.SFTPArgv()
	cmdlog.Add("SFTP "+p.Label(), sshmgr.CommandString(argv))
	tab := tw.CreateCommandTab("sftp "+p.Label(), argv, "")
	tw.maybeAutofillPassword(tab, p)
}

// RunTransfer starts an scp transfer in a new tab so progress is visible.
func (tw *TerminalWindow) RunTransfer(p *sshmgr.Profile, localPath, remotePath string, upload, recursive bool) {
	argv := p.SCPArgv(localPath, remotePath, upload, recursive)
	direction := "Download"
	if upload {
		direction = "Upload"
	}
	cmdlog.Add(direction+" "+p.Label(), sshmgr.CommandString(argv))
	tab := tw.CreateCommandTab("scp "+p.Label(), argv, "")
	tw.maybeAutofillPassword(tab, p)
}

// maybeAutofillPassword watches the new session for a password prompt
// and feeds the vault-stored password once, if available.
func (tw *TerminalWindow) maybeAutofillPassword(tab *TabInstance, p *sshmgr.Profile) {
	if tab == nil || p == nil || p.Auth != sshmgr.AuthPassword {
		return
	}
	v := appVault()
	if !v.IsUnlocked() {
		return
	}
	password, err := v.Get(vault.ProfilePasswordKey(p.ID))
	if err != nil || password == "" {
		return
	}

	ticks := 0
	fed := false
	glib.TimeoutAdd(400, func() bool {
		ticks++
		if fed || ticks > 150 { // give up after ~60s
			return false
		}
		if _, alive := tw.TabInstances[tab.ID]; !alive {
			return false
		}
		text := strings.TrimRight(tab.TermInst.GetText(), " \t\r\n")
		if text == "" {
			return true
		}
		lastLine := text
		if idx := strings.LastIndex(text, "\n"); idx != -1 {
			lastLine = text[idx+1:]
		}
		lastLine = strings.TrimSpace(strings.ToLower(lastLine))
		if strings.Contains(lastLine, "password") && strings.HasSuffix(lastLine, ":") {
			tab.TermInst.FeedChild(password + "\n")
			fed = true
			return false
		}
		return true
	})
}

// profilePassword returns the vault-stored password for a password-auth
// profile, or "" if not applicable/available. It does not prompt.
func profilePassword(p *sshmgr.Profile) string {
	if p == nil || p.Auth != sshmgr.AuthPassword {
		return ""
	}
	v := appVault()
	if !v.IsUnlocked() {
		return ""
	}
	if pw, err := v.Get(vault.ProfilePasswordKey(p.ID)); err == nil {
		return pw
	}
	return ""
}

// InsertIntoActiveTab types text into the focused terminal (no newline).
func (tw *TerminalWindow) InsertIntoActiveTab(text string) {
	if activeTab, ok := tw.TabInstances[tw.ActiveTabID]; ok {
		activeTab.TermInst.FeedChild(text)
		activeTab.TermInst.Widget.GrabFocus()
	}
}

// RunInActiveTab types text into the focused terminal and executes it.
func (tw *TerminalWindow) RunInActiveTab(text string) {
	tw.InsertIntoActiveTab(text + "\n")
}

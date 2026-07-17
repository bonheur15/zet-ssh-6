package window

import (
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/sshmgr"
	"zet-terminal/internal/vault"
)

var authOptions = []string{"SSH Agent", "Private Key", "Password"}
var knownHostsOptions = []string{"Prompt (default)", "Strict", "Accept new"}

func authIndex(a sshmgr.AuthMethod) int {
	switch a {
	case sshmgr.AuthKey:
		return 1
	case sshmgr.AuthPassword:
		return 2
	default:
		return 0
	}
}

func authFromIndex(i int) sshmgr.AuthMethod {
	switch i {
	case 1:
		return sshmgr.AuthKey
	case 2:
		return sshmgr.AuthPassword
	default:
		return sshmgr.AuthAgent
	}
}

func knownHostsIndex(k sshmgr.KnownHostsMode) int {
	switch k {
	case sshmgr.KnownHostsStrict:
		return 1
	case sshmgr.KnownHostsAccept:
		return 2
	default:
		return 0
	}
}

func knownHostsFromIndex(i int) sshmgr.KnownHostsMode {
	switch i {
	case 1:
		return sshmgr.KnownHostsStrict
	case 2:
		return sshmgr.KnownHostsAccept
	default:
		return sshmgr.KnownHostsPrompt
	}
}

// openProfileDialog edits an existing profile or creates a new one when
// p is nil. seed pre-fills fields (used by "Save as Profile" from a
// pasted ssh command).
func (tw *TerminalWindow) openProfileDialog(p *sshmgr.Profile, seed *sshmgr.Profile, onSaved func()) {
	editing := p != nil
	if p == nil {
		if seed != nil {
			p = seed
		} else {
			p = &sshmgr.Profile{Port: 22, Auth: sshmgr.AuthAgent, KnownHosts: sshmgr.KnownHostsPrompt}
		}
	}

	titleText := "New SSH Profile"
	if editing {
		titleText = "Edit SSH Profile"
	}
	dialog, box := tw.newDialog(titleText, 560, 720)
	dialogTitle(box, titleText, "Save a connection once and reuse it. Every field maps to a real ssh flag shown in the live command below.")

	content := scrollArea(box)

	entryName := labeledEntry(content, "Name:", p.Name, "prod-api")
	entryHost := labeledEntry(content, "Host / IP:", p.Host, "example.com")
	entryPort := labeledEntry(content, "Port:", strconv.Itoa(portOrDefault(p.Port)), "22")
	entryUser := labeledEntry(content, "Username:", p.User, "root")

	comboAuth := labeledCombo(content, "Auth Method:", authOptions, authIndex(p.Auth))

	// Key path (shown for key auth)
	keyBox := gtk.NewBox(gtk.OrientationVertical, 6)
	entryKey := labeledEntry(keyBox, "Private Key Path:", p.KeyPath, "~/.ssh/id_ed25519")
	content.Append(keyBox)

	// Password (shown for password auth) — stored in the vault, never here.
	pwBox := gtk.NewBox(gtk.OrientationVertical, 6)
	pwRow := gtk.NewBox(gtk.OrientationHorizontal, 8)
	pwRow.AddCSSClass("settings-row")
	pwLbl := gtk.NewLabel("Password:")
	pwLbl.AddCSSClass("settings-label")
	pwLbl.SetHAlign(gtk.AlignStart)
	pwLbl.SetHExpand(true)
	pwRow.Append(pwLbl)
	pwEntry := gtk.NewPasswordEntry()
	pwEntry.SetShowPeekIcon(true)
	pwEntry.AddCSSClass("settings-entry")
	pwRow.Append(pwEntry)
	pwBox.Append(pwRow)
	pwHint := gtk.NewLabel("Stored encrypted in the vault. Leave blank to keep the existing password.")
	pwHint.AddCSSClass("settings-hint")
	pwHint.SetHAlign(gtk.AlignStart)
	pwHint.SetWrap(true)
	pwBox.Append(pwHint)
	content.Append(pwBox)

	// Advanced
	adv := gtk.NewLabel("Advanced")
	adv.AddCSSClass("settings-section-title")
	adv.SetHAlign(gtk.AlignStart)
	adv.SetMarginTop(6)
	content.Append(adv)

	entryJump := labeledEntry(content, "Jump Host (ProxyJump):", p.JumpHost, "user@bastion")
	comboKnown := labeledCombo(content, "Host Key Checking:", knownHostsOptions, knownHostsIndex(p.KnownHosts))
	spinKeepalive := labeledSpin(content, "Keepalive (sec, 0=off):", float64(p.Keepalive), 0, 3600, 5)
	entryRemoteDir := labeledEntry(content, "Default Remote Dir:", p.RemoteDir, "/var/www")
	entryCiphers := labeledEntry(content, "Ciphers (advanced):", p.Ciphers, "")
	entryKex := labeledEntry(content, "KEX Algorithms (advanced):", p.Kex, "")
	entryExtra := labeledEntry(content, "Extra ssh args:", p.ExtraArgs, "-C -4")
	entryTags := labeledEntry(content, "Tags (comma sep):", strings.Join(p.Tags, ", "), "prod, api")

	// Live command preview
	previewLbl, previewCopy := commandPreview(content, "")

	// Build a scratch profile from the current form for preview/save.
	collect := func() *sshmgr.Profile {
		port, _ := strconv.Atoi(strings.TrimSpace(entryPort.Text()))
		if port == 0 {
			port = 22
		}
		np := &sshmgr.Profile{
			ID:         p.ID,
			Name:       strings.TrimSpace(entryName.Text()),
			Host:       strings.TrimSpace(entryHost.Text()),
			Port:       port,
			User:       strings.TrimSpace(entryUser.Text()),
			Auth:       authFromIndex(comboAuth.Active()),
			KeyPath:    strings.TrimSpace(entryKey.Text()),
			JumpHost:   strings.TrimSpace(entryJump.Text()),
			KnownHosts: knownHostsFromIndex(comboKnown.Active()),
			Keepalive:  int(spinKeepalive.Value()),
			RemoteDir:  strings.TrimSpace(entryRemoteDir.Text()),
			Ciphers:    strings.TrimSpace(entryCiphers.Text()),
			Kex:        strings.TrimSpace(entryKex.Text()),
			ExtraArgs:  strings.TrimSpace(entryExtra.Text()),
			LastUsed:   p.LastUsed,
		}
		for _, t := range strings.Split(entryTags.Text(), ",") {
			if t = strings.TrimSpace(t); t != "" {
				np.Tags = append(np.Tags, t)
			}
		}
		return np
	}

	refresh := func() {
		np := collect()
		if np.Host == "" {
			previewLbl.SetLabel("(enter a host to preview the command)")
		} else {
			previewLbl.SetLabel(sshmgr.CommandString(np.SSHArgv()))
		}
		idx := comboAuth.Active()
		keyBox.SetVisible(idx == 1)
		pwBox.SetVisible(idx == 2)
	}

	entryHost.ConnectChanged(refresh)
	entryPort.ConnectChanged(refresh)
	entryUser.ConnectChanged(refresh)
	entryKey.ConnectChanged(refresh)
	entryJump.ConnectChanged(refresh)
	entryRemoteDir.ConnectChanged(refresh)
	entryCiphers.ConnectChanged(refresh)
	entryKex.ConnectChanged(refresh)
	entryExtra.ConnectChanged(refresh)
	comboAuth.Connect("changed", refresh)
	comboKnown.Connect("changed", refresh)
	spinKeepalive.Connect("value-changed", refresh)
	refresh()

	previewCopy.ConnectClicked(func() {
		copyToClipboard(previewLbl.Text())
	})

	save := func() {
		np := collect()
		if np.Host == "" {
			entryHost.GrabFocus()
			return
		}
		store := profileStore()
		store.Upsert(np)

		// Persist the password to the vault if one was typed.
		if np.Auth == sshmgr.AuthPassword && pwEntry.Text() != "" {
			pw := pwEntry.Text()
			tw.WithVault(func(v *vault.Vault) {
				_ = v.Set(vault.ProfilePasswordKey(np.ID), pw)
			})
		}
		dialog.Close()
		if onSaved != nil {
			onSaved()
		}
		tw.renderWorkspace()
	}
	dialogButtons(box, dialog, "Save Profile", save)

	dialog.Present()
}

func portOrDefault(p int) int {
	if p == 0 {
		return 22
	}
	return p
}

// offerSavePastedCommand detects a pasted ssh command and offers to save
// it as a profile. Returns true if it handled the text.
func (tw *TerminalWindow) offerSaveAsProfile(cmdline string) bool {
	seed, ok := sshmgr.ParseSSHCommand(cmdline)
	if !ok {
		return false
	}
	tw.openProfileDialog(nil, seed, nil)
	return true
}

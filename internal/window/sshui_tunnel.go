package window

import (
	"os/exec"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/cmdlog"
	"zet-terminal/internal/sshmgr"
	"zet-terminal/internal/vault"
)

var tunnelTypeOptions = []string{"Local (-L)", "Remote (-R)", "Dynamic SOCKS (-D)"}

func tunnelTypeIndex(t sshmgr.TunnelType) int {
	switch t {
	case sshmgr.TunnelRemote:
		return 1
	case sshmgr.TunnelDynamic:
		return 2
	default:
		return 0
	}
}

func tunnelTypeFromIndex(i int) sshmgr.TunnelType {
	switch i {
	case 1:
		return sshmgr.TunnelRemote
	case 2:
		return sshmgr.TunnelDynamic
	default:
		return sshmgr.TunnelLocal
	}
}

// openTunnelDialog edits t, or creates a new tunnel when t is nil.
func (tw *TerminalWindow) openTunnelDialog(t *sshmgr.TunnelSpec, onSaved func()) {
	editing := t != nil
	if t == nil {
		t = &sshmgr.TunnelSpec{Type: sshmgr.TunnelLocal, BindPort: 8080, TargetHost: "localhost", TargetPort: 80}
	}

	profiles := profileStore().Sorted()
	if len(profiles) == 0 {
		tw.notify("No profiles yet", "Create an SSH profile first, then attach a tunnel to it.")
		return
	}

	titleText := "New Tunnel"
	if editing {
		titleText = "Edit Tunnel"
	}
	dialog, box := tw.newDialog(titleText, 540, 560)
	dialogTitle(box, titleText, "Forward ports through an SSH connection. The exact ssh -N command is shown below and logged when you start it.")

	content := scrollArea(box)

	entryName := labeledEntry(content, "Name:", t.Name, "db forward")

	// Profile picker
	var profLabels []string
	activeProf := 0
	for i, p := range profiles {
		profLabels = append(profLabels, p.Label())
		if p.ID == t.ProfileID {
			activeProf = i
		}
	}
	comboProfile := labeledCombo(content, "Via Profile:", profLabels, activeProf)

	comboType := labeledCombo(content, "Type:", tunnelTypeOptions, tunnelTypeIndex(t.Type))
	entryBindHost := labeledEntry(content, "Bind Host:", t.BindHost, "localhost")
	spinBindPort := labeledSpin(content, "Bind Port:", float64(t.BindPort), 1, 65535, 1)

	targetBox := gtk.NewBox(gtk.OrientationVertical, 6)
	entryTargetHost := labeledEntry(targetBox, "Target Host:", t.TargetHost, "localhost")
	spinTargetPort := labeledSpin(targetBox, "Target Port:", float64(defPort(t.TargetPort, 80)), 1, 65535, 1)
	content.Append(targetBox)

	previewLbl, previewCopy := commandPreview(content, "")

	collect := func() (*sshmgr.TunnelSpec, *sshmgr.Profile) {
		p := profiles[comboProfile.Active()]
		nt := &sshmgr.TunnelSpec{
			ID:         t.ID,
			Name:       strings.TrimSpace(entryName.Text()),
			ProfileID:  p.ID,
			Type:       tunnelTypeFromIndex(comboType.Active()),
			BindHost:   strings.TrimSpace(entryBindHost.Text()),
			BindPort:   int(spinBindPort.Value()),
			TargetHost: strings.TrimSpace(entryTargetHost.Text()),
			TargetPort: int(spinTargetPort.Value()),
		}
		return nt, p
	}

	refresh := func() {
		nt, p := collect()
		previewLbl.SetLabel(sshmgr.CommandString(nt.Argv(p)))
		targetBox.SetVisible(comboType.Active() != 2) // dynamic has no target
	}
	comboType.Connect("changed", refresh)
	comboProfile.Connect("changed", refresh)
	entryBindHost.ConnectChanged(refresh)
	entryTargetHost.ConnectChanged(refresh)
	spinBindPort.Connect("value-changed", refresh)
	spinTargetPort.Connect("value-changed", refresh)
	refresh()

	previewCopy.ConnectClicked(func() { copyToClipboard(previewLbl.Text()) })

	dialogButtons(box, dialog, "Save Tunnel", func() {
		nt, _ := collect()
		tunnelStore().Upsert(nt)
		dialog.Close()
		if onSaved != nil {
			onSaved()
		}
		tw.renderWorkspace()
	})

	dialog.Present()
}

// StartTunnel launches a tunnel and logs the command. Password-auth
// profiles are unlocked from the vault first so the background ssh can
// authenticate via sshpass.
func (tw *TerminalWindow) StartTunnel(t *sshmgr.TunnelSpec) {
	p := profileStore().Get(t.ProfileID)
	if p == nil {
		tw.notify("Missing profile", "This tunnel's profile no longer exists.")
		return
	}

	if p.Auth == sshmgr.AuthPassword {
		if _, err := exec.LookPath("sshpass"); err != nil {
			tw.notify("sshpass required", "Password-authenticated tunnels need the 'sshpass' tool. Install it, or use SSH agent / key auth for this host.")
			return
		}
		if !appVault().IsUnlocked() || profilePassword(p) == "" {
			tw.WithVault(func(v *vault.Vault) {
				if profilePassword(p) == "" {
					tw.notify("No saved password", "This host has no password stored in the vault. Edit the host to save one, then start the tunnel again.")
					return
				}
				tw.startTunnelNow(t, p)
			})
			return
		}
	}
	tw.startTunnelNow(t, p)
}

func (tw *TerminalWindow) startTunnelNow(t *sshmgr.TunnelSpec, p *sshmgr.Profile) {
	argv, err := tunnelStore().Start(t, p, profilePassword(p))
	if argv != nil {
		cmdlog.Add("Start tunnel "+tunnelLabel(t), sshmgr.CommandString(argv))
	}
	if err != nil {
		tw.notify("Tunnel failed", err.Error())
	}
	tw.renderWorkspace()
}

func tunnelLabel(t *sshmgr.TunnelSpec) string {
	if t.Name != "" {
		return t.Name
	}
	return t.Summary()
}

func defPort(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}

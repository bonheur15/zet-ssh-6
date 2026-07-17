package window

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"zet-terminal/internal/cmdlog"
	"zet-terminal/internal/sshmgr"
	"zet-terminal/internal/vault"
)

type remoteEntry struct {
	name  string
	isDir bool
}

// openFileBrowser shows a minimal remote directory browser for a profile.
// Listing runs over a shared ControlMaster connection so repeated calls
// are fast; downloads/uploads use scp and are logged. Password-auth hosts
// are unlocked from the vault first so listing can authenticate.
func (tw *TerminalWindow) openFileBrowser(p *sshmgr.Profile) {
	// Password-auth browsing needs the vault password (fed via sshpass)
	// and the sshpass tool. Ensure both before opening.
	if p != nil && p.Auth == sshmgr.AuthPassword {
		if _, err := exec.LookPath("sshpass"); err != nil {
			tw.notify("sshpass required", "Browsing password-authenticated hosts needs the 'sshpass' tool. Install it, or use SSH agent / key auth for this host.")
			return
		}
		if !appVault().IsUnlocked() || profilePassword(p) == "" {
			tw.WithVault(func(v *vault.Vault) {
				if profilePassword(p) == "" {
					tw.notify("No saved password", "This host has no password stored in the vault. Edit the host to save one, then browse again.")
					return
				}
				tw.showFileBrowser(p)
			})
			return
		}
	}
	tw.showFileBrowser(p)
}

func (tw *TerminalWindow) showFileBrowser(p *sshmgr.Profile) {
	dialog, box := tw.newDialog("Remote Files — "+p.Label(), 640, 560)

	// Shared control socket for this browser session.
	controlPath := fmt.Sprintf("/tmp/zet-ssh-%s-%d", sanitizeSock(p.ID), time.Now().Unix())

	header := gtk.NewBox(gtk.OrientationVertical, 4)
	title := gtk.NewLabel("Remote Files — " + p.Label())
	title.AddCSSClass("settings-page-title")
	title.SetHAlign(gtk.AlignStart)
	header.Append(title)
	pathLbl := gtk.NewLabel("")
	pathLbl.AddCSSClass("settings-page-subtitle")
	pathLbl.SetHAlign(gtk.AlignStart)
	pathLbl.SetEllipsize(pango.EllipsizeMiddle)
	header.Append(pathLbl)
	box.Append(header)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	box.Append(scrolled)
	listBox := gtk.NewBox(gtk.OrientationVertical, 2)
	scrolled.SetChild(listBox)

	cwd := p.RemoteDir
	if cwd == "" {
		cwd = "."
	}

	var loadDir func(dir string)

	setBusy := func(msg string) {
		for child := listBox.FirstChild(); child != nil; child = listBox.FirstChild() {
			listBox.Remove(child)
		}
		lbl := gtk.NewLabel(msg)
		lbl.AddCSSClass("sidebar-btn-label")
		lbl.SetMarginTop(20)
		lbl.SetHAlign(gtk.AlignCenter)
		listBox.Append(lbl)
	}

	loadDir = func(dir string) {
		setBusy("Listing " + dir + " …")
		go func() {
			entries, resolved, err := listRemoteDir(p, controlPath, dir)
			glib.IdleAdd(func() {
				if err != nil {
					setBusy("Error: " + err.Error())
					return
				}
				cwd = resolved
				pathLbl.SetLabel(resolved)
				for child := listBox.FirstChild(); child != nil; child = listBox.FirstChild() {
					listBox.Remove(child)
				}

				// Parent dir row
				if resolved != "/" {
					up := gtk.NewButton()
					up.AddCSSClass("sidebar-btn")
					up.SetHExpand(true)
					up.SetHAlign(gtk.AlignFill)
					upLbl := gtk.NewLabel("..")
					upLbl.AddCSSClass("sidebar-btn-label")
					upLbl.SetHAlign(gtk.AlignStart)
					upLbl.SetXAlign(0)
					up.SetChild(upLbl)
					up.ConnectClicked(func() { loadDir(path.Dir(resolved)) })
					listBox.Append(up)
				}

				for _, ent := range entries {
					e := ent
					row := gtk.NewBox(gtk.OrientationHorizontal, 4)
					row.AddCSSClass("tab-row")
					row.SetHExpand(true)

					icon := "text-x-generic-symbolic"
					if e.isDir {
						icon = "folder-symbolic"
					}
					nameBtn := gtk.NewButton()
					nameBtn.AddCSSClass("tab-select-btn")
					nameBtn.SetHExpand(true)
					nameBtn.SetHAlign(gtk.AlignFill)
					inner := gtk.NewBox(gtk.OrientationHorizontal, 8)
					inner.Append(gtk.NewImageFromIconName(icon))
					nm := gtk.NewLabel(e.name)
					nm.AddCSSClass("tab-label")
					nm.SetHAlign(gtk.AlignStart)
					nm.SetXAlign(0)
					nm.SetEllipsize(pango.EllipsizeEnd)
					nm.SetMaxWidthChars(40)
					inner.Append(nm)
					nameBtn.SetChild(inner)

					fullPath := path.Join(resolved, e.name)
					if e.isDir {
						nameBtn.ConnectClicked(func() { loadDir(fullPath) })
					} else {
						nameBtn.SetTooltipText(fullPath)
					}
					row.Append(nameBtn)

					if !e.isDir {
						row.Append(sidebarActionIcon("folder-download-symbolic", "Download", "", func() {
							tw.downloadRemote(p, fullPath)
						}))
					}
					listBox.Append(row)
				}

				if len(entries) == 0 {
					empty := gtk.NewLabel("(empty)")
					empty.AddCSSClass("sidebar-btn-label")
					empty.SetMarginTop(16)
					empty.SetHAlign(gtk.AlignCenter)
					listBox.Append(empty)
				}
			})
		}()
	}

	// Footer: upload + open sftp + close
	footer := gtk.NewBox(gtk.OrientationHorizontal, 10)
	footer.SetHAlign(gtk.AlignEnd)
	box.Append(footer)

	uploadBtn := gtk.NewButtonWithLabel("Upload here…")
	uploadBtn.AddCSSClass("sidebar-btn")
	uploadBtn.ConnectClicked(func() {
		chooser := gtk.NewFileChooserNative("Select file to upload", dialog, gtk.FileChooserActionOpen, "Upload", "Cancel")
		chooser.ConnectResponse(func(resp int) {
			if resp == int(gtk.ResponseAccept) {
				if f := chooser.File(); f != nil {
					local := f.Path()
					remote := path.Join(cwd, path.Base(local))
					tw.RunTransfer(p, local, remote, true, false)
				}
			}
			chooser.Destroy()
		})
		chooser.Show()
	})
	footer.Append(uploadBtn)

	sftpBtn := gtk.NewButtonWithLabel("Open SFTP")
	sftpBtn.AddCSSClass("sidebar-btn")
	sftpBtn.ConnectClicked(func() {
		tw.OpenSFTPTab(p)
		dialog.Close()
	})
	footer.Append(sftpBtn)

	closeBtn := gtk.NewButton()
	closeBtn.AddCSSClass("workspace-action-btn")
	closeLbl := gtk.NewLabel("Close")
	closeLbl.AddCSSClass("workspace-action-label")
	closeBtn.SetChild(closeLbl)
	closeBtn.ConnectClicked(func() { dialog.Close() })
	footer.Append(closeBtn)

	// Tear down the control master when the browser closes.
	dialog.ConnectCloseRequest(func() bool {
		go closeControlMaster(p, controlPath)
		return false
	})

	loadDir(cwd)
	dialog.Present()
}

func (tw *TerminalWindow) downloadRemote(p *sshmgr.Profile, remotePath string) {
	home, _ := os.UserHomeDir()
	local := home
	if local == "" {
		local = "."
	}
	tw.RunTransfer(p, local, remotePath, false, false)
}

// listRemoteDir runs `ls` remotely and parses names + types. Returns the
// resolved absolute path so ".." navigation works.
func listRemoteDir(p *sshmgr.Profile, controlPath, dir string) ([]remoteEntry, string, error) {
	// -p appends "/" to directories; pwd resolves the absolute path.
	remoteCmd := fmt.Sprintf("cd %s && pwd && ls -1Ap", shellQuoteArg(dir))

	// Password-auth hosts drive the ssh prompt with sshpass (BatchMode
	// off); agent/key hosts fail fast (BatchMode on).
	password := profilePassword(p)
	argv := p.ExecArgv(controlPath, remoteCmd, password == "")
	cmdlog.Add("List "+p.Label()+":"+dir, sshmgr.CommandString(argv))

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	wrapped := sshmgr.WrapSSHPass(argv, password)
	cmd := exec.CommandContext(ctx, wrapped[0], wrapped[1:]...)
	if password != "" {
		cmd.Env = append(os.Environ(), "SSHPASS="+password)
	}
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, dir, fmt.Errorf("connection timed out (this host needs agent or key auth for browsing)")
		}
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return nil, dir, fmt.Errorf("%s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, dir, err
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) == 0 {
		return nil, dir, nil
	}
	resolved := strings.TrimSpace(lines[0])
	var entries []remoteEntry
	for _, line := range lines[1:] {
		name := line
		if name == "" || name == "./" || name == "../" {
			continue
		}
		isDir := strings.HasSuffix(name, "/")
		name = strings.TrimSuffix(name, "/")
		entries = append(entries, remoteEntry{name: name, isDir: isDir})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir
		}
		return strings.ToLower(entries[i].name) < strings.ToLower(entries[j].name)
	})
	return entries, resolved, nil
}

func closeControlMaster(p *sshmgr.Profile, controlPath string) {
	argv := []string{"ssh", "-o", "ControlPath=" + controlPath, "-O", "exit", p.Target()}
	_ = exec.Command(argv[0], argv[1:]...).Run()
	_ = os.Remove(controlPath)
}

func sanitizeSock(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
}

// shellQuoteArg single-quotes an argument for safe remote shell use.
func shellQuoteArg(s string) string {
	if s == "" {
		return "'.'"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

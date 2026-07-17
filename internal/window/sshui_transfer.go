package window

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/sshmgr"
)

var transferDirections = []string{"Download (remote → local)", "Upload (local → remote)"}

// openTransferDialog builds an scp transfer for the given profile with a
// live command preview before running it in a visible tab.
func (tw *TerminalWindow) openTransferDialog(p *sshmgr.Profile) {
	dialog, box := tw.newDialog("Transfer Files", 560, 420)
	dialogTitle(box, "SCP Transfer — "+p.Label(), "Move files over scp. Review the exact command below before running it.")

	content := scrollArea(box)

	comboDir := labeledCombo(content, "Direction:", transferDirections, 0)
	comboRecursive := labeledCombo(content, "Recursive (-r):", []string{"No", "Yes"}, 0)

	remoteDefault := p.RemoteDir
	entryRemote := labeledEntry(content, "Remote Path:", remoteDefault, "/var/www/app.tar.gz")
	entryLocal := labeledEntry(content, "Local Path:", "", "~/Downloads")

	// Local file/dir browse button
	browseRow := gtk.NewBox(gtk.OrientationHorizontal, 8)
	browseRow.SetHAlign(gtk.AlignEnd)
	browseBtn := gtk.NewButtonWithLabel("Browse local…")
	browseBtn.AddCSSClass("sidebar-btn")
	browseBtn.ConnectClicked(func() {
		chooser := gtk.NewFileChooserNative("Select local path", dialog, gtk.FileChooserActionOpen, "Select", "Cancel")
		chooser.ConnectResponse(func(resp int) {
			if resp == int(gtk.ResponseAccept) {
				if f := chooser.File(); f != nil {
					entryLocal.SetText(f.Path())
				}
			}
			chooser.Destroy()
		})
		chooser.Show()
	})
	browseRow.Append(browseBtn)
	content.Append(browseRow)

	previewLbl, previewCopy := commandPreview(content, "")

	build := func() []string {
		upload := comboDir.Active() == 1
		recursive := comboRecursive.Active() == 1
		return p.SCPArgv(strings.TrimSpace(entryLocal.Text()), strings.TrimSpace(entryRemote.Text()), upload, recursive)
	}
	refresh := func() {
		previewLbl.SetLabel(sshmgr.CommandString(build()))
	}
	comboDir.Connect("changed", refresh)
	comboRecursive.Connect("changed", refresh)
	entryRemote.ConnectChanged(refresh)
	entryLocal.ConnectChanged(refresh)
	refresh()

	previewCopy.ConnectClicked(func() { copyToClipboard(previewLbl.Text()) })

	dialogButtons(box, dialog, "Run Transfer", func() {
		if strings.TrimSpace(entryLocal.Text()) == "" || strings.TrimSpace(entryRemote.Text()) == "" {
			return
		}
		upload := comboDir.Active() == 1
		recursive := comboRecursive.Active() == 1
		tw.RunTransfer(p, strings.TrimSpace(entryLocal.Text()), strings.TrimSpace(entryRemote.Text()), upload, recursive)
		dialog.Close()
	})

	dialog.Present()
}

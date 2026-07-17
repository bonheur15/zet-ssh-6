package window

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/cmdlog"
)

// notify shows a small modal message dialog.
func (tw *TerminalWindow) notify(title, message string) {
	dialog, box := tw.newDialog(title, 380, -1)
	dialogTitle(box, title, "")
	msg := gtk.NewLabel(message)
	msg.AddCSSClass("settings-page-subtitle")
	msg.SetHAlign(gtk.AlignStart)
	msg.SetWrap(true)
	box.Append(msg)

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginTop(6)
	box.Append(btnBox)
	ok := gtk.NewButton()
	ok.AddCSSClass("workspace-action-btn")
	lbl := gtk.NewLabel("OK")
	lbl.AddCSSClass("workspace-action-label")
	ok.SetChild(lbl)
	ok.ConnectClicked(func() { dialog.Close() })
	btnBox.Append(ok)
	dialog.Present()
}

// confirm shows a Yes/No dialog and calls onYes when confirmed.
func (tw *TerminalWindow) confirm(title, message string, onYes func()) {
	dialog, box := tw.newDialog(title, 400, -1)
	dialogTitle(box, title, "")
	msg := gtk.NewLabel(message)
	msg.AddCSSClass("settings-page-subtitle")
	msg.SetHAlign(gtk.AlignStart)
	msg.SetWrap(true)
	box.Append(msg)

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginTop(6)
	box.Append(btnBox)

	cancel := gtk.NewButtonWithLabel("Cancel")
	cancel.AddCSSClass("sidebar-btn")
	cancel.ConnectClicked(func() { dialog.Close() })
	btnBox.Append(cancel)

	yes := gtk.NewButton()
	yes.AddCSSClass("workspace-action-btn")
	lbl := gtk.NewLabel("Confirm")
	lbl.AddCSSClass("workspace-action-label")
	yes.SetChild(lbl)
	yes.ConnectClicked(func() {
		dialog.Close()
		onYes()
	})
	btnBox.Append(yes)
	dialog.Present()
}

// openCommandLog shows the auditable command history window.
func (tw *TerminalWindow) openCommandLog() {
	dialog, box := tw.newDialog("Command Log", 720, 560)
	dialogTitle(box, "Command Log", "Every SSH, SCP, and tunnel command Zet-SSH generated this session. Click a command to copy it.")

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	box.Append(scrolled)

	listBox := gtk.NewBox(gtk.OrientationVertical, 6)
	scrolled.SetChild(listBox)

	var rebuild func()
	rebuild = func() {
		for child := listBox.FirstChild(); child != nil; child = listBox.FirstChild() {
			listBox.Remove(child)
		}
		entries := cmdlog.Entries()
		if len(entries) == 0 {
			empty := gtk.NewLabel("No commands generated yet.")
			empty.AddCSSClass("settings-hint")
			empty.SetMarginTop(20)
			listBox.Append(empty)
			return
		}
		for _, e := range entries {
			entry := e
			card := gtk.NewBox(gtk.OrientationVertical, 2)
			card.AddCSSClass("cmdlog-card")

			head := gtk.NewBox(gtk.OrientationHorizontal, 8)
			meta := gtk.NewLabel(entry.Time.Format("15:04:05") + "  ·  " + entry.Label)
			meta.AddCSSClass("cmdlog-meta")
			meta.SetHAlign(gtk.AlignStart)
			meta.SetHExpand(true)
			head.Append(meta)

			copyBtn := gtk.NewButton()
			copyBtn.AddCSSClass("tab-action-btn")
			copyBtn.SetTooltipText("Copy command")
			copyBtn.SetChild(gtk.NewImageFromIconName("edit-copy-symbolic"))
			copyBtn.ConnectClicked(func() { copyToClipboard(entry.Command) })
			head.Append(copyBtn)
			card.Append(head)

			cmd := gtk.NewLabel(entry.Command)
			cmd.AddCSSClass("cmd-preview-label")
			cmd.SetHAlign(gtk.AlignStart)
			cmd.SetXAlign(0)
			cmd.SetWrap(true)
			cmd.SetSelectable(true)
			card.Append(cmd)

			listBox.Append(card)
		}
	}
	rebuild()

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetHAlign(gtk.AlignEnd)
	box.Append(btnBox)

	clear := gtk.NewButtonWithLabel("Clear Log")
	clear.AddCSSClass("sidebar-btn")
	clear.ConnectClicked(func() {
		cmdlog.Clear()
		rebuild()
	})
	btnBox.Append(clear)

	closeBtn := gtk.NewButton()
	closeBtn.AddCSSClass("workspace-action-btn")
	closeLbl := gtk.NewLabel("Close")
	closeLbl.AddCSSClass("workspace-action-label")
	closeBtn.SetChild(closeLbl)
	closeBtn.ConnectClicked(func() { dialog.Close() })
	btnBox.Append(closeBtn)

	dialog.Present()
}

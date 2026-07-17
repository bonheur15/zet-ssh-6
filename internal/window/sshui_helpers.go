package window

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// newDialog creates a modal styled window transient for this terminal.
func (tw *TerminalWindow) newDialog(title string, w, h int) (*gtk.Window, *gtk.Box) {
	dialog := gtk.NewWindow()
	dialog.SetTitle(title)
	dialog.SetTransientFor(&tw.Win.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(w, h)
	dialog.AddCSSClass("settings-dialog")

	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		if keyval == gdk.KEY_Escape {
			dialog.Close()
			return true
		}
		return false
	})
	dialog.AddController(keyCtrl)

	box := gtk.NewBox(gtk.OrientationVertical, 14)
	box.AddCSSClass("settings-box")
	dialog.SetChild(box)
	return dialog, box
}

// dialogTitle appends a title + subtitle header to a dialog box.
func dialogTitle(box *gtk.Box, title, subtitle string) {
	t := gtk.NewLabel(title)
	t.AddCSSClass("settings-page-title")
	t.SetHAlign(gtk.AlignStart)
	box.Append(t)
	if subtitle != "" {
		s := gtk.NewLabel(subtitle)
		s.AddCSSClass("settings-page-subtitle")
		s.SetHAlign(gtk.AlignStart)
		s.SetWrap(true)
		box.Append(s)
	}
}

// labeledEntry builds a "Label: [entry]" row.
func labeledEntry(parent *gtk.Box, label, value, placeholder string) *gtk.Entry {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	row.AddCSSClass("settings-row")
	lbl := gtk.NewLabel(label)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	row.Append(lbl)
	entry := gtk.NewEntry()
	entry.AddCSSClass("settings-entry")
	entry.SetWidthChars(24)
	if value != "" {
		entry.SetText(value)
	}
	if placeholder != "" {
		entry.SetPlaceholderText(placeholder)
	}
	row.Append(entry)
	parent.Append(row)
	return entry
}

// labeledCombo builds a "Label: [dropdown]" row with the given options.
func labeledCombo(parent *gtk.Box, label string, options []string, active int) *gtk.ComboBoxText {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	row.AddCSSClass("settings-row")
	lbl := gtk.NewLabel(label)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	row.Append(lbl)
	combo := gtk.NewComboBoxText()
	combo.AddCSSClass("settings-dropdown")
	for _, o := range options {
		combo.AppendText(o)
	}
	if active >= 0 {
		combo.SetActive(active)
	}
	row.Append(combo)
	parent.Append(row)
	return combo
}

// labeledSpin builds a "Label: [spin]" row.
func labeledSpin(parent *gtk.Box, label string, val, min, max, step float64) *gtk.SpinButton {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	row.AddCSSClass("settings-row")
	lbl := gtk.NewLabel(label)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	row.Append(lbl)
	adj := gtk.NewAdjustment(val, min, max, step, step*10, 0)
	spin := gtk.NewSpinButton(adj, step, 0)
	spin.AddCSSClass("settings-entry")
	row.Append(spin)
	parent.Append(row)
	return spin
}

// dialogButtons appends a right-aligned Cancel + primary button pair and
// returns the primary button; onOK runs the save action.
func dialogButtons(box *gtk.Box, dialog *gtk.Window, okLabel string, onOK func()) *gtk.Button {
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginTop(6)
	box.Append(btnBox)

	cancel := gtk.NewButtonWithLabel("Cancel")
	cancel.AddCSSClass("sidebar-btn")
	cancel.ConnectClicked(func() { dialog.Close() })
	btnBox.Append(cancel)

	ok := gtk.NewButton()
	ok.AddCSSClass("workspace-action-btn")
	lbl := gtk.NewLabel(okLabel)
	lbl.AddCSSClass("workspace-action-label")
	ok.SetChild(lbl)
	ok.ConnectClicked(onOK)
	btnBox.Append(ok)
	return ok
}

// commandPreview builds a monospace, selectable, copyable command box.
func commandPreview(parent *gtk.Box, initial string) (*gtk.Label, *gtk.Button) {
	wrap := gtk.NewBox(gtk.OrientationHorizontal, 6)
	wrap.AddCSSClass("cmd-preview-row")

	lbl := gtk.NewLabel(initial)
	lbl.AddCSSClass("cmd-preview-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	lbl.SetSelectable(true)
	lbl.SetWrap(true)
	lbl.SetXAlign(0)
	wrap.Append(lbl)

	copyBtn := gtk.NewButton()
	copyBtn.AddCSSClass("tab-action-btn")
	copyBtn.SetTooltipText("Copy command")
	copyImg := gtk.NewImageFromIconName("edit-copy-symbolic")
	copyBtn.SetChild(copyImg)
	wrap.Append(copyBtn)

	parent.Append(wrap)
	return lbl, copyBtn
}

// copyToClipboard puts text on the system clipboard.
func copyToClipboard(text string) {
	if display := gdk.DisplayGetDefault(); display != nil {
		if cb := display.Clipboard(); cb != nil {
			cb.SetText(text)
		}
	}
}

// scrollArea wraps content in a vertical scroller that expands.
func scrollArea(box *gtk.Box) *gtk.Box {
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	inner := gtk.NewBox(gtk.OrientationVertical, 12)
	scrolled.SetChild(inner)
	box.Append(scrolled)
	return inner
}

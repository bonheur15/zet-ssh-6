package window

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/config"
)

func (tw *TerminalWindow) openSettingsDialog() {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Settings")
	dialog.SetTransientFor(&tw.Win.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(400, 350)
	dialog.AddCSSClass("settings-dialog")

	// Set up layout
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.AddCSSClass("settings-box")
	dialog.SetChild(box)

	// Scrollable content area
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	box.Append(scrolled)

	contentBox := gtk.NewBox(gtk.OrientationVertical, 12)
	scrolled.SetChild(contentBox)

	// --- SECTION: SHELL & GENERAL ---
	lblGenTitle := gtk.NewLabel("General Settings")
	lblGenTitle.AddCSSClass("settings-section-title")
	lblGenTitle.SetHAlign(gtk.AlignStart)
	contentBox.Append(lblGenTitle)

	// Shell row
	rowShell := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblShell := gtk.NewLabel("Default Shell:")
	lblShell.AddCSSClass("settings-label")
	lblShell.SetHAlign(gtk.AlignStart)
	lblShell.SetHExpand(true)
	entryShell := gtk.NewEntry()
	entryShell.AddCSSClass("settings-entry")
	entryShell.SetText(tw.Cfg.Shell)
	rowShell.Append(lblShell)
	rowShell.Append(entryShell)
	contentBox.Append(rowShell)

	// Scrollback lines row
	rowScrollback := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblScrollback := gtk.NewLabel("Scrollback Lines:")
	lblScrollback.AddCSSClass("settings-label")
	lblScrollback.SetHAlign(gtk.AlignStart)
	lblScrollback.SetHExpand(true)
	adjScrollback := gtk.NewAdjustment(float64(tw.Cfg.ScrollbackLines), 100, 1000000, 500, 5000, 0)
	spinScrollback := gtk.NewSpinButton(adjScrollback, 100, 0)
	spinScrollback.AddCSSClass("settings-entry")
	rowScrollback.Append(lblScrollback)
	rowScrollback.Append(spinScrollback)
	contentBox.Append(rowScrollback)

	// --- SECTION: APPEARANCE ---
	lblAppTitle := gtk.NewLabel("Appearance")
	lblAppTitle.AddCSSClass("settings-section-title")
	lblAppTitle.SetHAlign(gtk.AlignStart)
	lblAppTitle.SetMarginTop(12)
	contentBox.Append(lblAppTitle)

	// Font Family row
	rowFontName := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblFontName := gtk.NewLabel("Font Family:")
	lblFontName.AddCSSClass("settings-label")
	lblFontName.SetHAlign(gtk.AlignStart)
	lblFontName.SetHExpand(true)
	entryFontName := gtk.NewEntry()
	entryFontName.AddCSSClass("settings-entry")
	entryFontName.SetText(tw.Cfg.FontName)
	rowFontName.Append(lblFontName)
	rowFontName.Append(entryFontName)
	contentBox.Append(rowFontName)

	// Font Size row
	rowFontSize := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblFontSize := gtk.NewLabel("Font Size:")
	lblFontSize.AddCSSClass("settings-label")
	lblFontSize.SetHAlign(gtk.AlignStart)
	lblFontSize.SetHExpand(true)
	adjFontSize := gtk.NewAdjustment(float64(tw.Cfg.FontSize), 4, 72, 1, 5, 0)
	spinFontSize := gtk.NewSpinButton(adjFontSize, 1, 0)
	spinFontSize.AddCSSClass("settings-entry")
	rowFontSize.Append(lblFontSize)
	rowFontSize.Append(spinFontSize)
	contentBox.Append(rowFontSize)

	// Cursor Shape row
	rowCursorShape := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblCursorShape := gtk.NewLabel("Cursor Shape:")
	lblCursorShape.AddCSSClass("settings-label")
	lblCursorShape.SetHAlign(gtk.AlignStart)
	lblCursorShape.SetHExpand(true)

	comboCursorShape := gtk.NewComboBoxText()
	comboCursorShape.AddCSSClass("settings-dropdown")
	comboCursorShape.AppendText("Block")
	comboCursorShape.AppendText("I-Beam")
	comboCursorShape.AppendText("Underline")
	comboCursorShape.SetActive(tw.Cfg.CursorShape)
	rowCursorShape.Append(lblCursorShape)
	rowCursorShape.Append(comboCursorShape)
	contentBox.Append(rowCursorShape)

	// Cursor Blink Mode row
	rowCursorBlink := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblCursorBlink := gtk.NewLabel("Cursor Blink:")
	lblCursorBlink.AddCSSClass("settings-label")
	lblCursorBlink.SetHAlign(gtk.AlignStart)
	lblCursorBlink.SetHExpand(true)

	comboCursorBlink := gtk.NewComboBoxText()
	comboCursorBlink.AddCSSClass("settings-dropdown")
	comboCursorBlink.AppendText("System")
	comboCursorBlink.AppendText("On")
	comboCursorBlink.AppendText("Off")
	comboCursorBlink.SetActive(tw.Cfg.CursorBlinkMode)
	rowCursorBlink.Append(lblCursorBlink)
	rowCursorBlink.Append(comboCursorBlink)
	contentBox.Append(rowCursorBlink)

	// --- BOTTOM BUTTONS: SAVE & CANCEL ---
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 12)
	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginTop(20)
	box.Append(btnBox)

	btnCancel := gtk.NewButtonWithLabel("Cancel")
	btnCancel.AddCSSClass("sidebar-btn")
	btnCancel.ConnectClicked(func() {
		dialog.Close()
	})
	btnBox.Append(btnCancel)

	btnSave := gtk.NewButton()
	btnSave.AddCSSClass("workspace-action-btn")
	lblSave := gtk.NewLabel("Save Configuration")
	lblSave.AddCSSClass("workspace-action-label")
	btnSave.SetChild(lblSave)

	btnSave.ConnectClicked(func() {
		tw.Cfg.Shell = strings.TrimSpace(entryShell.Text())
		tw.Cfg.ScrollbackLines = int(spinScrollback.Value())
		tw.Cfg.FontName = strings.TrimSpace(entryFontName.Text())
		tw.Cfg.FontSize = int(spinFontSize.Value())

		shapeIdx := comboCursorShape.Active()
		if shapeIdx >= 0 {
			tw.Cfg.CursorShape = shapeIdx
		}

		blinkIdx := comboCursorBlink.Active()
		if blinkIdx >= 0 {
			tw.Cfg.CursorBlinkMode = blinkIdx
		}

		_ = config.SaveConfig(tw.Cfg)

		// Apply configuration to all active terminals immediately
		for _, tab := range tw.TabInstances {
			tw.applyConfigToInstance(tab.TermInst)
			tab.TermInst.SetScrollbackLines(tw.Cfg.ScrollbackLines)
			tab.TermInst.SetCursorBlinkMode(tw.Cfg.CursorBlinkMode)
			tab.TermInst.SetCursorShape(tw.Cfg.CursorShape)
		}

		dialog.Close()
	})
	btnBox.Append(btnSave)

	dialog.Present()
}

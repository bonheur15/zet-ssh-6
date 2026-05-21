package window

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/config"
	"zet-terminal/internal/theme"
)

var uiAccents = []string{"cyan", "purple", "emerald", "amber", "crimson", "steel", "custom"}

func getAccentIndex(accent string) int {
	for i, a := range uiAccents {
		if a == accent {
			return i
		}
	}
	return 0
}

var termThemes = []string{"default", "nord", "gruvbox", "solarized", "monokai", "onehalf", "custom"}

func getTermThemeIndex(theme string) int {
	for i, t := range termThemes {
		if t == theme {
			return i
		}
	}
	return 0
}

func isValidColor(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) != 3 && len(h) != 6 && len(h) != 8 {
			return false
		}
		for _, r := range h {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		return true
	}
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		return true
	}
	return false
}

func setupColorPreview(entry *gtk.Entry, preview *gtk.Box) {
	preview.AddCSSClass("color-preview")
	provider := gtk.NewCSSProvider()
	preview.StyleContext().AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	updatePreview := func() {
		color := strings.TrimSpace(entry.Text())
		if isValidColor(color) {
			provider.LoadFromData(fmt.Sprintf(
				".color-preview { background-color: %s; min-width: 16px; min-height: 16px; border-radius: 4px; border: 1px solid rgba(255, 255, 255, 0.25); }",
				color,
			))
		} else {
			provider.LoadFromData(".color-preview { background-color: transparent; min-width: 16px; min-height: 16px; border-radius: 4px; border: 1px solid rgba(255, 255, 255, 0.1); }")
		}
	}
	entry.Connect("changed", updatePreview)
	updatePreview()
}

func createColorField(labelText, defaultValue string) (*gtk.Box, *gtk.Entry) {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	
	lbl := gtk.NewLabel(labelText)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	row.Append(lbl)

	entry := gtk.NewEntry()
	entry.AddCSSClass("settings-entry")
	entry.SetText(defaultValue)
	
	preview := gtk.NewBox(gtk.OrientationHorizontal, 0)
	preview.SetSizeRequest(16, 16)
	preview.SetVAlign(gtk.AlignCenter)
	
	setupColorPreview(entry, preview)

	row.Append(entry)
	row.Append(preview)

	return row, entry
}

func (tw *TerminalWindow) openSettingsDialog() {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Settings")
	dialog.SetTransientFor(&tw.Win.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(450, 550)
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

	// --- SECTION: UI THEME ACCENT ---
	lblThemeTitle := gtk.NewLabel("UI Theme Accent")
	lblThemeTitle.AddCSSClass("settings-section-title")
	lblThemeTitle.SetHAlign(gtk.AlignStart)
	lblThemeTitle.SetMarginTop(12)
	contentBox.Append(lblThemeTitle)

	// Accent Preset Row
	rowAccent := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblAccent := gtk.NewLabel("UI Accent Preset:")
	lblAccent.AddCSSClass("settings-label")
	lblAccent.SetHAlign(gtk.AlignStart)
	lblAccent.SetHExpand(true)

	comboAccent := gtk.NewComboBoxText()
	comboAccent.AddCSSClass("settings-dropdown")
	comboAccent.AppendText("Cyan (Default)")
	comboAccent.AppendText("Purple")
	comboAccent.AppendText("Emerald")
	comboAccent.AppendText("Amber")
	comboAccent.AppendText("Crimson")
	comboAccent.AppendText("Steel")
	comboAccent.AppendText("Custom")
	comboAccent.SetActive(getAccentIndex(tw.Cfg.UIThemeAccent))
	rowAccent.Append(lblAccent)
	rowAccent.Append(comboAccent)
	contentBox.Append(rowAccent)

	// Custom Accent Box (shown only when "Custom" selected)
	customUIBox := gtk.NewBox(gtk.OrientationVertical, 8)
	
	rowCustomAccent, entryCustomAccent := createColorField("Accent Color (Hex):", tw.Cfg.CustomAccentColor)
	customUIBox.Append(rowCustomAccent)
	
	rowCustomGlow, entryCustomGlow := createColorField("Accent Glow Color:", tw.Cfg.CustomGlowColor)
	customUIBox.Append(rowCustomGlow)
	contentBox.Append(customUIBox)

	// --- SECTION: TERMINAL THEME & PALETTE ---
	lblTermTitle := gtk.NewLabel("Terminal Theme")
	lblTermTitle.AddCSSClass("settings-section-title")
	lblTermTitle.SetHAlign(gtk.AlignStart)
	lblTermTitle.SetMarginTop(12)
	contentBox.Append(lblTermTitle)

	rowTermTheme := gtk.NewBox(gtk.OrientationHorizontal, 8)
	lblTermTheme := gtk.NewLabel("Terminal Preset:")
	lblTermTheme.AddCSSClass("settings-label")
	lblTermTheme.SetHAlign(gtk.AlignStart)
	lblTermTheme.SetHExpand(true)

	comboTermTheme := gtk.NewComboBoxText()
	comboTermTheme.AddCSSClass("settings-dropdown")
	comboTermTheme.AppendText("Default")
	comboTermTheme.AppendText("Nord")
	comboTermTheme.AppendText("Gruvbox")
	comboTermTheme.AppendText("Solarized")
	comboTermTheme.AppendText("Monokai")
	comboTermTheme.AppendText("One Half")
	comboTermTheme.AppendText("Custom")
	comboTermTheme.SetActive(getTermThemeIndex(tw.Cfg.TermThemePreset))
	rowTermTheme.Append(lblTermTheme)
	rowTermTheme.Append(comboTermTheme)
	contentBox.Append(rowTermTheme)

	// Custom Terminal Color Box (shown only when "Custom" selected)
	customTermBox := gtk.NewBox(gtk.OrientationVertical, 8)
	
	rowTermFg, entryTermFg := createColorField("Foreground Color (Hex):", tw.Cfg.TermForeground)
	customTermBox.Append(rowTermFg)
	
	rowTermBg, entryTermBg := createColorField("Background Color (Hex):", tw.Cfg.TermBackground)
	customTermBox.Append(rowTermBg)

	// Palette section
	lblPaletteTitle := gtk.NewLabel("ANSI Palette (16 Colors)")
	lblPaletteTitle.AddCSSClass("settings-label")
	lblPaletteTitle.SetHAlign(gtk.AlignStart)
	lblPaletteTitle.SetMarginTop(6)
	customTermBox.Append(lblPaletteTitle)

	grid := gtk.NewGrid()
	grid.SetColumnSpacing(8)
	grid.SetRowSpacing(8)
	grid.SetHAlign(gtk.AlignStart)

	var entryPalette [16]*gtk.Entry
	for i := 0; i < 16; i++ {
		rowVal := i / 4
		colVal := i % 4

		cell := gtk.NewBox(gtk.OrientationHorizontal, 4)
		
		idxLbl := gtk.NewLabel(fmt.Sprintf("%d:", i))
		idxLbl.AddCSSClass("settings-label")
		idxLbl.SetWidthChars(2)
		cell.Append(idxLbl)

		entry := gtk.NewEntry()
		entry.AddCSSClass("settings-entry")
		entry.SetWidthChars(7)
		if i < len(tw.Cfg.TermPalette) {
			entry.SetText(tw.Cfg.TermPalette[i])
		} else {
			entry.SetText("#ffffff")
		}

		preview := gtk.NewBox(gtk.OrientationHorizontal, 0)
		preview.SetSizeRequest(14, 14)
		preview.SetVAlign(gtk.AlignCenter)
		
		setupColorPreview(entry, preview)

		cell.Append(entry)
		cell.Append(preview)

		grid.Attach(cell, colVal, rowVal, 1, 1)
		entryPalette[i] = entry
	}
	customTermBox.Append(grid)
	contentBox.Append(customTermBox)

	// Visibility toggle functions
	updateUIVisibility := func() {
		isCustom := comboAccent.Active() == 6
		customUIBox.SetVisible(isCustom)
	}
	comboAccent.Connect("changed", updateUIVisibility)
	updateUIVisibility()

	updateTermVisibility := func() {
		isCustom := comboTermTheme.Active() == 6
		customTermBox.SetVisible(isCustom)
	}
	comboTermTheme.Connect("changed", updateTermVisibility)
	updateTermVisibility()

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

		// Accent settings
		accentIdx := comboAccent.Active()
		if accentIdx >= 0 && accentIdx < len(uiAccents) {
			tw.Cfg.UIThemeAccent = uiAccents[accentIdx]
		}
		tw.Cfg.CustomAccentColor = strings.TrimSpace(entryCustomAccent.Text())
		tw.Cfg.CustomGlowColor = strings.TrimSpace(entryCustomGlow.Text())

		// Terminal Theme settings
		termThemeIdx := comboTermTheme.Active()
		if termThemeIdx >= 0 && termThemeIdx < len(termThemes) {
			tw.Cfg.TermThemePreset = termThemes[termThemeIdx]
		}
		tw.Cfg.TermForeground = strings.TrimSpace(entryTermFg.Text())
		tw.Cfg.TermBackground = strings.TrimSpace(entryTermBg.Text())

		var newPalette []string
		for i := 0; i < 16; i++ {
			newPalette = append(newPalette, strings.TrimSpace(entryPalette[i].Text()))
		}
		tw.Cfg.TermPalette = newPalette

		// Save configuration
		_ = config.SaveConfig(tw.Cfg)

		// 1. Reload the Global CSS Provider instantly
		if GlobalCSSProvider != nil {
			GlobalCSSProvider.LoadFromData(theme.GetCSS(tw.Cfg))
		}

		// 2. Apply config updates to all tabs in all active windows instantly
		for w := range activeWindows {
			w.Cfg = tw.Cfg
			for _, tab := range w.TabInstances {
				w.applyConfigToInstance(tab.TermInst)
				tab.TermInst.SetScrollbackLines(w.Cfg.ScrollbackLines)
				tab.TermInst.SetCursorBlinkMode(w.Cfg.CursorBlinkMode)
				tab.TermInst.SetCursorShape(w.Cfg.CursorShape)
			}
		}

		dialog.Close()
	})
	btnBox.Append(btnSave)

	dialog.Present()
}

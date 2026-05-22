package window

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
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

func createColorButtonField(labelText, hexValue string) (*gtk.Box, *gtk.ColorButton) {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	row.AddCSSClass("settings-row")

	lbl := gtk.NewLabel(labelText)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	row.Append(lbl)

	btn := gtk.NewColorButton()
	btn.AddCSSClass("settings-entry")

	c := gdk.NewRGBA(0, 0, 0, 0)
	c.Parse(hexValue)
	btn.SetRGBA(&c)

	row.Append(btn)
	return row, btn
}

func colorToHex(btn *gtk.ColorButton) string {
	rgba := btn.RGBA()
	r := int(rgba.Red() * 255)
	g := int(rgba.Green() * 255)
	b := int(rgba.Blue() * 255)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func createKeybindingField(labelText, value string) (*gtk.Box, *gtk.Entry) {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	row.AddCSSClass("settings-row")

	lbl := gtk.NewLabel(labelText)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	lbl.SetHExpand(true)
	row.Append(lbl)

	entry := gtk.NewEntry()
	entry.AddCSSClass("settings-entry")
	entry.SetWidthChars(16)
	entry.SetText(value)

	row.Append(entry)
	return row, entry
}

func createSettingsCard(title, description string) (*gtk.Box, *gtk.Box) {
	card := gtk.NewBox(gtk.OrientationVertical, 10)
	card.AddCSSClass("settings-card")

	header := gtk.NewBox(gtk.OrientationVertical, 2)
	header.AddCSSClass("settings-card-header")

	titleLabel := gtk.NewLabel(title)
	titleLabel.AddCSSClass("settings-section-title")
	titleLabel.SetHAlign(gtk.AlignStart)
	header.Append(titleLabel)

	if description != "" {
		descLabel := gtk.NewLabel(description)
		descLabel.AddCSSClass("settings-section-description")
		descLabel.SetHAlign(gtk.AlignStart)
		descLabel.SetWrap(true)
		header.Append(descLabel)
	}

	card.Append(header)

	content := gtk.NewBox(gtk.OrientationVertical, 10)
	content.AddCSSClass("settings-card-content")
	card.Append(content)

	return card, content
}

func createSwitchField(labelText, description string, active bool) (*gtk.Box, *gtk.Switch) {
	row := gtk.NewBox(gtk.OrientationHorizontal, 12)
	row.AddCSSClass("settings-row")

	copyBox := gtk.NewBox(gtk.OrientationVertical, 2)
	copyBox.SetHExpand(true)

	lbl := gtk.NewLabel(labelText)
	lbl.AddCSSClass("settings-label")
	lbl.SetHAlign(gtk.AlignStart)
	copyBox.Append(lbl)

	if description != "" {
		desc := gtk.NewLabel(description)
		desc.AddCSSClass("settings-hint")
		desc.SetHAlign(gtk.AlignStart)
		desc.SetWrap(true)
		copyBox.Append(desc)
	}

	sw := gtk.NewSwitch()
	sw.SetActive(active)

	row.Append(copyBox)
	row.Append(sw)
	return row, sw
}

func (tw *TerminalWindow) openSettingsDialog() {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Settings")
	dialog.SetTransientFor(&tw.Win.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(760, 680)
	dialog.AddCSSClass("settings-dialog")

	// Set up layout
	box := gtk.NewBox(gtk.OrientationVertical, 18)
	box.AddCSSClass("settings-box")
	dialog.SetChild(box)

	header := gtk.NewBox(gtk.OrientationVertical, 4)
	header.AddCSSClass("settings-header")
	box.Append(header)

	title := gtk.NewLabel("Terminal Preferences")
	title.AddCSSClass("settings-page-title")
	title.SetHAlign(gtk.AlignStart)
	header.Append(title)

	subtitle := gtk.NewLabel("Tune launch behavior, appearance, keyboard shortcuts, and terminal rendering.")
	subtitle.AddCSSClass("settings-page-subtitle")
	subtitle.SetHAlign(gtk.AlignStart)
	subtitle.SetWrap(true)
	header.Append(subtitle)

	// Scrollable content area
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	box.Append(scrolled)

	contentBox := gtk.NewBox(gtk.OrientationVertical, 16)
	scrolled.SetChild(contentBox)

	// --- SECTION: SHELL & GENERAL ---
	generalCard, generalContent := createSettingsCard("General", "Control the shell, window defaults, and session history behavior.")
	contentBox.Append(generalCard)

	// Shell row
	rowShell := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowShell.AddCSSClass("settings-row")
	lblShell := gtk.NewLabel("Default Shell:")
	lblShell.AddCSSClass("settings-label")
	lblShell.SetHAlign(gtk.AlignStart)
	lblShell.SetHExpand(true)
	entryShell := gtk.NewEntry()
	entryShell.AddCSSClass("settings-entry")
	entryShell.SetText(tw.Cfg.Shell)
	rowShell.Append(lblShell)
	rowShell.Append(entryShell)
	generalContent.Append(rowShell)

	// Scrollback lines row
	rowScrollback := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowScrollback.AddCSSClass("settings-row")
	lblScrollback := gtk.NewLabel("Scrollback Lines:")
	lblScrollback.AddCSSClass("settings-label")
	lblScrollback.SetHAlign(gtk.AlignStart)
	lblScrollback.SetHExpand(true)
	adjScrollback := gtk.NewAdjustment(float64(tw.Cfg.ScrollbackLines), 100, 1000000, 500, 5000, 0)
	spinScrollback := gtk.NewSpinButton(adjScrollback, 100, 0)
	spinScrollback.AddCSSClass("settings-entry")
	rowScrollback.Append(lblScrollback)
	rowScrollback.Append(spinScrollback)
	generalContent.Append(rowScrollback)

	rowHistoryLimit := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowHistoryLimit.AddCSSClass("settings-row")
	lblHistoryLimit := gtk.NewLabel("Saved Past Commands:")
	lblHistoryLimit.AddCSSClass("settings-label")
	lblHistoryLimit.SetHAlign(gtk.AlignStart)
	lblHistoryLimit.SetHExpand(true)
	adjHistoryLimit := gtk.NewAdjustment(float64(tw.Cfg.CommandHistoryLimit), 10, 500, 5, 25, 0)
	spinHistoryLimit := gtk.NewSpinButton(adjHistoryLimit, 5, 0)
	spinHistoryLimit.AddCSSClass("settings-entry")
	rowHistoryLimit.Append(lblHistoryLimit)
	rowHistoryLimit.Append(spinHistoryLimit)
	generalContent.Append(rowHistoryLimit)

	rowWindowWidth := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowWindowWidth.AddCSSClass("settings-row")
	lblWindowWidth := gtk.NewLabel("Default Window Width:")
	lblWindowWidth.AddCSSClass("settings-label")
	lblWindowWidth.SetHAlign(gtk.AlignStart)
	lblWindowWidth.SetHExpand(true)
	adjWindowWidth := gtk.NewAdjustment(float64(tw.Cfg.WindowWidth), 640, 3840, 20, 100, 0)
	spinWindowWidth := gtk.NewSpinButton(adjWindowWidth, 20, 0)
	spinWindowWidth.AddCSSClass("settings-entry")
	rowWindowWidth.Append(lblWindowWidth)
	rowWindowWidth.Append(spinWindowWidth)
	generalContent.Append(rowWindowWidth)

	rowWindowHeight := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowWindowHeight.AddCSSClass("settings-row")
	lblWindowHeight := gtk.NewLabel("Default Window Height:")
	lblWindowHeight.AddCSSClass("settings-label")
	lblWindowHeight.SetHAlign(gtk.AlignStart)
	lblWindowHeight.SetHExpand(true)
	adjWindowHeight := gtk.NewAdjustment(float64(tw.Cfg.WindowHeight), 420, 2160, 20, 100, 0)
	spinWindowHeight := gtk.NewSpinButton(adjWindowHeight, 20, 0)
	spinWindowHeight.AddCSSClass("settings-entry")
	rowWindowHeight.Append(lblWindowHeight)
	rowWindowHeight.Append(spinWindowHeight)
	generalContent.Append(rowWindowHeight)

	rowSidebarPinned, switchSidebarPinned := createSwitchField("Keep sidebar pinned by default", "New windows open with the workspace sidebar already visible.", tw.Cfg.SidebarPinned)
	generalContent.Append(rowSidebarPinned)

	// --- SECTION: APPEARANCE ---
	appearanceCard, appearanceContent := createSettingsCard("Appearance", "Set typography, cursor behavior, and the overall application accent.")
	contentBox.Append(appearanceCard)

	// Font Family row
	rowFontName := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowFontName.AddCSSClass("settings-row")
	lblFontName := gtk.NewLabel("Font Family:")
	lblFontName.AddCSSClass("settings-label")
	lblFontName.SetHAlign(gtk.AlignStart)
	lblFontName.SetHExpand(true)
	entryFontName := gtk.NewEntry()
	entryFontName.AddCSSClass("settings-entry")
	entryFontName.SetText(tw.Cfg.FontName)
	rowFontName.Append(lblFontName)
	rowFontName.Append(entryFontName)
	appearanceContent.Append(rowFontName)

	// Font Size row
	rowFontSize := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowFontSize.AddCSSClass("settings-row")
	lblFontSize := gtk.NewLabel("Font Size:")
	lblFontSize.AddCSSClass("settings-label")
	lblFontSize.SetHAlign(gtk.AlignStart)
	lblFontSize.SetHExpand(true)
	adjFontSize := gtk.NewAdjustment(float64(tw.Cfg.FontSize), 4, 72, 1, 5, 0)
	spinFontSize := gtk.NewSpinButton(adjFontSize, 1, 0)
	spinFontSize.AddCSSClass("settings-entry")
	rowFontSize.Append(lblFontSize)
	rowFontSize.Append(spinFontSize)
	appearanceContent.Append(rowFontSize)

	// Cursor Shape row
	rowCursorShape := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowCursorShape.AddCSSClass("settings-row")
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
	appearanceContent.Append(rowCursorShape)

	// Cursor Blink Mode row
	rowCursorBlink := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowCursorBlink.AddCSSClass("settings-row")
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
	appearanceContent.Append(rowCursorBlink)

	// --- SECTION: UI THEME ACCENT ---
	accentCard, accentContent := createSettingsCard("Interface Accent", "Choose the frame glow and accent color used across the terminal chrome.")
	contentBox.Append(accentCard)

	// Accent Preset Row
	rowAccent := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowAccent.AddCSSClass("settings-row")
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
	accentContent.Append(rowAccent)

	// Custom Accent Box (shown only when "Custom" selected)
	customUIBox := gtk.NewBox(gtk.OrientationVertical, 8)

	rowCustomAccent, btnCustomAccent := createColorButtonField("Custom Accent Color:", tw.Cfg.CustomAccentColor)
	customUIBox.Append(rowCustomAccent)

	rowCustomGlow, btnCustomGlow := createColorButtonField("Custom Glow Accent Color:", tw.Cfg.CustomGlowColor)
	customUIBox.Append(rowCustomGlow)
	accentContent.Append(customUIBox)

	// Glow Intensity / Opacity slider
	rowGlowOpacity := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowGlowOpacity.AddCSSClass("settings-row")
	lblGlowOpacity := gtk.NewLabel("Glow Intensity:")
	lblGlowOpacity.AddCSSClass("settings-label")
	lblGlowOpacity.SetHAlign(gtk.AlignStart)
	lblGlowOpacity.SetHExpand(true)

	adjGlowOpacity := gtk.NewAdjustment(tw.Cfg.UIThemeGlowOpacity*100, 0, 100, 5, 10, 0)
	scaleGlowOpacity := gtk.NewScale(gtk.OrientationHorizontal, adjGlowOpacity)
	scaleGlowOpacity.SetHExpand(true)
	scaleGlowOpacity.SetSizeRequest(160, -1)
	scaleGlowOpacity.SetDrawValue(true)
	scaleGlowOpacity.SetValuePos(gtk.PosRight)
	scaleGlowOpacity.SetDigits(0)
	rowGlowOpacity.Append(lblGlowOpacity)
	rowGlowOpacity.Append(scaleGlowOpacity)
	accentContent.Append(rowGlowOpacity)

	// --- SECTION: TERMINAL THEME & PALETTE ---
	terminalCard, terminalContent := createSettingsCard("Terminal Palette", "Pick a preset or customize the ANSI palette used by VTE.")
	contentBox.Append(terminalCard)

	rowTermTheme := gtk.NewBox(gtk.OrientationHorizontal, 8)
	rowTermTheme.AddCSSClass("settings-row")
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
	terminalContent.Append(rowTermTheme)

	// Custom Terminal Color Box (shown only when "Custom" selected)
	customTermBox := gtk.NewBox(gtk.OrientationVertical, 8)

	rowTermFg, btnTermFg := createColorButtonField("Foreground Color:", tw.Cfg.TermForeground)
	customTermBox.Append(rowTermFg)

	rowTermBg, btnTermBg := createColorButtonField("Background Color:", tw.Cfg.TermBackground)
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

	var btnPalette [16]*gtk.ColorButton
	for i := 0; i < 16; i++ {
		rowVal := i / 4
		colVal := i % 4

		cell := gtk.NewBox(gtk.OrientationHorizontal, 4)
		cell.SetVAlign(gtk.AlignCenter)

		idxLbl := gtk.NewLabel(fmt.Sprintf("%2d:", i))
		idxLbl.AddCSSClass("settings-label")
		idxLbl.SetWidthChars(3)
		cell.Append(idxLbl)

		btn := gtk.NewColorButton()
		btn.AddCSSClass("settings-entry")
		btn.SetSizeRequest(36, 24)

		hexVal := "#ffffff"
		if i < len(tw.Cfg.TermPalette) {
			hexVal = tw.Cfg.TermPalette[i]
		}

		c := gdk.NewRGBA(0, 0, 0, 0)
		c.Parse(hexVal)
		btn.SetRGBA(&c)

		cell.Append(btn)
		grid.Attach(cell, colVal, rowVal, 1, 1)
		btnPalette[i] = btn
	}
	customTermBox.Append(grid)
	terminalContent.Append(customTermBox)

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

	// --- SECTION: KEYBOARD SHORTCUTS ---
	keyboardCard, keyboardContent := createSettingsCard("Keyboard Shortcuts", "All shortcuts are editable. Use GTK key names such as ctrl+shift+n or ctrl+Page_Down.")
	contentBox.Append(keyboardCard)

	rowNewTab, entryNewTab := createKeybindingField("New Tab:", tw.Cfg.Keybindings.NewTab)
	keyboardContent.Append(rowNewTab)

	rowNewWindow, entryNewWindow := createKeybindingField("New Window:", tw.Cfg.Keybindings.NewWindow)
	keyboardContent.Append(rowNewWindow)

	rowCloseTab, entryCloseTab := createKeybindingField("Close Tab:", tw.Cfg.Keybindings.CloseTab)
	keyboardContent.Append(rowCloseTab)

	rowNextTab, entryNextTab := createKeybindingField("Next Tab:", tw.Cfg.Keybindings.NextTab)
	keyboardContent.Append(rowNextTab)

	rowPrevTab, entryPrevTab := createKeybindingField("Previous Tab:", tw.Cfg.Keybindings.PrevTab)
	keyboardContent.Append(rowPrevTab)

	rowCopy, entryCopy := createKeybindingField("Copy Selection:", tw.Cfg.Keybindings.Copy)
	keyboardContent.Append(rowCopy)

	rowPaste, entryPaste := createKeybindingField("Paste Clipboard:", tw.Cfg.Keybindings.Paste)
	keyboardContent.Append(rowPaste)

	rowToggleSidebar, entryToggleSidebar := createKeybindingField("Toggle Sidebar:", tw.Cfg.Keybindings.ToggleSidebar)
	keyboardContent.Append(rowToggleSidebar)

	rowZoomIn, entryZoomIn := createKeybindingField("Zoom In:", tw.Cfg.Keybindings.ZoomIn)
	keyboardContent.Append(rowZoomIn)

	rowZoomOut, entryZoomOut := createKeybindingField("Zoom Out:", tw.Cfg.Keybindings.ZoomOut)
	keyboardContent.Append(rowZoomOut)

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
		tw.Cfg.CommandHistoryLimit = int(spinHistoryLimit.Value())
		tw.Cfg.WindowWidth = int(spinWindowWidth.Value())
		tw.Cfg.WindowHeight = int(spinWindowHeight.Value())
		tw.Cfg.SidebarPinned = switchSidebarPinned.Active()
		tw.Cfg.FontName = strings.TrimSpace(entryFontName.Text())
		tw.Cfg.FontSize = int(spinFontSize.Value())

		// Keybindings
		tw.Cfg.Keybindings.NewTab = strings.TrimSpace(entryNewTab.Text())
		tw.Cfg.Keybindings.NewWindow = strings.TrimSpace(entryNewWindow.Text())
		tw.Cfg.Keybindings.CloseTab = strings.TrimSpace(entryCloseTab.Text())
		tw.Cfg.Keybindings.NextTab = strings.TrimSpace(entryNextTab.Text())
		tw.Cfg.Keybindings.PrevTab = strings.TrimSpace(entryPrevTab.Text())
		tw.Cfg.Keybindings.Copy = strings.TrimSpace(entryCopy.Text())
		tw.Cfg.Keybindings.Paste = strings.TrimSpace(entryPaste.Text())
		tw.Cfg.Keybindings.ToggleSidebar = strings.TrimSpace(entryToggleSidebar.Text())
		tw.Cfg.Keybindings.ZoomIn = strings.TrimSpace(entryZoomIn.Text())
		tw.Cfg.Keybindings.ZoomOut = strings.TrimSpace(entryZoomOut.Text())

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
		tw.Cfg.CustomAccentColor = colorToHex(btnCustomAccent)
		tw.Cfg.CustomGlowColor = colorToHex(btnCustomGlow)
		tw.Cfg.UIThemeGlowOpacity = scaleGlowOpacity.Value() / 100.0

		// Terminal Theme settings
		termThemeIdx := comboTermTheme.Active()
		if termThemeIdx >= 0 && termThemeIdx < len(termThemes) {
			tw.Cfg.TermThemePreset = termThemes[termThemeIdx]
		}
		tw.Cfg.TermForeground = colorToHex(btnTermFg)
		tw.Cfg.TermBackground = colorToHex(btnTermBg)

		var newPalette []string
		for i := 0; i < 16; i++ {
			newPalette = append(newPalette, colorToHex(btnPalette[i]))
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
			w.Cfg = cloneConfig(tw.Cfg)
			w.SidebarPinned = w.Cfg.SidebarPinned
			w.Revealer.SetRevealChild(w.SidebarPinned)
			w.Win.SetDefaultSize(w.Cfg.WindowWidth, w.Cfg.WindowHeight)
			for _, tab := range w.TabInstances {
				w.applyConfigToInstance(tab.TermInst)
				tab.TermInst.SetScrollbackLines(w.Cfg.ScrollbackLines)
				tab.TermInst.SetCursorBlinkMode(w.Cfg.CursorBlinkMode)
				tab.TermInst.SetCursorShape(w.Cfg.CursorShape)
			}
			if w.HistoryRevealer != nil && w.HistoryRevealer.RevealChild() {
				w.updateHistoryUI()
			}
			w.renderWorkspace()
		}

		dialog.Close()
	})
	btnBox.Append(btnSave)

	dialog.Present()
}

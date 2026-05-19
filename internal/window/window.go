package window

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/terminal"
	"zet-terminal/internal/theme"
)

type TerminalWindow struct {
	Win              *gtk.ApplicationWindow
	App              *gtk.Application
	Cfg              *config.Config
	TermInst         *terminal.VteTerminalInstance
	Container        *gtk.Box
	Overlay          *gtk.Overlay
	DockBox          *gtk.Box
	Revealer         *gtk.Revealer
	HistoryListBox   *gtk.Box
	currentCmdBuffer string
	popover          *gtk.Popover
	btnCopy          *gtk.Button
	btnPaste         *gtk.Button
}

func NewTerminalWindow(app *gtk.Application, cfg *config.Config) *TerminalWindow {
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("Terminal")
	win.SetDefaultSize(850, 550)
	win.AddCSSClass("terminal-window")

	tw := &TerminalWindow{
		Win: win,
		App: app,
		Cfg: cfg,
	}

	tw.setupUI()
	tw.setupShortcuts()

	// Show and present window to grab focus at OS/WM level
	win.Present()

	// Apply fonts & theme colors AFTER realization/show to ensure correct DPI and cell metrics
	tw.applyConfigToInstance(tw.TermInst)

	// Grab input focus directly on the terminal widget so the user can type immediately
	tw.TermInst.Widget.GrabFocus()

	return tw
}

func (tw *TerminalWindow) setupUI() {
	// ─── Minimalistic Title Bar ───
	header := gtk.NewHeaderBar()
	titleLabel := gtk.NewLabel("Terminal")
	titleLabel.SetHAlign(gtk.AlignCenter)
	titleLabel.SetVAlign(gtk.AlignCenter)
	header.SetTitleWidget(titleLabel)
	tw.Win.SetTitlebar(header)

	// Single terminal instance
	tw.TermInst = terminal.NewVteTerminal()
	tw.TermInst.Widget.AddCSSClass("vte-terminal-widget")
	tw.TermInst.Widget.SetHExpand(true)
	tw.TermInst.Widget.SetVExpand(true)
	tw.TermInst.SetScrollbackLines(tw.Cfg.ScrollbackLines)
	tw.TermInst.SetCursorBlinkMode(tw.Cfg.CursorBlinkMode)
	tw.TermInst.SetCursorShape(tw.Cfg.CursorShape)

	// Container box
	tw.Container = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.Container.AddCSSClass("terminal-container")
	tw.Container.SetHExpand(true)
	tw.Container.SetVExpand(true)
	tw.Container.Append(tw.TermInst.Widget)

	// Wrap in Overlay for side dock
	tw.Overlay = gtk.NewOverlay()
	tw.Overlay.SetChild(tw.Container)
	tw.Win.SetChild(tw.Overlay)

	// Set up the left hover dock
	tw.setupSideDock()

	// Spawn default shell inside PTY
	tw.TermInst.SpawnShell(tw.Cfg.Shell, "")

	// Signal Handlers
	tw.TermInst.OnChildExited(func(status int) {
		tw.TermInst.Destroy()
		tw.Win.Close()
	})

	tw.TermInst.OnWindowTitleChanged(func(title string) {
		if title == "" {
			title = "Terminal"
		}
		titleLabel.SetLabel(title)
	})

	tw.setupContextMenu()
}

func (tw *TerminalWindow) applyConfigToInstance(inst *terminal.VteTerminalInstance) {
	inst.SetFont(tw.Cfg.FontName, tw.Cfg.FontSize)
	inst.SetColors(theme.DefaultPalette.Foreground, theme.DefaultPalette.Background, theme.DefaultPalette.Palette)
}

func (tw *TerminalWindow) setupShortcuts() {
	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		isCtrl := (state & gdk.ControlMask) != 0
		isShift := (state & gdk.ShiftMask) != 0
		isAlt := (state & gdk.AltMask) != 0

		if isCtrl && isShift {
			switch keyval {
			case 'N', 'n': // Ctrl+Shift+N -> New Window
				NewTerminalWindow(tw.App, tw.Cfg)
				return true
			case 'C', 'c': // Ctrl+Shift+C -> Copy
				if tw.TermInst != nil {
					tw.TermInst.Copy()
				}
				return true
			case 'V', 'v': // Ctrl+Shift+V -> Paste
				if tw.TermInst != nil {
					tw.TermInst.Paste()
				}
				return true
			case '+', '=': // Ctrl+Shift++ or Ctrl+Shift+= -> Zoom In
				tw.Cfg.FontSize++
				if tw.Cfg.FontSize > 72 {
					tw.Cfg.FontSize = 72
				}
				_ = config.SaveConfig(tw.Cfg)
				tw.applyConfigToInstance(tw.TermInst)
				return true
			case '-', '_': // Ctrl+Shift+- or Ctrl+Shift+_ -> Zoom Out
				tw.Cfg.FontSize--
				if tw.Cfg.FontSize < 4 {
					tw.Cfg.FontSize = 4
				}
				_ = config.SaveConfig(tw.Cfg)
				tw.applyConfigToInstance(tw.TermInst)
				return true
			}
		}

		// Track command history typing silently in the background
		if !isCtrl && !isAlt {
			switch keyval {
			case 0xff0d, 0xff8d: // Enter key (Return or Keypad Enter)
				tw.AddCommandToHistory(tw.currentCmdBuffer)
				tw.currentCmdBuffer = ""
			case 0xff08: // Backspace key
				if len(tw.currentCmdBuffer) > 0 {
					runes := []rune(tw.currentCmdBuffer)
					tw.currentCmdBuffer = string(runes[:len(runes)-1])
				}
			default:
				// Track only standard printable ASCII characters
				if keyval >= 32 && keyval <= 126 {
					tw.currentCmdBuffer += string(rune(keyval))
				}
			}
		} else if isCtrl && (keyval == 'C' || keyval == 'c' || keyval == 'D' || keyval == 'd') {
			// Ctrl+C or Ctrl+D cancels current line input
			tw.currentCmdBuffer = ""
		}

		return false
	})
	tw.Win.AddController(keyCtrl)
}

func (tw *TerminalWindow) setupSideDock() {
	// ─── Revealer (slides horizontal) ───
	tw.Revealer = gtk.NewRevealer()
	tw.Revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideRight)
	tw.Revealer.SetTransitionDuration(250)
	tw.Revealer.SetRevealChild(false)

	// Panel Container Box
	panelBox := gtk.NewBox(gtk.OrientationVertical, 0)
	panelBox.AddCSSClass("sidebar-panel")
	panelBox.SetSizeRequest(220, -1)
	panelBox.SetVExpand(true)

	// Panel Title
	titleLabel := gtk.NewLabel("COMMANDS")
	titleLabel.AddCSSClass("sidebar-title")
	titleLabel.SetHAlign(gtk.AlignStart)
	panelBox.Append(titleLabel)

	// Scrolled window for history list
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetHExpand(true)
	panelBox.Append(scrolled)

	// List container inside scrolled
	tw.HistoryListBox = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.HistoryListBox.SetVExpand(true)
	tw.HistoryListBox.SetHExpand(true)
	scrolled.SetChild(tw.HistoryListBox)

	tw.Revealer.SetChild(panelBox)

	// Thin vertical hover trigger bar
	triggerBar := gtk.NewBox(gtk.OrientationVertical, 0)
	triggerBar.AddCSSClass("sidebar-trigger")
	triggerBar.SetSizeRequest(6, -1)
	triggerBar.SetVExpand(true)

	// Main Dock Box holding [ Revealer | TriggerBar ]
	tw.DockBox = gtk.NewBox(gtk.OrientationHorizontal, 0)
	tw.DockBox.AddCSSClass("sidebar-dock")
	tw.DockBox.SetHAlign(gtk.AlignStart)
	tw.DockBox.SetVAlign(gtk.AlignFill)
	tw.DockBox.Append(tw.Revealer)
	tw.DockBox.Append(triggerBar)

	// Add DockBox as an overlay child
	tw.Overlay.AddOverlay(tw.DockBox)

	// Add Hover Event Controller
	motionCtrl := gtk.NewEventControllerMotion()
	motionCtrl.ConnectEnter(func(x float64, y float64) {
		tw.updateHistoryUI()
		tw.Revealer.SetRevealChild(true)
	})
	motionCtrl.ConnectLeave(func() {
		tw.Revealer.SetRevealChild(false)
	})
	tw.DockBox.AddController(motionCtrl)

	// Populate initially
	tw.updateHistoryUI()
}

func (tw *TerminalWindow) updateHistoryUI() {
	// Clear existing list items
	for child := tw.HistoryListBox.FirstChild(); child != nil; child = tw.HistoryListBox.FirstChild() {
		tw.HistoryListBox.Remove(child)
	}

	// If history is empty, show a placeholder label
	if len(tw.Cfg.CommandHistory) == 0 {
		emptyLabel := gtk.NewLabel("No commands run yet.")
		emptyLabel.AddCSSClass("sidebar-btn-label")
		emptyLabel.SetMarginTop(20)
		emptyLabel.SetHAlign(gtk.AlignCenter)
		tw.HistoryListBox.Append(emptyLabel)
		return
	}

	// Populate flat command rows: [ Text Button (Copy on click) | Arrow Button (Execute on click) ]
	for _, cmd := range tw.Cfg.CommandHistory {
		cmdStr := cmd

		// Row Box
		row := gtk.NewBox(gtk.OrientationHorizontal, 0)
		row.AddCSSClass("sidebar-row")
		row.SetHExpand(true)

		// Text Button showing command (Copies command on click)
		btnText := gtk.NewButton()
		btnText.AddCSSClass("sidebar-btn")
		btnText.SetHExpand(true)
		btnText.SetHAlign(gtk.AlignFill)
		btnText.SetTooltipText("Click to copy to clipboard")

		lbl := gtk.NewLabel(cmdStr)
		lbl.AddCSSClass("sidebar-btn-label")
		lbl.SetHAlign(gtk.AlignStart)
		lbl.SetXAlign(0.0)
		btnText.SetChild(lbl)

		btnText.ConnectClicked(func() {
			display := gdk.DisplayGetDefault()
			if display != nil {
				clipboard := display.Clipboard()
				if clipboard != nil {
					clipboard.SetText(cmdStr)
				}
			}
		})
		row.Append(btnText)

		// Arrow Button to execute command
		btnArrow := gtk.NewButton()
		btnArrow.AddCSSClass("sidebar-arrow-btn")
		btnArrow.SetTooltipText("Execute command")

		lblArrow := gtk.NewLabel("→")
		lblArrow.AddCSSClass("sidebar-arrow-label")
		btnArrow.SetChild(lblArrow)

		btnArrow.ConnectClicked(func() {
			if tw.TermInst != nil {
				tw.TermInst.FeedChild(cmdStr + "\n")
				tw.Revealer.SetRevealChild(false)
				tw.TermInst.Widget.GrabFocus()
			}
		})
		row.Append(btnArrow)

		tw.HistoryListBox.Append(row)
	}
}

func (tw *TerminalWindow) AddCommandToHistory(cmd string) {
	cmd = strings.TrimSpace(cmd)
	// We ignore commands that are too short (< 2 chars) or too long (> 60 chars)
	if len(cmd) < 2 || len(cmd) > 60 {
		return
	}

	// We ignore denylisted commands: ls, clear, top
	cmdLower := strings.ToLower(cmd)
	if cmdLower == "ls" || cmdLower == "clear" || cmdLower == "top" {
		return
	}

	// Don't add if it is the same as the most recent command
	if len(tw.Cfg.CommandHistory) > 0 && tw.Cfg.CommandHistory[0] == cmd {
		return
	}

	// Prepend to command history
	tw.Cfg.CommandHistory = append([]string{cmd}, tw.Cfg.CommandHistory...)

	// Truncate to maximum 10 items
	if len(tw.Cfg.CommandHistory) > 10 {
		tw.Cfg.CommandHistory = tw.Cfg.CommandHistory[:10]
	}

	// Save to config file
	_ = config.SaveConfig(tw.Cfg)
}

func (tw *TerminalWindow) setupContextMenu() {
	tw.popover = gtk.NewPopover()
	tw.popover.SetParent(tw.TermInst.Widget)
	tw.popover.SetHasArrow(true)

	box := gtk.NewBox(gtk.OrientationVertical, 2)

	// Copy Button
	tw.btnCopy = gtk.NewButton()
	tw.btnCopy.AddCSSClass("menu-item-btn")
	lblCopy := gtk.NewLabel("📋  Copy")
	lblCopy.AddCSSClass("menu-item-label")
	lblCopy.SetHAlign(gtk.AlignStart)
	tw.btnCopy.SetChild(lblCopy)
	tw.btnCopy.ConnectClicked(func() {
		if tw.TermInst != nil {
			tw.TermInst.Copy()
		}
		tw.popover.Popdown()
	})
	box.Append(tw.btnCopy)

	// Paste Button
	tw.btnPaste = gtk.NewButton()
	tw.btnPaste.AddCSSClass("menu-item-btn")
	lblPaste := gtk.NewLabel("📋  Paste")
	lblPaste.AddCSSClass("menu-item-label")
	lblPaste.SetHAlign(gtk.AlignStart)
	tw.btnPaste.SetChild(lblPaste)
	tw.btnPaste.ConnectClicked(func() {
		if tw.TermInst != nil {
			tw.TermInst.Paste()
		}
		tw.popover.Popdown()
	})
	box.Append(tw.btnPaste)

	tw.popover.SetChild(box)

	// Click gesture to intercept right clicks and selection releases
	clickGesture := gtk.NewGestureClick()
	clickGesture.SetButton(0) // Listen to all mouse buttons (1 = Left, 3 = Right)
	clickGesture.ConnectReleased(func(nPress int, x float64, y float64) {
		button := clickGesture.CurrentButton()
		hasSel := tw.TermInst.HasSelection()

		if button == 3 { // Right click: ALWAYS show appropriate menu
			tw.btnCopy.SetVisible(hasSel)
			tw.btnPaste.SetVisible(!hasSel)

			rect := gdk.NewRectangle(int(x), int(y), 1, 1)
			tw.popover.SetPointingTo(&rect)
			tw.popover.Popup()
		} else if button == 1 { // Left click release:
			// If left click is released and we have selection, show Copy!
			if hasSel {
				tw.btnCopy.SetVisible(true)
				tw.btnPaste.SetVisible(false)

				rect := gdk.NewRectangle(int(x), int(y), 1, 1)
				tw.popover.SetPointingTo(&rect)
				tw.popover.Popup()
			} else {
				// If left click and no selection, make sure popover is closed
				tw.popover.Popdown()
			}
		}
	})
	tw.TermInst.Widget.AddController(clickGesture)
}

package window

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/terminal"
	"zet-terminal/internal/theme"
)

type TerminalWindow struct {
	Win       *gtk.ApplicationWindow
	App       *gtk.Application
	Cfg       *config.Config
	TermInst  *terminal.VteTerminalInstance
	Container *gtk.Box
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

	// Show window
	win.Show()
	return tw
}

func (tw *TerminalWindow) setupUI() {
	// ─── Minimalistic Title Bar ───
	header := gtk.NewHeaderBar()
	titleLabel := gtk.NewLabel("Terminal")
	titleLabel.SetHAlign(gtk.AlignCenter)
	header.SetTitleWidget(titleLabel)
	tw.Win.SetTitlebar(header)

	// Single terminal instance
	tw.TermInst = terminal.NewVteTerminal()
	tw.TermInst.SetScrollbackLines(tw.Cfg.ScrollbackLines)
	tw.TermInst.SetCursorBlinkMode(tw.Cfg.CursorBlinkMode)
	tw.TermInst.SetCursorShape(tw.Cfg.CursorShape)

	// Container box
	tw.Container = gtk.NewBox(gtk.OrientationVertical, 0)
	tw.Container.AddCSSClass("terminal-container")
	tw.Container.SetHExpand(true)
	tw.Container.SetVExpand(true)
	tw.Container.Append(tw.TermInst.Widget)

	tw.Win.SetChild(tw.Container)

	// Apply fonts & theme colors
	tw.applyConfigToInstance(tw.TermInst)

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
}

func (tw *TerminalWindow) applyConfigToInstance(inst *terminal.VteTerminalInstance) {
	inst.SetFont(tw.Cfg.FontName, tw.Cfg.FontSize)
	inst.SetColors(theme.DefaultPalette.Foreground, theme.DefaultPalette.Background, theme.DefaultPalette.Palette)
}

func (tw *TerminalWindow) setupShortcuts() {
	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		isCtrl := (state & gdk.ControlMask) != 0
		isShift := (state & gdk.ShiftMask) != 0

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
			}
		}
		return false
	})
	tw.Win.AddController(keyCtrl)
}

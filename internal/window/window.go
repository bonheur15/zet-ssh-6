package window

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/terminal"
	"zet-terminal/internal/theme"
)

type TerminalWindow struct {
	Win           *gtk.ApplicationWindow
	App           *gtk.Application
	Cfg           *config.Config
	Notebook      *gtk.Notebook
	ActiveTabInst *terminal.VteTerminalInstance
}

func NewTerminalWindow(app *gtk.Application, cfg *config.Config) *TerminalWindow {
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("Terminal")
	win.SetDefaultSize(900, 600)
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

	// Buttons
	newTabBtn := gtk.NewButton()
	newTabBtn.SetLabel("New Tab")
	newTabBtn.AddCSSClass("header-btn")
	newTabBtn.ConnectClicked(func() {
		tw.NewTab("")
	})
	header.PackStart(newTabBtn)

	closeTabBtn := gtk.NewButton()
	closeTabBtn.SetLabel("Close Tab")
	closeTabBtn.AddCSSClass("header-btn")
	closeTabBtn.ConnectClicked(func() {
		tw.closeCurrentTab()
	})
	header.PackEnd(closeTabBtn)

	// Notebook (Tabs) taking full window space
	tw.Notebook = gtk.NewNotebook()
	tw.Notebook.SetHExpand(true)
	tw.Notebook.SetVExpand(true)
	tw.Win.SetChild(tw.Notebook)

	tw.Notebook.Connect("switch-page", func(nb *gtk.Notebook, page *gtk.Widget, pageNum uint) {
		tw.updateActiveTab(int(pageNum))
	})

	// Add the first tab
	tw.NewTab("")
}

func (tw *TerminalWindow) NewTab(workingDir string) {
	inst := terminal.NewVteTerminal()
	inst.SetScrollbackLines(tw.Cfg.ScrollbackLines)
	inst.SetCursorBlinkMode(tw.Cfg.CursorBlinkMode)
	inst.SetCursorShape(tw.Cfg.CursorShape)
	
	// Padded container box for tab
	termContainer := gtk.NewBox(gtk.OrientationVertical, 0)
	termContainer.AddCSSClass("terminal-container")
	termContainer.SetHExpand(true)
	termContainer.SetVExpand(true)
	termContainer.Append(inst.Widget)

	// Tab header label
	tabLabelBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	lbl := gtk.NewLabel("Terminal")
	tabLabelBox.Append(lbl)

	closeBtn := gtk.NewButton()
	closeBtn.SetLabel("x")
	closeBtn.ConnectClicked(func() {
		idx := tw.Notebook.PageNum(termContainer)
		if idx >= 0 {
			tw.Notebook.RemovePage(idx)
			inst.Destroy()
			if tw.Notebook.NPages() == 0 {
				tw.Win.Close()
			}
		}
	})
	tabLabelBox.Append(closeBtn)

	// Add to Notebook
	pageNum := tw.Notebook.AppendPage(termContainer, tabLabelBox)
	tw.Notebook.SetTabReorderable(termContainer, true)

	// Apply fonts & theme colors
	tw.applyConfigToInstance(inst)

	// Spawn default shell inside PTY
	inst.SpawnShell(tw.Cfg.Shell, workingDir)

	// Signal Handlers
	inst.OnChildExited(func(status int) {
		idx := tw.Notebook.PageNum(termContainer)
		if idx >= 0 {
			tw.Notebook.RemovePage(idx)
			inst.Destroy()
			if tw.Notebook.NPages() == 0 {
				tw.Win.Close()
			}
		}
	})

	inst.OnWindowTitleChanged(func(title string) {
		if title == "" {
			title = "Terminal"
		}
		lbl.SetLabel(title)
	})

	// Switch to new tab
	tw.Notebook.SetCurrentPage(pageNum)
}

func (tw *TerminalWindow) applyConfigToInstance(inst *terminal.VteTerminalInstance) {
	inst.SetFont(tw.Cfg.FontName, tw.Cfg.FontSize)
	inst.SetColors(theme.DefaultPalette.Foreground, theme.DefaultPalette.Background, theme.DefaultPalette.Palette)
}

func (tw *TerminalWindow) closeCurrentTab() {
	currentPage := tw.Notebook.CurrentPage()
	if currentPage >= 0 {
		termBox := tw.Notebook.NthPage(currentPage).(*gtk.Box)
		tw.Notebook.RemovePage(currentPage)
		
		termWidget := termBox.FirstChild()
		if termWidget != nil {
			for _, inst := range terminal.ActiveTerminals() {
				if inst.Widget.Object.Native() == termWidget.(*gtk.Widget).Object.Native() {
					inst.Destroy()
					break
				}
			}
		}
		if tw.Notebook.NPages() == 0 {
			tw.Win.Close()
		}
	}
}

func (tw *TerminalWindow) updateActiveTab(pageNum int) {
	page := tw.Notebook.NthPage(pageNum)
	if page == nil {
		return
	}
	termWidget := page.(*gtk.Box).FirstChild()
	if termWidget != nil {
		for _, inst := range terminal.ActiveTerminals() {
			if inst.Widget.Object.Native() == termWidget.(*gtk.Widget).Object.Native() {
				tw.ActiveTabInst = inst
				break
			}
		}
	}
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
			case 'T', 't': // Ctrl+Shift+T -> New Tab
				tw.NewTab("")
				return true
			case 'W', 'w': // Ctrl+Shift+W -> Close Current Tab
				tw.closeCurrentTab()
				return true
			case 'C', 'c': // Ctrl+Shift+C -> Copy
				if tw.ActiveTabInst != nil {
					tw.ActiveTabInst.Copy()
				}
				return true
			case 'V', 'v': // Ctrl+Shift+V -> Paste
				if tw.ActiveTabInst != nil {
					tw.ActiveTabInst.Paste()
				}
				return true
			}
		}
		return false
	})
	tw.Win.AddController(keyCtrl)
}

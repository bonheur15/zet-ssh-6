package window

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
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
	Sidebar       *gtk.Box
	SidebarToggle *gtk.Button
	ActiveTabInst *terminal.VteTerminalInstance
	
	// Sidebar UI elements
	CpuProgress   *gtk.ProgressBar
	CpuLabel      *gtk.Label
	RamProgress   *gtk.ProgressBar
	RamLabel      *gtk.Label
	HostValue     *gtk.Label
	ShellValue    *gtk.Label
	
	// Metrics state
	prevIdleTime  uint64
	prevTotalTime uint64
}

type TabContent struct {
	Box          *gtk.Box
	TermInstance *terminal.VteTerminalInstance
	Label        *gtk.Label
}

func NewTerminalWindow(app *gtk.Application, cfg *config.Config) *TerminalWindow {
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("ZET TERMINAL")
	win.SetDefaultSize(1000, 750)
	win.AddCSSClass("terminal-window")

	tw := &TerminalWindow{
		Win: win,
		App: app,
		Cfg: cfg,
	}

	tw.setupUI()
	tw.setupShortcuts()
	tw.startMetricsUpdater()

	// Show window
	win.Show()
	return tw
}

func (tw *TerminalWindow) setupUI() {
	// ─── Custom HeaderBar ───
	header := gtk.NewHeaderBar()
	titleLabel := gtk.NewLabel("⬡ ZET TERMINAL")
	titleLabel.SetHAlign(gtk.AlignCenter)
	header.SetTitleWidget(titleLabel)
	tw.Win.SetTitlebar(header)

	// Header Bar Buttons
	newTabBtn := gtk.NewButton()
	newTabBtn.SetLabel("➕ Tab")
	newTabBtn.AddCSSClass("header-btn")
	newTabBtn.ConnectClicked(func() {
		tw.NewTab("")
	})
	header.PackStart(newTabBtn)

	sidebarBtn := gtk.NewButton()
	sidebarBtn.SetLabel("📊 Dashboard")
	sidebarBtn.AddCSSClass("header-btn")
	sidebarBtn.ConnectClicked(func() {
		tw.ToggleSidebar()
	})
	header.PackEnd(sidebarBtn)

	// ─── Layout: Main Box (Horizontal) ───
	mainBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	tw.Win.SetChild(mainBox)

	// Layout: Notebook (Tabs) taking center/expanding space
	tw.Notebook = gtk.NewNotebook()
	tw.Notebook.SetHExpand(true)
	tw.Notebook.SetVExpand(true)
	mainBox.Append(tw.Notebook)

	tw.Notebook.Connect("switch-page", func(nb *gtk.Notebook, page *gtk.Widget, pageNum uint) {
		tw.updateActiveTab(int(pageNum))
	})

	// ─── Sidebar Setup ───
	tw.setupSidebar(mainBox)

	// Add the first initial tab
	tw.NewTab("")
}

func (tw *TerminalWindow) setupSidebar(parent *gtk.Box) {
	tw.Sidebar = gtk.NewBox(gtk.OrientationVertical, 16)
	tw.Sidebar.AddCSSClass("sidebar")
	tw.Sidebar.SetVExpand(true)
	parent.Append(tw.Sidebar)

	// Sidebar Title
	title := gtk.NewLabel("SYSTEM DASHBOARD")
	title.AddCSSClass("sidebar-title")
	title.SetHAlign(gtk.AlignStart)
	tw.Sidebar.Append(title)

	// CPU Monitor Card
	cpuCard := gtk.NewBox(gtk.OrientationVertical, 6)
	cpuCard.AddCSSClass("monitor-card")
	
	tw.CpuLabel = gtk.NewLabel("CPU: --%")
	tw.CpuLabel.AddCSSClass("monitor-value")
	tw.CpuLabel.SetHAlign(gtk.AlignStart)
	cpuCard.Append(tw.CpuLabel)
	
	tw.CpuProgress = gtk.NewProgressBar()
	tw.CpuProgress.AddCSSClass("sidebar-progress")
	tw.CpuProgress.SetFraction(0)
	cpuCard.Append(tw.CpuProgress)
	
	lbl1 := gtk.NewLabel("Processor Load")
	lbl1.AddCSSClass("monitor-label")
	lbl1.SetHAlign(gtk.AlignStart)
	cpuCard.Append(lbl1)
	tw.Sidebar.Append(cpuCard)

	// RAM Monitor Card
	ramCard := gtk.NewBox(gtk.OrientationVertical, 6)
	ramCard.AddCSSClass("monitor-card")
	
	tw.RamLabel = gtk.NewLabel("RAM: -- GB / -- GB")
	tw.RamLabel.AddCSSClass("monitor-value")
	tw.RamLabel.SetHAlign(gtk.AlignStart)
	ramCard.Append(tw.RamLabel)
	
	tw.RamProgress = gtk.NewProgressBar()
	tw.RamProgress.AddCSSClass("sidebar-progress")
	tw.RamProgress.SetFraction(0)
	ramCard.Append(tw.RamProgress)
	
	lbl2 := gtk.NewLabel("System Memory")
	lbl2.AddCSSClass("monitor-label")
	lbl2.SetHAlign(gtk.AlignStart)
	ramCard.Append(lbl2)
	tw.Sidebar.Append(ramCard)

	// Terminal info grid
	infoGrid := gtk.NewGrid()
	infoGrid.SetColumnSpacing(20)
	infoGrid.SetRowSpacing(8)

	lblH := gtk.NewLabel("HOSTNAME")
	lblH.AddCSSClass("monitor-label")
	infoGrid.Attach(lblH, 0, 0, 1, 1)

	hostname, _ := os.Hostname()
	tw.HostValue = gtk.NewLabel(hostname)
	tw.HostValue.AddCSSClass("monitor-value")
	tw.HostValue.SetHAlign(gtk.AlignStart)
	infoGrid.Attach(tw.HostValue, 1, 0, 1, 1)

	lblS := gtk.NewLabel("SHELL")
	lblS.AddCSSClass("monitor-label")
	infoGrid.Attach(lblS, 0, 1, 1, 1)

	tw.ShellValue = gtk.NewLabel(tw.Cfg.Shell)
	tw.ShellValue.AddCSSClass("monitor-value")
	tw.ShellValue.SetHAlign(gtk.AlignStart)
	infoGrid.Attach(tw.ShellValue, 1, 1, 1, 1)

	tw.Sidebar.Append(infoGrid)

	// Quick Settings Box
	settingsBox := gtk.NewBox(gtk.OrientationVertical, 8)
	settingsBox.SetMarginTop(12)

	// Theme selection buttons
	themeLabel := gtk.NewLabel("SELECT THEME")
	themeLabel.AddCSSClass("monitor-label")
	themeLabel.SetHAlign(gtk.AlignStart)
	settingsBox.Append(themeLabel)

	themes := []string{"cyberpunk", "tokyonight", "dracula", "nord"}
	for _, tName := range themes {
		btn := gtk.NewButton()
		btn.SetLabel(theme.Themes[tName].Name)
		btn.AddCSSClass("sidebar-btn")
		
		name := tName
		btn.ConnectClicked(func() {
			tw.ApplyTheme(name)
		})
		settingsBox.Append(btn)
	}

	tw.Sidebar.Append(settingsBox)
}

func (tw *TerminalWindow) NewTab(workingDir string) {
	inst := terminal.NewVteTerminal()
	inst.SetScrollbackLines(tw.Cfg.ScrollbackLines)
	inst.SetCursorBlinkMode(tw.Cfg.CursorBlinkMode)
	inst.SetCursorShape(tw.Cfg.CursorShape)
	
	// Create padded container box for tab
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
	closeBtn.SetLabel("×")
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

	// Apply active config
	tw.applyConfigToInstance(inst)

	// Spawn default shell inside VTE PTY
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

func (tw *TerminalWindow) ApplyTheme(themeName string) {
	tw.Cfg.Theme = themeName
	_ = config.SaveConfig(tw.Cfg)
	
	// Apply to all current tabs
	for i := 0; i < tw.Notebook.NPages(); i++ {
		page := tw.Notebook.NthPage(i)
		// Traverse custom activeTerminals to find matched instance
		termWidget := page.(*gtk.Box).FirstChild()
		if termWidget != nil {
			for _, inst := range terminal.ActiveTerminals() {
				if inst.Widget.Object.Native() == termWidget.(*gtk.Widget).Object.Native() {
					tw.applyConfigToInstance(inst)
				}
			}
		}
	}
}

func (tw *TerminalWindow) applyConfigToInstance(inst *terminal.VteTerminalInstance) {
	inst.SetFont(tw.Cfg.FontName, tw.Cfg.FontSize)
	
	palette, exists := theme.Themes[tw.Cfg.Theme]
	if !exists {
		palette = theme.Themes["cyberpunk"]
	}
	inst.SetColors(palette.Foreground, palette.Background, palette.Palette)
}

func (tw *TerminalWindow) ToggleSidebar() {
	visible := tw.Sidebar.IsVisible()
	tw.Sidebar.SetVisible(!visible)
	tw.Cfg.ShowSidebar = !visible
	_ = config.SaveConfig(tw.Cfg)
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
				// Emit action or call app callback to spawn new window
				tw.App.Activate() // Call activation to open a new window
				return true
			case 'T', 't': // Ctrl+Shift+T -> New Tab
				tw.NewTab("")
				return true
			case 'W', 'w': // Ctrl+Shift+W -> Close Tab
				currentPage := tw.Notebook.CurrentPage()
				if currentPage >= 0 {
					termBox := tw.Notebook.NthPage(currentPage).(*gtk.Box)
					tw.Notebook.RemovePage(currentPage)
					// Destroy matching instance
					termWidget := termBox.FirstChild()
					for _, inst := range terminal.ActiveTerminals() {
						if inst.Widget.Object.Native() == termWidget.(*gtk.Widget).Object.Native() {
							inst.Destroy()
						}
					}
					if tw.Notebook.NPages() == 0 {
						tw.Win.Close()
					}
				}
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

func (tw *TerminalWindow) startMetricsUpdater() {
	glib.TimeoutAdd(1000, func() bool {
		if !tw.Sidebar.IsVisible() {
			return true // Keep running, but don't compute if hidden
		}

		// 1. CPU Reading
		cpuPercent := tw.readCPU()
		tw.CpuLabel.SetLabel(fmt.Sprintf("CPU: %.1f%%", cpuPercent))
		tw.CpuProgress.SetFraction(cpuPercent / 100.0)

		// 2. RAM Reading
		usedRAM, totalRAM := tw.readRAM()
		if totalRAM > 0 {
			percentRAM := (usedRAM / totalRAM) * 100.0
			tw.RamLabel.SetLabel(fmt.Sprintf("RAM: %.1f / %.1f GB", usedRAM, totalRAM))
			tw.RamProgress.SetFraction(percentRAM / 100.0)
		}
		
		return true
	})
}

func (tw *TerminalWindow) readCPU() float64 {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0
	}
	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0
	}

	var total uint64
	var idle uint64
	for i, field := range fields {
		if i == 0 {
			continue
		}
		val, _ := strconv.ParseUint(field, 10, 64)
		total += val
		if i == 4 { // Idle time is 4th index (user, nice, system, idle...)
			idle = val
		}
	}

	idleDiff := idle - tw.prevIdleTime
	totalDiff := total - tw.prevTotalTime

	tw.prevIdleTime = idle
	tw.prevTotalTime = total

	if totalDiff == 0 {
		return 0
	}

	return 100.0 * float64(totalDiff-idleDiff) / float64(totalDiff)
}

func (tw *TerminalWindow) readRAM() (float64, float64) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	var total uint64
	var available uint64
	var free uint64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		if strings.HasPrefix(line, "MemTotal:") {
			total = val
		} else if strings.HasPrefix(line, "MemAvailable:") {
			available = val
		} else if strings.HasPrefix(line, "MemFree:") {
			free = val
		}
	}

	// MemAvailable is preferred on modern Linux, fallback to MemFree
	var used uint64
	if available > 0 {
		used = total - available
	} else {
		used = total - free
	}

	// Convert KB to GB
	return float64(used) / 1024.0 / 1024.0, float64(total) / 1024.0 / 1024.0
}

//go:build linux

package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const appCSS = `
/* ─── Global Reset & Dark Theme ─── */
window {
	background: linear-gradient(145deg, #0a0a0f 0%, #111128 40%, #0d1b2a 70%, #0a0a0f 100%);
	color: #e0e0e0;
}

/* ─── Animated Background Overlay ─── */
.bg-overlay {
	background: transparent;
}

/* ─── Header Bar ─── */
headerbar {
	background: linear-gradient(135deg, rgba(20, 20, 50, 0.92), rgba(15, 25, 45, 0.92));
	border-bottom: 1px solid rgba(100, 120, 255, 0.15);
	box-shadow: 0 2px 20px rgba(0, 0, 0, 0.5), 0 0 40px rgba(80, 100, 255, 0.05);
	min-height: 46px;
	padding: 0 12px;
}

headerbar title {
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 13px;
	font-weight: 600;
	color: rgba(180, 190, 255, 0.9);
	letter-spacing: 2px;
	text-transform: uppercase;
}

/* ─── Main Glass Card ─── */
.glass-card {
	background: linear-gradient(135deg, rgba(25, 25, 60, 0.55), rgba(20, 30, 55, 0.45));
	border: 1px solid rgba(100, 130, 255, 0.18);
	border-radius: 24px;
	box-shadow:
		0 8px 32px rgba(0, 0, 0, 0.4),
		0 0 60px rgba(80, 100, 255, 0.06),
		inset 0 1px 0 rgba(255, 255, 255, 0.05);
	padding: 48px 40px;
	margin: 24px;
}

/* ─── Hero Title ─── */
.hero-title {
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 48px;
	font-weight: 800;
	color: #ffffff;
	letter-spacing: -1px;
	margin-bottom: 8px;
}

/* ─── Accent Gradient Text ─── */
.gradient-text {
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 20px;
	font-weight: 500;
	color: #8b9cf7;
	letter-spacing: 0.5px;
}

/* ─── Status Pill ─── */
.status-pill {
	background: linear-gradient(135deg, rgba(40, 200, 120, 0.15), rgba(40, 200, 120, 0.08));
	border: 1px solid rgba(40, 200, 120, 0.3);
	border-radius: 100px;
	padding: 8px 20px;
	color: #5eeaa0;
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 12px;
	font-weight: 600;
	letter-spacing: 1.5px;
	text-transform: uppercase;
}

/* ─── Divider ─── */
.divider {
	background: linear-gradient(90deg, transparent, rgba(100, 130, 255, 0.25), transparent);
	min-height: 1px;
	margin: 16px 48px;
}

/* ─── Info Grid Labels ─── */
.info-label {
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 11px;
	font-weight: 600;
	color: rgba(140, 150, 200, 0.7);
	letter-spacing: 1.5px;
	text-transform: uppercase;
}

.info-value {
	font-family: 'JetBrains Mono', 'Fira Code', monospace;
	font-size: 14px;
	font-weight: 500;
	color: #c5ceff;
}

/* ─── Animated Orbs (via label hack) ─── */
.orb {
	border-radius: 50%;
	opacity: 0.12;
}

.orb-purple {
	background: radial-gradient(circle, #7c3aed, transparent 70%);
}

.orb-blue {
	background: radial-gradient(circle, #3b82f6, transparent 70%);
}

.orb-pink {
	background: radial-gradient(circle, #ec4899, transparent 70%);
}

/* ─── Particle dots ─── */
.particle {
	background: radial-gradient(circle, rgba(140, 160, 255, 0.8), transparent 70%);
	border-radius: 50%;
	min-width: 3px;
	min-height: 3px;
}

/* ─── Action Buttons ─── */
.action-btn {
	background: linear-gradient(135deg, rgba(80, 100, 255, 0.2), rgba(120, 80, 255, 0.15));
	border: 1px solid rgba(100, 120, 255, 0.25);
	border-radius: 14px;
	padding: 14px 32px;
	color: #a5b4fc;
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 14px;
	font-weight: 600;
	letter-spacing: 0.5px;
	transition: all 200ms ease;
	min-height: 20px;
}

.action-btn:hover {
	background: linear-gradient(135deg, rgba(80, 100, 255, 0.35), rgba(120, 80, 255, 0.3));
	border-color: rgba(100, 120, 255, 0.5);
	box-shadow: 0 4px 20px rgba(80, 100, 255, 0.2), 0 0 40px rgba(80, 100, 255, 0.1);
	color: #ffffff;
}

.action-btn:active {
	background: linear-gradient(135deg, rgba(80, 100, 255, 0.5), rgba(120, 80, 255, 0.45));
}

.primary-btn {
	background: linear-gradient(135deg, #5b6cf0, #7c3aed);
	border: 1px solid rgba(120, 100, 255, 0.4);
	color: #ffffff;
	box-shadow: 0 4px 15px rgba(91, 108, 240, 0.3);
}

.primary-btn:hover {
	background: linear-gradient(135deg, #6b7cf8, #8c4afd);
	box-shadow: 0 6px 25px rgba(91, 108, 240, 0.45), 0 0 50px rgba(91, 108, 240, 0.15);
}

/* ─── Metric Cards ─── */
.metric-card {
	background: linear-gradient(135deg, rgba(30, 30, 70, 0.5), rgba(25, 35, 60, 0.4));
	border: 1px solid rgba(100, 130, 255, 0.12);
	border-radius: 16px;
	padding: 20px 24px;
}

.metric-value {
	font-family: 'JetBrains Mono', 'Fira Code', monospace;
	font-size: 28px;
	font-weight: 700;
	color: #ffffff;
}

.metric-label {
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 11px;
	font-weight: 600;
	color: rgba(140, 150, 200, 0.6);
	letter-spacing: 1.5px;
	text-transform: uppercase;
}

/* ─── Pulse Ring Animation ─── */
.pulse-ring {
	border: 2px solid rgba(91, 108, 240, 0.3);
	border-radius: 50%;
	min-width: 12px;
	min-height: 12px;
}

.pulse-dot {
	background: #5b6cf0;
	border-radius: 50%;
	min-width: 8px;
	min-height: 8px;
	box-shadow: 0 0 12px rgba(91, 108, 240, 0.6);
}

/* ─── Footer ─── */
.footer-text {
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 11px;
	color: rgba(140, 150, 200, 0.35);
	letter-spacing: 0.5px;
}

/* ─── Scrollbar Styling ─── */
scrollbar {
	background: transparent;
}

scrollbar slider {
	background: rgba(100, 120, 255, 0.2);
	border-radius: 10px;
	min-width: 6px;
}

scrollbar slider:hover {
	background: rgba(100, 120, 255, 0.35);
}

/* ─── Toast / Notification ─── */
.toast {
	background: linear-gradient(135deg, rgba(40, 200, 120, 0.2), rgba(40, 200, 120, 0.1));
	border: 1px solid rgba(40, 200, 120, 0.3);
	border-radius: 12px;
	padding: 12px 24px;
	color: #5eeaa0;
	font-family: 'Inter', 'Cantarell', sans-serif;
	font-size: 13px;
	font-weight: 500;
}

/* ─── Progress Bar ─── */
.custom-progress {
	background: rgba(30, 30, 70, 0.5);
	border-radius: 8px;
	min-height: 6px;
}

.custom-progress trough {
	background: rgba(30, 30, 70, 0.5);
	border-radius: 8px;
	min-height: 6px;
}

.custom-progress progress {
	background: linear-gradient(90deg, #5b6cf0, #7c3aed, #ec4899);
	border-radius: 8px;
	min-height: 6px;
}

/* ─── Animated Gradient Border for main card ─── */
.glow-border {
	border-radius: 26px;
	padding: 2px;
	background: linear-gradient(135deg, rgba(91, 108, 240, 0.3), rgba(124, 58, 237, 0.2), rgba(236, 72, 153, 0.15), rgba(91, 108, 240, 0.3));
}
`

func main() {
	app := gtk.NewApplication("com.example.helloworld", 0)

	app.ConnectActivate(func() {
		activate(app)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	// Load CSS
	cssProvider := gtk.NewCSSProvider()
	cssProvider.LoadFromData(appCSS)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	// ─── Window Setup ───
	win := gtk.NewApplicationWindow(app)
	win.SetTitle("Hello World")
	win.SetDefaultSize(900, 700)

	// Custom header bar
	header := gtk.NewHeaderBar()
	titleLabel := gtk.NewLabel("⬡ HELLO WORLD")
	header.SetTitleWidget(titleLabel)
	win.SetTitlebar(header)

	// ─── Main Overlay for background effects ───
	overlay := gtk.NewOverlay()
	win.SetChild(overlay)

	// Background fixed container for floating orbs
	bgFixed := gtk.NewFixed()
	bgFixed.AddCSSClass("bg-overlay")
	bgFixed.SetHExpand(true)
	bgFixed.SetVExpand(true)
	overlay.SetChild(bgFixed)

	// Floating orbs are created below in the animation section

	// Floating particles
	particles := make([]*gtk.Label, 0, 30)
	for i := 0; i < 30; i++ {
		p := gtk.NewLabel("")
		p.AddCSSClass("particle")
		p.SetSizeRequest(2+rand.Intn(4), 2+rand.Intn(4))
		x := float64(rand.Intn(880))
		y := float64(rand.Intn(680))
		bgFixed.Put(p, x, y)
		particles = append(particles, p)
	}

	// ─── Main Content (scrollable) ───
	scrollWin := gtk.NewScrolledWindow()
	scrollWin.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrollWin.SetHExpand(true)
	scrollWin.SetVExpand(true)
	overlay.AddOverlay(scrollWin)

	// Content container
	content := gtk.NewBox(gtk.OrientationVertical, 0)
	content.SetHAlign(gtk.AlignCenter)
	content.SetVAlign(gtk.AlignCenter)
	content.SetHExpand(true)
	content.SetVExpand(true)
	content.SetMarginTop(20)
	content.SetMarginBottom(20)
	scrollWin.SetChild(content)

	// ─── Glow Border Wrapper ───
	glowBorder := gtk.NewBox(gtk.OrientationVertical, 0)
	glowBorder.AddCSSClass("glow-border")
	glowBorder.SetHAlign(gtk.AlignCenter)
	glowBorder.SetSizeRequest(680, -1)
	content.Append(glowBorder)

	// ─── Glass Card ───
	card := gtk.NewBox(gtk.OrientationVertical, 16)
	card.AddCSSClass("glass-card")
	card.SetHAlign(gtk.AlignFill)
	glowBorder.Append(card)

	// Status pill
	statusBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	statusBox.SetHAlign(gtk.AlignStart)

	pulseOverlay := gtk.NewOverlay()
	pulseRing := gtk.NewLabel("")
	pulseRing.AddCSSClass("pulse-ring")
	pulseDot := gtk.NewLabel("")
	pulseDot.AddCSSClass("pulse-dot")
	pulseDot.SetHAlign(gtk.AlignCenter)
	pulseDot.SetVAlign(gtk.AlignCenter)
	pulseOverlay.SetChild(pulseRing)
	pulseOverlay.AddOverlay(pulseDot)
	pulseOverlay.SetHAlign(gtk.AlignCenter)
	pulseOverlay.SetVAlign(gtk.AlignCenter)

	statusLabel := gtk.NewLabel("SYSTEM ACTIVE")
	statusLabel.AddCSSClass("status-pill")

	statusBox.Append(pulseOverlay)
	statusBox.Append(statusLabel)
	card.Append(statusBox)

	// Spacer
	spacer1 := gtk.NewBox(gtk.OrientationVertical, 0)
	spacer1.SetSizeRequest(-1, 8)
	card.Append(spacer1)

	// Hero title
	heroTitle := gtk.NewLabel("Hello, World!")
	heroTitle.AddCSSClass("hero-title")
	heroTitle.SetHAlign(gtk.AlignStart)
	card.Append(heroTitle)

	// Subtitle
	subtitle := gtk.NewLabel("A premium GTK4 experience, crafted with Go")
	subtitle.AddCSSClass("gradient-text")
	subtitle.SetHAlign(gtk.AlignStart)
	card.Append(subtitle)

	// Divider
	divider1 := gtk.NewBox(gtk.OrientationHorizontal, 0)
	divider1.AddCSSClass("divider")
	divider1.SetMarginTop(8)
	divider1.SetMarginBottom(8)
	card.Append(divider1)

	// ─── Metric Cards Row ───
	metricsRow := gtk.NewBox(gtk.OrientationHorizontal, 12)
	metricsRow.SetHomogeneous(true)

	uptimeValue := gtk.NewLabel("0s")
	uptimeValue.AddCSSClass("metric-value")
	createMetricCard(metricsRow, uptimeValue, "UPTIME")

	clickValue := gtk.NewLabel("0")
	clickValue.AddCSSClass("metric-value")
	createMetricCard(metricsRow, clickValue, "CLICKS")

	fpsValue := gtk.NewLabel("60")
	fpsValue.AddCSSClass("metric-value")
	createMetricCard(metricsRow, fpsValue, "FPS")

	card.Append(metricsRow)

	// ─── Progress Bar ───
	progressBox := gtk.NewBox(gtk.OrientationVertical, 6)
	progressBox.SetMarginTop(8)

	progressLabel := gtk.NewLabel("EXPERIENCE LOADING")
	progressLabel.AddCSSClass("info-label")
	progressLabel.SetHAlign(gtk.AlignStart)
	progressBox.Append(progressLabel)

	progressBar := gtk.NewProgressBar()
	progressBar.AddCSSClass("custom-progress")
	progressBar.SetFraction(0)
	progressBox.Append(progressBar)

	card.Append(progressBox)

	// Divider
	divider2 := gtk.NewBox(gtk.OrientationHorizontal, 0)
	divider2.AddCSSClass("divider")
	divider2.SetMarginTop(8)
	divider2.SetMarginBottom(4)
	card.Append(divider2)

	// ─── Info Grid ───
	infoGrid := gtk.NewGrid()
	infoGrid.SetColumnSpacing(40)
	infoGrid.SetRowSpacing(12)

	addInfoRow(infoGrid, 0, "RUNTIME", "Go 1.25 + GTK4")
	addInfoRow(infoGrid, 1, "PLATFORM", "Linux / Wayland")
	addInfoRow(infoGrid, 2, "TOOLKIT", "gotk4 v0.3.1")
	addInfoRow(infoGrid, 3, "RENDERER", "Native GPU")

	card.Append(infoGrid)

	// ─── Toast / Message Area ───
	toastLabel := gtk.NewLabel("")
	toastLabel.AddCSSClass("toast")
	toastLabel.SetVisible(false)
	toastLabel.SetMarginTop(12)
	card.Append(toastLabel)

	// ─── Action Buttons ───
	buttonRow := gtk.NewBox(gtk.OrientationHorizontal, 12)
	buttonRow.SetHAlign(gtk.AlignCenter)
	buttonRow.SetMarginTop(12)

	clickCount := 0

	greetBtn := gtk.NewButton()
	greetBtn.SetLabel("✦  Say Hello")
	greetBtn.AddCSSClass("action-btn")
	greetBtn.AddCSSClass("primary-btn")
	greetBtn.ConnectClicked(func() {
		clickCount++
		clickValue.SetLabel(fmt.Sprintf("%d", clickCount))

		greetings := []string{
			"Hello from the GTK4 universe! 🌌",
			"Greetings, fellow human! 🚀",
			"The void says hello back! ✨",
			"Go + GTK4 = Pure magic! 🎩",
			"Welcome to the dark side! 🌙",
			"Pixels have never looked better! 💎",
			"Your GPU is having a great time! 🎮",
		}
		msg := greetings[clickCount%len(greetings)]
		toastLabel.SetLabel(fmt.Sprintf("  %s  ", msg))
		toastLabel.SetVisible(true)

		// Auto-hide toast after 3 seconds
		glib.TimeoutAdd(3000, func() bool {
			toastLabel.SetVisible(false)
			return false
		})
	})
	buttonRow.Append(greetBtn)

	themeBtn := gtk.NewButton()
	themeBtn.SetLabel("◐  Pulse")
	themeBtn.AddCSSClass("action-btn")
	themeBtn.ConnectClicked(func() {
		clickCount++
		clickValue.SetLabel(fmt.Sprintf("%d", clickCount))
		toastLabel.SetLabel("  ✧ Pulse sent through the matrix ✧  ")
		toastLabel.SetVisible(true)
		glib.TimeoutAdd(2000, func() bool {
			toastLabel.SetVisible(false)
			return false
		})
	})
	buttonRow.Append(themeBtn)

	infoBtn := gtk.NewButton()
	infoBtn.SetLabel("⚙  System")
	infoBtn.AddCSSClass("action-btn")
	infoBtn.ConnectClicked(func() {
		clickCount++
		clickValue.SetLabel(fmt.Sprintf("%d", clickCount))
		hostname, _ := os.Hostname()
		toastLabel.SetLabel(fmt.Sprintf("  Host: %s  |  PID: %d  ", hostname, os.Getpid()))
		toastLabel.SetVisible(true)
		glib.TimeoutAdd(4000, func() bool {
			toastLabel.SetVisible(false)
			return false
		})
	})
	buttonRow.Append(infoBtn)

	card.Append(buttonRow)

	// ─── Footer ───
	footer := gtk.NewLabel("Built with Go + GTK4  •  Linux Native  •  2026")
	footer.AddCSSClass("footer-text")
	footer.SetMarginTop(24)
	footer.SetMarginBottom(8)
	content.Append(footer)

	// ─── Animation Timers ───
	startTime := time.Now()

	// Uptime counter
	glib.TimeoutAdd(1000, func() bool {
		elapsed := time.Since(startTime)
		if elapsed.Hours() >= 1 {
			uptimeValue.SetLabel(fmt.Sprintf("%.0fh%dm", elapsed.Hours(), int(elapsed.Minutes())%60))
		} else if elapsed.Minutes() >= 1 {
			uptimeValue.SetLabel(fmt.Sprintf("%dm%ds", int(elapsed.Minutes()), int(elapsed.Seconds())%60))
		} else {
			uptimeValue.SetLabel(fmt.Sprintf("%ds", int(elapsed.Seconds())))
		}
		return true // keep running
	})

	// Progress bar animation
	glib.TimeoutAdd(50, func() bool {
		current := progressBar.Fraction()
		if current < 1.0 {
			progressBar.SetFraction(current + 0.005)
			return true
		}
		progressBar.SetFraction(1.0)
		progressLabel.SetLabel("EXPERIENCE LOADED")
		return false
	})

	// Floating particle animation
	type particleState struct {
		x, y   float64
		vx, vy float64
	}
	pStates := make([]particleState, len(particles))
	for i := range pStates {
		pStates[i] = particleState{
			x:  float64(rand.Intn(860)) + 10,
			y:  float64(rand.Intn(660)) + 10,
			vx: (rand.Float64() - 0.5) * 1.2,
			vy: (rand.Float64() - 0.5) * 1.2,
		}
	}

	glib.TimeoutAdd(33, func() bool { // ~30fps particle update
		for i, p := range particles {
			s := &pStates[i]
			s.x += s.vx
			s.y += s.vy

			// Gentle sine wave drift
			t := float64(time.Since(startTime).Milliseconds()) / 1000.0
			s.x += math.Sin(t*0.5+float64(i)*0.3) * 0.3
			s.y += math.Cos(t*0.4+float64(i)*0.2) * 0.3

			// Boundary wrapping
			if s.x < -10 {
				s.x = 890
			} else if s.x > 900 {
				s.x = -5
			}
			if s.y < -10 {
				s.y = 690
			} else if s.y > 700 {
				s.y = -5
			}

			bgFixed.Move(p, s.x, s.y)
		}
		return true
	})

	// Floating orb gentle drift
	type orbInfo struct {
		widget       *gtk.Label
		baseX, baseY float64
	}
	orbs := []orbInfo{
		{createOrbWidget(bgFixed, "orb-purple", 80, 60, 300), 80, 60},
		{createOrbWidget(bgFixed, "orb-blue", 550, 350, 250), 550, 350},
		{createOrbWidget(bgFixed, "orb-pink", 300, 500, 200), 300, 500},
	}
	glib.TimeoutAdd(50, func() bool {
		t := float64(time.Since(startTime).Milliseconds()) / 1000.0
		for i, orb := range orbs {
			dx := math.Sin(t*0.3+float64(i)*2.1) * 30
			dy := math.Cos(t*0.25+float64(i)*1.7) * 25
			bgFixed.Move(orb.widget, orb.baseX+dx, orb.baseY+dy)
		}
		return true
	})

	// FPS counter (approximate)
	frameCount := 0
	lastFPSUpdate := time.Now()
	glib.TimeoutAdd(16, func() bool { // ~60fps tick
		frameCount++
		if time.Since(lastFPSUpdate) >= time.Second {
			fpsValue.SetLabel(fmt.Sprintf("%d", frameCount))
			frameCount = 0
			lastFPSUpdate = time.Now()
		}
		return true
	})

	win.Show()
}

func createOrbWidget(fixed *gtk.Fixed, cssClass string, x, y float64, size int) *gtk.Label {
	orb := gtk.NewLabel("")
	orb.AddCSSClass("orb")
	orb.AddCSSClass(cssClass)
	orb.SetSizeRequest(size, size)
	fixed.Put(orb, x, y)
	return orb
}

func createMetricCard(parent *gtk.Box, valueLabel *gtk.Label, label string) {
	card := gtk.NewBox(gtk.OrientationVertical, 4)
	card.AddCSSClass("metric-card")
	card.SetHAlign(gtk.AlignFill)

	valueLabel.SetHAlign(gtk.AlignStart)
	card.Append(valueLabel)

	lbl := gtk.NewLabel(label)
	lbl.AddCSSClass("metric-label")
	lbl.SetHAlign(gtk.AlignStart)
	card.Append(lbl)

	parent.Append(card)
}

func addInfoRow(grid *gtk.Grid, row int, label, value string) {
	lbl := gtk.NewLabel(label)
	lbl.AddCSSClass("info-label")
	lbl.SetHAlign(gtk.AlignStart)
	grid.Attach(lbl, 0, row, 1, 1)

	val := gtk.NewLabel(value)
	val.AddCSSClass("info-value")
	val.SetHAlign(gtk.AlignStart)
	grid.Attach(val, 1, row, 1, 1)
}

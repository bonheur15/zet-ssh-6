package window

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"zet-terminal/internal/sshmgr"
)

type paletteItem struct {
	title    string
	subtitle string
	kind     string // host | tunnel | snippet | action
	action   func()
}

// openQuickConnect shows a fuzzy command palette over hosts, tunnels,
// snippets, and quick actions.
func (tw *TerminalWindow) openQuickConnect() {
	dialog := gtk.NewWindow()
	dialog.SetTitle("Quick Connect")
	dialog.SetTransientFor(&tw.Win.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(560, 460)
	dialog.AddCSSClass("palette-dialog")

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.AddCSSClass("palette-box")
	dialog.SetChild(box)

	search := gtk.NewEntry()
	search.AddCSSClass("palette-search")
	search.SetPlaceholderText("Connect, run a tunnel or snippet…")
	box.Append(search)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	box.Append(scrolled)

	listBox := gtk.NewBox(gtk.OrientationVertical, 2)
	scrolled.SetChild(listBox)

	items := tw.buildPaletteItems()

	var rows []*gtk.Button
	selected := 0

	activate := func(idx int) {
		filtered := filterPalette(items, search.Text())
		if idx < 0 || idx >= len(filtered) {
			return
		}
		dialog.Close()
		filtered[idx].action()
	}

	var rebuild func()
	rebuild = func() {
		for child := listBox.FirstChild(); child != nil; child = listBox.FirstChild() {
			listBox.Remove(child)
		}
		rows = nil
		filtered := filterPalette(items, search.Text())
		if selected >= len(filtered) {
			selected = len(filtered) - 1
		}
		if selected < 0 {
			selected = 0
		}
		if len(filtered) == 0 {
			empty := gtk.NewLabel("No matches.")
			empty.AddCSSClass("sidebar-btn-label")
			empty.SetMarginTop(20)
			listBox.Append(empty)
			return
		}
		for i, it := range filtered {
			item := it
			idx := i
			btn := gtk.NewButton()
			btn.AddCSSClass("palette-item")
			if i == selected {
				btn.AddCSSClass("active")
			}
			btn.SetHExpand(true)

			rowBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
			badge := gtk.NewLabel(paletteBadge(item.kind))
			badge.AddCSSClass("palette-badge")
			badge.AddCSSClass("badge-" + item.kind)
			rowBox.Append(badge)

			textBox := gtk.NewBox(gtk.OrientationVertical, 0)
			textBox.SetHExpand(true)
			title := gtk.NewLabel(item.title)
			title.AddCSSClass("palette-title")
			title.SetHAlign(gtk.AlignStart)
			title.SetXAlign(0)
			title.SetEllipsize(pango.EllipsizeEnd)
			textBox.Append(title)
			if item.subtitle != "" {
				sub := gtk.NewLabel(item.subtitle)
				sub.AddCSSClass("palette-subtitle")
				sub.SetHAlign(gtk.AlignStart)
				sub.SetXAlign(0)
				sub.SetEllipsize(pango.EllipsizeEnd)
				textBox.Append(sub)
			}
			rowBox.Append(textBox)
			btn.SetChild(rowBox)
			btn.ConnectClicked(func() { activate(idx) })
			listBox.Append(btn)
			rows = append(rows, btn)
		}
	}
	rebuild()

	search.ConnectChanged(func() {
		selected = 0
		rebuild()
	})

	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.SetPropagationPhase(gtk.PhaseCapture)
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		switch keyval {
		case gdk.KEY_Escape:
			dialog.Close()
			return true
		case gdk.KEY_Down:
			selected++
			rebuild()
			return true
		case gdk.KEY_Up:
			selected--
			rebuild()
			return true
		case gdk.KEY_Return, gdk.KEY_KP_Enter:
			activate(selected)
			return true
		}
		return false
	})
	dialog.AddController(keyCtrl)

	dialog.Present()
	glib.IdleAdd(func() { search.GrabFocus() })
}

func (tw *TerminalWindow) buildPaletteItems() []paletteItem {
	var items []paletteItem

	for _, prof := range profileStore().Sorted() {
		p := prof
		items = append(items, paletteItem{
			title:    p.Label(),
			subtitle: p.Target(),
			kind:     "host",
			action:   func() { tw.ConnectProfile(p) },
		})
	}
	for _, tun := range tunnelStore().Tunnels {
		t := tun
		label := tunnelLabel(t)
		items = append(items, paletteItem{
			title:    label,
			subtitle: t.Summary(),
			kind:     "tunnel",
			action:   func() { tw.toggleTunnel(t) },
		})
	}
	for _, snip := range snippetStore().Search("") {
		sn := snip
		items = append(items, paletteItem{
			title:    sn.Name,
			subtitle: sn.Command,
			kind:     "snippet",
			action:   func() { tw.RunSnippet(sn, true) },
		})
	}
	// Quick actions
	items = append(items,
		paletteItem{title: "New SSH Host", kind: "action", action: func() { tw.openProfileDialog(nil, nil, nil) }},
		paletteItem{title: "New Tunnel", kind: "action", action: func() { tw.openTunnelDialog(nil, nil) }},
		paletteItem{title: "New Snippet", kind: "action", action: func() { tw.openSnippetDialog(nil, nil) }},
		paletteItem{title: "Open Command Log", kind: "action", action: func() { tw.openCommandLog() }},
	)
	return items
}

func (tw *TerminalWindow) toggleTunnel(t *sshmgr.TunnelSpec) {
	store := tunnelStore()
	if store.IsRunning(t.ID) {
		store.Stop(t.ID)
	} else {
		tw.StartTunnel(t)
	}
}

func paletteBadge(kind string) string {
	switch kind {
	case "host":
		return "SSH"
	case "tunnel":
		return "TUN"
	case "snippet":
		return "CMD"
	default:
		return "•"
	}
}

// filterPalette does a simple subsequence-tolerant substring match.
func filterPalette(items []paletteItem, query string) []paletteItem {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return items
	}
	var out []paletteItem
	for _, it := range items {
		hay := strings.ToLower(it.title + " " + it.subtitle + " " + it.kind)
		if fuzzyMatch(hay, query) {
			out = append(out, it)
		}
	}
	return out
}

func fuzzyMatch(haystack, needle string) bool {
	if strings.Contains(haystack, needle) {
		return true
	}
	// subsequence match: all needle chars appear in order
	i := 0
	for _, r := range haystack {
		if i < len(needle) && rune(needle[i]) == r {
			i++
		}
	}
	return i == len(needle)
}

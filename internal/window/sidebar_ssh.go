package window

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"zet-terminal/internal/sshmgr"
)

// sectionActionButton builds the full-width "+ New …" button.
func sectionActionButton(label string, onClick func()) *gtk.Button {
	btn := gtk.NewButton()
	btn.AddCSSClass("workspace-action-btn")
	lbl := gtk.NewLabel(label)
	lbl.AddCSSClass("workspace-action-label")
	btn.SetChild(lbl)
	btn.ConnectClicked(onClick)
	return btn
}

// sidebarActionIcon builds a small trailing action icon button for a row.
func sidebarActionIcon(icon, tip string, extraClass string, onClick func()) *gtk.Button {
	btn := gtk.NewButton()
	btn.AddCSSClass("tab-action-btn")
	if extraClass != "" {
		btn.AddCSSClass(extraClass)
	}
	btn.SetTooltipText(tip)
	btn.SetChild(gtk.NewImageFromIconName(icon))
	btn.ConnectClicked(onClick)
	return btn
}

func sidebarEmptyLabel(text string) *gtk.Label {
	lbl := gtk.NewLabel(text)
	lbl.AddCSSClass("sidebar-btn-label")
	lbl.SetMarginTop(16)
	lbl.SetHAlign(gtk.AlignCenter)
	lbl.SetWrap(true)
	return lbl
}

func (tw *TerminalWindow) renderHostsSection() {
	tw.WorkspaceBox.Append(sectionActionButton("+ New Host", func() {
		tw.openProfileDialog(nil, nil, nil)
	}))

	profiles := profileStore().Sorted()
	if len(profiles) == 0 {
		tw.WorkspaceBox.Append(sidebarEmptyLabel("No saved hosts yet.\nAdd one, or paste an ssh command in a terminal."))
		return
	}

	for _, prof := range profiles {
		p := prof
		row := gtk.NewBox(gtk.OrientationHorizontal, 4)
		row.AddCSSClass("tab-row")
		row.SetHExpand(true)

		connectBtn := gtk.NewButton()
		connectBtn.AddCSSClass("tab-select-btn")
		connectBtn.SetHExpand(true)
		connectBtn.SetHAlign(gtk.AlignFill)
		connectBtn.SetTooltipText("Connect — " + p.Target())

		labelBox := gtk.NewBox(gtk.OrientationVertical, 0)
		name := gtk.NewLabel(p.Label())
		name.AddCSSClass("tab-label")
		name.SetHAlign(gtk.AlignStart)
		name.SetXAlign(0)
		name.SetEllipsize(pango.EllipsizeEnd)
		name.SetMaxWidthChars(24)
		labelBox.Append(name)
		sub := gtk.NewLabel(p.Target())
		sub.AddCSSClass("host-sub-label")
		sub.SetHAlign(gtk.AlignStart)
		sub.SetXAlign(0)
		sub.SetEllipsize(pango.EllipsizeEnd)
		sub.SetMaxWidthChars(28)
		labelBox.Append(sub)
		connectBtn.SetChild(labelBox)
		connectBtn.ConnectClicked(func() {
			tw.ConnectProfile(p)
		})
		row.Append(connectBtn)

		row.Append(sidebarActionIcon("folder-remote-symbolic", "Browse files", "", func() {
			tw.openFileBrowser(p)
		}))
		row.Append(sidebarActionIcon("document-send-symbolic", "Transfer files (scp)", "", func() {
			tw.openTransferDialog(p)
		}))
		row.Append(sidebarActionIcon("document-edit-symbolic", "Edit host", "", func() {
			tw.openProfileDialog(p, nil, nil)
		}))
		row.Append(sidebarActionIcon("user-trash-symbolic", "Delete host", "tab-close-btn", func() {
			tw.confirm("Delete host", "Delete profile \""+p.Label()+"\"? This cannot be undone.", func() {
				profileStore().Delete(p.ID)
				tw.renderWorkspace()
			})
		}))

		tw.WorkspaceBox.Append(row)
	}
}

func (tw *TerminalWindow) renderTunnelsSection() {
	tw.WorkspaceBox.Append(sectionActionButton("+ New Tunnel", func() {
		tw.openTunnelDialog(nil, nil)
	}))

	store := tunnelStore()
	if len(store.Tunnels) == 0 {
		tw.WorkspaceBox.Append(sidebarEmptyLabel("No tunnels yet.\nForward a port through one of your hosts."))
		return
	}

	for _, tun := range store.Tunnels {
		t := tun
		running := store.IsRunning(t.ID)

		row := gtk.NewBox(gtk.OrientationHorizontal, 4)
		row.AddCSSClass("tab-row")
		if running {
			row.AddCSSClass("active")
		}
		row.SetHExpand(true)

		labelBox := gtk.NewBox(gtk.OrientationVertical, 0)
		name := gtk.NewLabel(tunnelLabel(t))
		name.AddCSSClass("tab-label")
		name.SetHAlign(gtk.AlignStart)
		name.SetXAlign(0)
		name.SetEllipsize(pango.EllipsizeEnd)
		name.SetMaxWidthChars(22)
		labelBox.Append(name)
		status := t.Summary()
		if running {
			status = "● running   " + status
		}
		sub := gtk.NewLabel(status)
		sub.AddCSSClass("host-sub-label")
		sub.SetHAlign(gtk.AlignStart)
		sub.SetXAlign(0)
		sub.SetEllipsize(pango.EllipsizeEnd)
		sub.SetMaxWidthChars(28)
		labelBox.Append(sub)

		infoBtn := gtk.NewButton()
		infoBtn.AddCSSClass("tab-select-btn")
		infoBtn.SetHExpand(true)
		infoBtn.SetHAlign(gtk.AlignFill)
		infoBtn.SetChild(labelBox)
		if p := profileStore().Get(t.ProfileID); p != nil {
			infoBtn.SetTooltipText(sshmgr.CommandString(t.Argv(p)))
		}
		infoBtn.ConnectClicked(func() {
			if store.IsRunning(t.ID) {
				store.Stop(t.ID)
			} else {
				tw.StartTunnel(t)
			}
		})
		row.Append(infoBtn)

		if running {
			row.Append(sidebarActionIcon("media-playback-stop-symbolic", "Stop tunnel", "tab-close-btn", func() {
				store.Stop(t.ID)
			}))
		} else {
			row.Append(sidebarActionIcon("media-playback-start-symbolic", "Start tunnel", "", func() {
				tw.StartTunnel(t)
			}))
		}
		row.Append(sidebarActionIcon("document-edit-symbolic", "Edit tunnel", "", func() {
			tw.openTunnelDialog(t, nil)
		}))
		row.Append(sidebarActionIcon("user-trash-symbolic", "Delete tunnel", "tab-close-btn", func() {
			tw.confirm("Delete tunnel", "Delete tunnel \""+tunnelLabel(t)+"\"?", func() {
				store.Delete(t.ID)
				tw.renderWorkspace()
			})
		}))

		tw.WorkspaceBox.Append(row)
	}
}

func (tw *TerminalWindow) renderSnippetsSection() {
	tw.WorkspaceBox.Append(sectionActionButton("+ New Snippet", func() {
		tw.openSnippetDialog(nil, nil)
	}))

	list := snippetStore().Search("")
	if len(list) == 0 {
		tw.WorkspaceBox.Append(sidebarEmptyLabel("No snippets yet.\nSave commands with ${VAR} placeholders."))
		return
	}

	for _, snip := range list {
		sn := snip
		row := gtk.NewBox(gtk.OrientationHorizontal, 4)
		row.AddCSSClass("tab-row")
		row.SetHExpand(true)

		runBtn := gtk.NewButton()
		runBtn.AddCSSClass("tab-select-btn")
		runBtn.SetHExpand(true)
		runBtn.SetHAlign(gtk.AlignFill)
		runBtn.SetTooltipText(sn.Command)

		labelBox := gtk.NewBox(gtk.OrientationVertical, 0)
		name := gtk.NewLabel(sn.Name)
		name.AddCSSClass("tab-label")
		name.SetHAlign(gtk.AlignStart)
		name.SetXAlign(0)
		name.SetEllipsize(pango.EllipsizeEnd)
		name.SetMaxWidthChars(24)
		labelBox.Append(name)
		sub := gtk.NewLabel(sn.Command)
		sub.AddCSSClass("host-sub-label")
		sub.SetHAlign(gtk.AlignStart)
		sub.SetXAlign(0)
		sub.SetEllipsize(pango.EllipsizeEnd)
		sub.SetMaxWidthChars(28)
		labelBox.Append(sub)
		runBtn.SetChild(labelBox)
		runBtn.ConnectClicked(func() {
			tw.RunSnippet(sn, true)
		})
		row.Append(runBtn)

		row.Append(sidebarActionIcon("input-keyboard-symbolic", "Insert without running", "", func() {
			tw.RunSnippet(sn, false)
		}))
		row.Append(sidebarActionIcon("document-edit-symbolic", "Edit snippet", "", func() {
			tw.openSnippetDialog(sn, nil)
		}))
		row.Append(sidebarActionIcon("user-trash-symbolic", "Delete snippet", "tab-close-btn", func() {
			tw.confirm("Delete snippet", "Delete snippet \""+sn.Name+"\"?", func() {
				snippetStore().Delete(sn.ID)
				tw.renderWorkspace()
			})
		}))

		tw.WorkspaceBox.Append(row)
	}
}

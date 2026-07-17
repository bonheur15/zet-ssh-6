package window

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/snippets"
	"zet-terminal/internal/vault"
)

func snippetStore() *snippets.Store {
	return snippets.Load(cfgDir())
}

// openSnippetDialog edits sn, or creates a new snippet when sn is nil.
func (tw *TerminalWindow) openSnippetDialog(sn *snippets.Snippet, onSaved func()) {
	editing := sn != nil
	if sn == nil {
		sn = &snippets.Snippet{}
	}

	titleText := "New Snippet"
	if editing {
		titleText = "Edit Snippet"
	}
	dialog, box := tw.newDialog(titleText, 560, 520)
	dialogTitle(box, titleText, "Save a reusable command. Use ${VAR} placeholders — they're filled in when you run it, and can be backed by the vault.")

	content := scrollArea(box)

	entryName := labeledEntry(content, "Name:", sn.Name, "restart nginx")
	entryDesc := labeledEntry(content, "Description:", sn.Description, "optional")
	entryTags := labeledEntry(content, "Tags (comma sep):", strings.Join(sn.Tags, ", "), "ops, web")

	cmdLbl := gtk.NewLabel("Command")
	cmdLbl.AddCSSClass("settings-label")
	cmdLbl.SetHAlign(gtk.AlignStart)
	cmdLbl.SetMarginTop(4)
	content.Append(cmdLbl)

	textView := gtk.NewTextView()
	textView.AddCSSClass("snippet-editor")
	textView.SetWrapMode(gtk.WrapWordChar)
	textView.SetSizeRequest(-1, 120)
	buf := textView.Buffer()
	buf.SetText(sn.Command)
	frame := gtk.NewScrolledWindow()
	frame.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	frame.SetSizeRequest(-1, 130)
	frame.SetChild(textView)
	content.Append(frame)

	varsHint := gtk.NewLabel("")
	varsHint.AddCSSClass("settings-hint")
	varsHint.SetHAlign(gtk.AlignStart)
	varsHint.SetWrap(true)
	content.Append(varsHint)

	getCommand := func() string {
		start, end := buf.Bounds()
		return buf.Text(start, end, false)
	}
	refresh := func() {
		tmp := &snippets.Snippet{Command: getCommand()}
		names := tmp.VarNames()
		if len(names) == 0 {
			varsHint.SetLabel("No variables.")
		} else {
			varsHint.SetLabel("Variables: ${" + strings.Join(names, "}, ${") + "}")
		}
	}
	buf.ConnectChanged(refresh)
	refresh()

	dialogButtons(box, dialog, "Save Snippet", func() {
		sn.Name = strings.TrimSpace(entryName.Text())
		sn.Description = strings.TrimSpace(entryDesc.Text())
		sn.Command = strings.TrimSpace(getCommand())
		sn.Tags = nil
		for _, t := range strings.Split(entryTags.Text(), ",") {
			if t = strings.TrimSpace(t); t != "" {
				sn.Tags = append(sn.Tags, t)
			}
		}
		if sn.Name == "" || sn.Command == "" {
			entryName.GrabFocus()
			return
		}
		snippetStore().Upsert(sn)
		dialog.Close()
		if onSaved != nil {
			onSaved()
		}
		tw.renderWorkspace()
	})

	dialog.Present()
}

// RunSnippet expands a snippet (prompting for any missing variables) and
// runs it in the active terminal. If runNow is false it only inserts.
func (tw *TerminalWindow) RunSnippet(sn *snippets.Snippet, runNow bool) {
	resolve := func(name string) (string, bool) {
		v := appVault()
		if v.IsUnlocked() {
			if val, err := v.Get(vault.SnippetVarKey(name)); err == nil && val != "" {
				return val, true
			}
		}
		return "", false
	}
	expanded, missing := sn.Expand(resolve)
	if len(missing) == 0 {
		tw.dispatchSnippet(expanded, runNow)
		return
	}
	tw.promptSnippetVars(sn, missing, runNow)
}

func (tw *TerminalWindow) dispatchSnippet(command string, runNow bool) {
	if runNow {
		tw.RunInActiveTab(command)
	} else {
		tw.InsertIntoActiveTab(command)
	}
}

// promptSnippetVars asks for values for the missing ${VAR} placeholders.
func (tw *TerminalWindow) promptSnippetVars(sn *snippets.Snippet, missing []string, runNow bool) {
	dialog, box := tw.newDialog("Fill Variables", 460, -1)
	dialogTitle(box, "Fill Variables", "Provide values for the placeholders in \""+sn.Name+"\".")

	entries := map[string]*gtk.Entry{}
	saveToVault := map[string]*gtk.CheckButton{}
	for _, name := range missing {
		entries[name] = labeledEntry(box, name+":", "", "")
		chk := gtk.NewCheckButtonWithLabel("Remember in vault as ${" + name + "}")
		chk.AddCSSClass("settings-hint")
		box.Append(chk)
		saveToVault[name] = chk
	}

	dialogButtons(box, dialog, "Run", func() {
		values := map[string]string{}
		for name, entry := range entries {
			values[name] = entry.Text()
		}
		// Persist any "remember" values to the vault.
		toSave := map[string]string{}
		for name, chk := range saveToVault {
			if chk.Active() && values[name] != "" {
				toSave[name] = values[name]
			}
		}
		finish := func() {
			expanded, _ := sn.Expand(func(n string) (string, bool) {
				if v, ok := values[n]; ok {
					return v, true
				}
				return "", false
			})
			tw.dispatchSnippet(expanded, runNow)
		}
		if len(toSave) > 0 {
			tw.WithVault(func(v *vault.Vault) {
				for n, val := range toSave {
					_ = v.Set(vault.SnippetVarKey(n), val)
				}
				finish()
			})
		} else {
			finish()
		}
		dialog.Close()
	})

	dialog.Present()
}

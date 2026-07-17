package window

import (
	"sync"
	"time"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"zet-terminal/internal/config"
	"zet-terminal/internal/vault"
)

func appVault() *vault.Vault {
	return vault.Global(config.GetConfigDir())
}

var autoLockOnce sync.Once

// startVaultAutoLock arms a periodic idle check that locks the vault
// after the configured number of minutes without use.
func startVaultAutoLock() {
	autoLockOnce.Do(func() {
		glib.TimeoutSecondsAdd(30, func() bool {
			tw := GetFirstActiveWindow()
			if tw == nil {
				return true
			}
			mins := tw.Cfg.VaultAutoLockMin
			if mins > 0 {
				appVault().MaybeAutoLock(time.Duration(mins) * time.Minute)
			}
			return true
		})
	})
}

// WithVault runs onUnlocked once the vault is usable, prompting for the
// master password (or first-time creation) if needed.
func (tw *TerminalWindow) WithVault(onUnlocked func(v *vault.Vault)) {
	v := appVault()
	if v.IsUnlocked() {
		onUnlocked(v)
		return
	}
	tw.openVaultDialog(onUnlocked)
}

func (tw *TerminalWindow) openVaultDialog(onUnlocked func(v *vault.Vault)) {
	v := appVault()
	creating := !v.Exists()

	dialog := gtk.NewWindow()
	if creating {
		dialog.SetTitle("Create Vault")
	} else {
		dialog.SetTitle("Unlock Vault")
	}
	dialog.SetTransientFor(&tw.Win.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(420, -1)
	dialog.AddCSSClass("settings-dialog")

	box := gtk.NewBox(gtk.OrientationVertical, 12)
	box.AddCSSClass("settings-box")
	dialog.SetChild(box)

	title := gtk.NewLabel("Encrypted Vault")
	title.AddCSSClass("settings-page-title")
	title.SetHAlign(gtk.AlignStart)
	box.Append(title)

	subText := "Enter your master password to unlock secrets."
	if creating {
		subText = "Choose a master password. It encrypts passwords, key passphrases, and snippet secrets locally (Argon2id + XChaCha20-Poly1305). It cannot be recovered if lost."
	}
	sub := gtk.NewLabel(subText)
	sub.AddCSSClass("settings-page-subtitle")
	sub.SetHAlign(gtk.AlignStart)
	sub.SetWrap(true)
	box.Append(sub)

	pw := gtk.NewPasswordEntry()
	pw.SetShowPeekIcon(true)
	pw.AddCSSClass("settings-entry")
	box.Append(pw)

	var pwConfirm *gtk.PasswordEntry
	if creating {
		pwConfirm = gtk.NewPasswordEntry()
		pwConfirm.SetShowPeekIcon(true)
		pwConfirm.AddCSSClass("settings-entry")
		box.Append(pwConfirm)
		pw.Object.SetObjectProperty("placeholder-text", "Master password")
		pwConfirm.Object.SetObjectProperty("placeholder-text", "Confirm master password")
	} else {
		pw.Object.SetObjectProperty("placeholder-text", "Master password")
	}

	errLbl := gtk.NewLabel("")
	errLbl.AddCSSClass("vault-error-label")
	errLbl.SetHAlign(gtk.AlignStart)
	errLbl.SetVisible(false)
	box.Append(errLbl)

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetHAlign(gtk.AlignEnd)
	box.Append(btnBox)

	btnCancel := gtk.NewButtonWithLabel("Cancel")
	btnCancel.AddCSSClass("sidebar-btn")
	btnCancel.ConnectClicked(func() { dialog.Close() })
	btnBox.Append(btnCancel)

	actionText := "Unlock"
	if creating {
		actionText = "Create Vault"
	}
	btnOK := gtk.NewButton()
	btnOK.AddCSSClass("workspace-action-btn")
	lblOK := gtk.NewLabel(actionText)
	lblOK.AddCSSClass("workspace-action-label")
	btnOK.SetChild(lblOK)
	btnBox.Append(btnOK)

	showErr := func(msg string) {
		errLbl.SetLabel(msg)
		errLbl.SetVisible(true)
	}

	submit := func() {
		master := pw.Text()
		if master == "" {
			showErr("Password cannot be empty.")
			return
		}
		if creating {
			if pwConfirm.Text() != master {
				showErr("Passwords do not match.")
				return
			}
			if err := v.Create(master); err != nil {
				showErr(err.Error())
				return
			}
		} else {
			if err := v.Unlock(master); err != nil {
				showErr("Wrong master password.")
				return
			}
		}
		startVaultAutoLock()
		dialog.Close()
		if onUnlocked != nil {
			onUnlocked(v)
		}
	}

	pw.ConnectActivate(func() {
		if creating {
			pwConfirm.GrabFocus()
		} else {
			submit()
		}
	})
	if pwConfirm != nil {
		pwConfirm.ConnectActivate(submit)
	}
	btnOK.ConnectClicked(submit)

	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		if keyval == gdk.KEY_Escape {
			dialog.Close()
			return true
		}
		return false
	})
	dialog.AddController(keyCtrl)

	dialog.Present()
	pw.GrabFocus()
}

// LockVaultNow locks the vault and gives quiet feedback via the title.
func (tw *TerminalWindow) LockVaultNow() {
	appVault().Lock()
}

package window

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

var keyNameAliases = map[string]string{
	"esc":        "Escape",
	"escape":     "Escape",
	"enter":      "Return",
	"return":     "Return",
	"backspace":  "BackSpace",
	"tab":        "Tab",
	"space":      "space",
	"spacebar":   "space",
	"pgdn":       "Page_Down",
	"pagedown":   "Page_Down",
	"page_down":  "Page_Down",
	"pgup":       "Page_Up",
	"pageup":     "Page_Up",
	"page_up":    "Page_Up",
	"home":       "Home",
	"end":        "End",
	"del":        "Delete",
	"delete":     "Delete",
	"ins":        "Insert",
	"insert":     "Insert",
	"up":         "Up",
	"down":       "Down",
	"left":       "Left",
	"right":      "Right",
	"plus":       "plus",
	"minus":      "minus",
	"equal":      "equal",
	"equals":     "equal",
}

// ParseShortcut parses a shortcut string (e.g. "ctrl+shift+n") into a GDK keyval and modifiers.
func ParseShortcut(shortcutStr string) (uint, gdk.ModifierType, bool) {
	shortcutStr = strings.ReplaceAll(shortcutStr, " ", "")
	parts := strings.Split(shortcutStr, "+")
	if len(parts) == 0 {
		return 0, 0, false
	}

	var mods gdk.ModifierType
	var keyval uint
	hasKey := false

	for _, p := range parts {
		low := strings.ToLower(p)
		switch low {
		case "ctrl", "control", "ctl":
			mods |= gdk.ControlMask
		case "shift", "shft":
			mods |= gdk.ShiftMask
		case "alt", "mod1":
			mods |= gdk.AltMask
		case "super", "meta", "win", "logo", "mod4":
			mods |= gdk.SuperMask
		default:
			var targetName string
			if alias, exists := keyNameAliases[low]; exists {
				targetName = alias
			} else {
				targetName = p
			}

			kv := gdk.KeyvalFromName(targetName)
			if kv == 16777215 { // VoidSymbol / Not found
				if len(targetName) > 0 {
					capitalized := strings.ToUpper(targetName[:1]) + targetName[1:]
					kv = gdk.KeyvalFromName(capitalized)
				}
			}

			if kv != 16777215 {
				keyval = uint(kv)
				hasKey = true
			}
		}
	}

	return keyval, mods, hasKey
}

// MatchShortcut checks if the GDK key event's keyval and state modifiers match the target shortcut string.
func MatchShortcut(eventKeyval uint, eventState gdk.ModifierType, shortcutStr string) bool {
	if shortcutStr == "" {
		return false
	}
	shortcutKeyval, shortcutMods, ok := ParseShortcut(shortcutStr)
	if !ok {
		return false
	}

	// Filter modifier masks to exclude lock keys (CapsLock, NumLock)
	eventMods := eventState & (gdk.ControlMask | gdk.ShiftMask | gdk.AltMask | gdk.SuperMask)
	shortcutMods = shortcutMods & (gdk.ControlMask | gdk.ShiftMask | gdk.AltMask | gdk.SuperMask)

	if eventMods != shortcutMods {
		return false
	}

	return gdk.KeyvalToLower(eventKeyval) == gdk.KeyvalToLower(shortcutKeyval)
}

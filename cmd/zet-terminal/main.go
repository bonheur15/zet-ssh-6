//go:build linux

package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/window"
)

func main() {
	app := gtk.NewApplication("com.example.zetterminal", 0)

	cfg := config.LoadConfig()

	app.ConnectActivate(func() {
		// Prefer dark theme for window decorations
		settings := gtk.SettingsGetDefault()
		if settings != nil {
			settings.SetObjectProperty("gtk-application-prefer-dark-theme", true)
		}

		// Load Dynamic Theme CSS
		window.InitGlobalCSS(cfg)

		// Create a new terminal window
		window.NewTerminalWindow(app, cfg)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

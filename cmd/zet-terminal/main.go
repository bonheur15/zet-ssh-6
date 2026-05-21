//go:build linux

package main

import (
	"os"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/window"
)

func main() {
	app := gtk.NewApplication("com.example.zetterminal", 0)

	cfg := config.LoadConfig()

	// Parse command-line arguments for initial directory *before* running the app
	initialDir := ""
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--working-directory=") {
			dir := strings.TrimPrefix(arg, "--working-directory=")
			if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
				initialDir = dir
				break
			}
		} else if !strings.HasPrefix(arg, "-") {
			if stat, err := os.Stat(arg); err == nil && stat.IsDir() {
				initialDir = arg
				break
			}
		}
	}

	app.ConnectActivate(func() {
		// Prefer dark theme for window decorations
		settings := gtk.SettingsGetDefault()
		if settings != nil {
			settings.SetObjectProperty("gtk-application-prefer-dark-theme", true)
		}

		// Load Dynamic Theme CSS
		window.InitGlobalCSS(cfg)

		// Create a new terminal window
		window.NewTerminalWindow(app, cfg, "", initialDir)
	})

	// Pass only the application name to app.Run to prevent GLib/GIO from trying to
	// parse directory/file arguments which causes the "This application can not open files" crash.
	if code := app.Run([]string{os.Args[0]}); code > 0 {
		os.Exit(code)
	}
}

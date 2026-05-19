//go:build linux

package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/theme"
	"zet-terminal/internal/window"
)

func main() {
	app := gtk.NewApplication("com.example.zetterminal", 0)

	cfg := config.LoadConfig()

	app.ConnectActivate(func() {
		// Load Theme CSS
		cssProvider := gtk.NewCSSProvider()
		cssProvider.LoadFromData(theme.TerminalCSS)
		gtk.StyleContextAddProviderForDisplay(
			gdk.DisplayGetDefault(),
			cssProvider,
			gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
		)

		// Create a new terminal window
		window.NewTerminalWindow(app, cfg)
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

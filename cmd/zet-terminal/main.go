//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"zet-terminal/internal/config"
	"zet-terminal/internal/window"
)

func main() {
	// Detach from the calling terminal so the shell gets control back
	// immediately and closing the calling terminal won't kill this app.
	if os.Getenv("ZET_DAEMON") != "1" {
		exe, err := os.Executable()
		if err != nil {
			exe = os.Args[0]
		}
		cmd := exec.Command(exe, os.Args[1:]...)
		cmd.Env = append(os.Environ(), "ZET_DAEMON=1")
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start zet-terminal: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Unset ZET_DAEMON so that child shell instances spawned inside VTE
	// do not inherit it and can spawn sub-daemons if needed.
	os.Unsetenv("ZET_DAEMON")

	app := gtk.NewApplication("com.example.zetterminal", gio.ApplicationHandlesCommandLine)

	cfg := config.LoadConfig()

	var isInitialized bool

	app.ConnectStartup(func() {
		// Prefer dark theme for window decorations
		settings := gtk.SettingsGetDefault()
		if settings != nil {
			settings.SetObjectProperty("gtk-application-prefer-dark-theme", true)
		}

		// Load Dynamic Theme CSS
		window.InitGlobalCSS(cfg)
		isInitialized = true
	})

	app.ConnectCommandLine(func(cmdLine *gio.ApplicationCommandLine) int {
		args := cmdLine.Arguments()
		cwd := cmdLine.Cwd()

		// Reload configuration to ensure we have the latest tab groups/state
		cfg = config.LoadConfig()

		if !isInitialized {
			window.InitGlobalCSS(cfg)
			isInitialized = true
		}

		// Parse the arguments for directory
		dir := ""
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "--working-directory=") {
				d := strings.TrimPrefix(arg, "--working-directory=")
				if !filepath.IsAbs(d) {
					d = filepath.Join(cwd, d)
				}
				if stat, err := os.Stat(d); err == nil && stat.IsDir() {
					dir = d
					break
				}
			} else if !strings.HasPrefix(arg, "-") {
				d := arg
				if !filepath.IsAbs(d) {
					d = filepath.Join(cwd, d)
				}
				if stat, err := os.Stat(d); err == nil && stat.IsDir() {
					dir = d
					break
				}
			}
		}

		if dir != "" {
			activeCount := window.ActiveWindowsCount()
			if activeCount > 0 {
				tw := window.GetFirstActiveWindow()
				groupID := ""
				if tw != nil {
					groupID = tw.GetActiveTabGroupID()
				}
				if groupID == "" && len(cfg.TabGroups) > 0 {
					groupID = cfg.TabGroups[0].ID
				}
				if groupID == "" {
					groupID = "group-general"
				}

				// Create the new tab synced to all windows
				tabID := ""
				if tw != nil {
					tabID = tw.CreateNewTabInGroup(groupID, dir, nil)
				} else {
					tabID = "tab-" + strconv.FormatInt(time.Now().UnixNano(), 10)
					if len(cfg.TabGroups) > 0 {
						cfg.TabGroups[0].Tabs = append(cfg.TabGroups[0].Tabs, config.TabConfig{
							ID:   tabID,
							Name: "Console",
						})
						_ = config.SaveConfig(cfg)
					}
				}

				// Open a new terminal window showing that tab
				window.NewTerminalWindow(app, cfg, tabID, dir)
			} else {
				// No active windows, start fresh with the new tab in that dir
				tabID := "tab-" + strconv.FormatInt(time.Now().UnixNano(), 10)
				if len(cfg.TabGroups) == 0 {
					cfg.TabGroups = []config.GroupConfig{
						{
							ID:        "group-general",
							Name:      "General Workspace",
							Collapsed: false,
							Tabs:      []config.TabConfig{},
						},
					}
				}
				cfg.TabGroups[0].Tabs = append(cfg.TabGroups[0].Tabs, config.TabConfig{
					ID:   tabID,
					Name: "Console",
				})
				_ = config.SaveConfig(cfg)

				window.NewTerminalWindow(app, cfg, tabID, dir)
			}
		} else {
			// No directory argument
			activeCount := window.ActiveWindowsCount()
			if activeCount > 0 {
				// Focus the existing window
				window.FocusActiveWindow()
			} else {
				// Open first terminal window (loads default/restored tabs)
				window.NewTerminalWindow(app, cfg, "", "")
			}
		}

		return 0
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

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

type launchRequest struct {
	WorkingDir     string
	ForceNewWindow bool
}

func parseLaunchRequest(args []string, cwd string) launchRequest {
	req := launchRequest{}

	resolveDir := func(candidate string) string {
		if candidate == "" {
			return ""
		}
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(cwd, candidate)
		}
		if stat, err := os.Stat(candidate); err == nil && stat.IsDir() {
			return candidate
		}
		return ""
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--new-window":
			req.ForceNewWindow = true
		case arg == "--working-directory" && i+1 < len(args):
			if dir := resolveDir(args[i+1]); dir != "" {
				req.WorkingDir = dir
				req.ForceNewWindow = true
			}
			i++
		case strings.HasPrefix(arg, "--working-directory="):
			if dir := resolveDir(strings.TrimPrefix(arg, "--working-directory=")); dir != "" {
				req.WorkingDir = dir
				req.ForceNewWindow = true
			}
		case !strings.HasPrefix(arg, "-"):
			if dir := resolveDir(arg); dir != "" {
				req.WorkingDir = dir
				req.ForceNewWindow = true
			}
		}
	}

	return req
}

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

		req := parseLaunchRequest(args, cwd)
		dir := req.WorkingDir

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
				newWin := window.NewTerminalWindow(app, cfg, tabID, dir)
				if newWin != nil {
					newWin.Win.Present()
				}
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

				newWin := window.NewTerminalWindow(app, cfg, tabID, dir)
				if newWin != nil {
					newWin.Win.Present()
				}
			}
		} else {
			activeCount := window.ActiveWindowsCount()
			if req.ForceNewWindow {
				newWin := window.NewTerminalWindow(app, cfg, "", cwd)
				if newWin != nil {
					newWin.Win.Present()
				}
			} else if activeCount > 0 {
				// Focus the existing window
				window.FocusActiveWindow()
			} else {
				// Open first terminal window (loads default/restored tabs)
				newWin := window.NewTerminalWindow(app, cfg, "", "")
				if newWin != nil {
					newWin.Win.Present()
				}
			}
		}

		return 0
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

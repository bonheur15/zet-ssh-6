# Zet Terminal

Zet Terminal is a modern GTK4 + VTE terminal for Linux.

It is built to feel fast, clean, and practical:

- Modern window chrome and settings UI
- Multi-window workflow with synced workspace state
- Sidebar for tabs, groups, and past commands
- Configurable keyboard shortcuts
- Theme accents and terminal palette presets
- "Open in Zet Terminal" desktop action for folders

## Build

Requirements:

- Go
- GTK4
- VTE for GTK4

Build with:

```bash
make build
```

Run with:

```bash
./bin/zet-terminal
```

## Install

Install the app and desktop entries with:

```bash
make install
```

## Highlights

- Open a new window in any directory
- Organize tabs into groups
- Pin important tabs
- Search recent commands from the sidebar
- Tune fonts, colors, glow, window defaults, and shortcuts

## Configuration

Zet stores its config in:

```text
~/.config/zet-terminal/config.json
```

You can manage most settings directly from the in-app preferences window.

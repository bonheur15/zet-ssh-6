package terminal

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ReadShellHistory reads the last N commands from the configured shell's history file.
func ReadShellHistory(shellPath string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var historyFile string
	isZsh := false

	if strings.Contains(shellPath, "zsh") {
		historyFile = filepath.Join(home, ".zsh_history")
		isZsh = true
	} else {
		// Default to bash history
		historyFile = filepath.Join(home, ".bash_history")
	}

	file, err := os.Open(historyFile)
	if err != nil {
		return nil
	}
	defer file.Close()

	var rawLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		rawLines = append(rawLines, line)
	}

	// We only care about the last 100 lines to parse the most recent history items.
	startIdx := len(rawLines) - 100
	if startIdx < 0 {
		startIdx = 0
	}

	var parsedCmds []string
	seen := make(map[string]bool)

	// Process lines from newest to oldest
	for i := len(rawLines) - 1; i >= startIdx; i-- {
		line := rawLines[i]
		cmd := strings.TrimSpace(line)
		if cmd == "" {
			continue
		}

		if isZsh {
			// Zsh history format is usually: `: 1621345678:0;command`
			if strings.HasPrefix(cmd, ":") {
				if semiIdx := strings.Index(cmd, ";"); semiIdx != -1 {
					cmd = cmd[semiIdx+1:]
				}
			}
		}

		cmd = strings.TrimSpace(cmd)
		if len(cmd) < 2 || len(cmd) > 60 {
			continue
		}

		cmdLower := strings.ToLower(cmd)
		if cmdLower == "ls" || cmdLower == "clear" || cmdLower == "top" || cmdLower == "exit" {
			continue
		}

		// Avoid duplicates in the output list
		if seen[cmd] {
			continue
		}
		seen[cmd] = true

		parsedCmds = append(parsedCmds, cmd)
		if len(parsedCmds) >= 10 { // Max 10 items in history UI
			break
		}
	}

	return parsedCmds
}

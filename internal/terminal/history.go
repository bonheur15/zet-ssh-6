package terminal

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func recentHistoryWindow(limit int) int {
	if limit <= 0 {
		limit = 50
	}
	window := limit * 20
	if window < 200 {
		window = 200
	}
	if window > 5000 {
		window = 5000
	}
	return window
}

// ReadShellHistory reads a bounded set of recent commands from the configured shell history file.
func ReadShellHistory(shellPath string, limit int) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	if limit <= 0 {
		limit = 50
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

	windowSize := recentHistoryWindow(limit)
	rawLines := make([]string, 0, windowSize)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawLines = append(rawLines, scanner.Text())
		if len(rawLines) > windowSize {
			copy(rawLines, rawLines[1:])
			rawLines = rawLines[:windowSize]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil
	}

	var parsedCmds []string
	seen := make(map[string]bool)

	// Process lines from newest to oldest
	for i := len(rawLines) - 1; i >= 0; i-- {
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
		if len(parsedCmds) >= limit {
			break
		}
	}

	return parsedCmds
}

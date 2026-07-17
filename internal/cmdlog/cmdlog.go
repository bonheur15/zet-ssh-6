// Package cmdlog records every command Zet-SSH generates (connects,
// transfers, tunnels) so the user can audit and copy the exact
// invocation behind each UI action.
package cmdlog

import (
	"sync"
	"time"
)

const maxEntries = 300

// Entry is one logged action.
type Entry struct {
	Time    time.Time
	Label   string // human action, e.g. "Connect prod-api"
	Command string // exact shell command
}

var (
	mu        sync.Mutex
	entries   []Entry
	listeners []func()
)

// Add appends an entry and notifies listeners (on the caller's goroutine).
func Add(label, command string) {
	mu.Lock()
	entries = append(entries, Entry{Time: time.Now(), Label: label, Command: command})
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}
	cbs := append([]func(){}, listeners...)
	mu.Unlock()
	for _, fn := range cbs {
		fn()
	}
}

// Entries returns a snapshot, newest first.
func Entries() []Entry {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Entry, len(entries))
	for i, e := range entries {
		out[len(entries)-1-i] = e
	}
	return out
}

// Subscribe registers a change callback.
func Subscribe(fn func()) {
	mu.Lock()
	listeners = append(listeners, fn)
	mu.Unlock()
}

// Clear empties the log.
func Clear() {
	mu.Lock()
	entries = nil
	cbs := append([]func(){}, listeners...)
	mu.Unlock()
	for _, fn := range cbs {
		fn()
	}
}

// Package snippets implements the Zet-SSH command library: reusable
// commands with ${VAR} placeholders that can be filled interactively
// or from the encrypted vault.
package snippets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Snippet is one saved command.
type Snippet struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

var varPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// VarNames returns the unique ${VAR} placeholders in order of appearance.
func (s *Snippet) VarNames() []string {
	seen := map[string]bool{}
	var names []string
	for _, m := range varPattern.FindAllStringSubmatch(s.Command, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			names = append(names, m[1])
		}
	}
	return names
}

// Expand substitutes placeholders using resolve. Unresolved variables
// are returned in missing.
func (s *Snippet) Expand(resolve func(name string) (string, bool)) (out string, missing []string) {
	out = varPattern.ReplaceAllStringFunc(s.Command, func(m string) string {
		name := varPattern.FindStringSubmatch(m)[1]
		if val, ok := resolve(name); ok {
			return val
		}
		missing = append(missing, name)
		return m
	})
	return out, missing
}

// ─── Store ───

type Store struct {
	path     string
	Snippets []*Snippet `json:"snippets"`
}

var store *Store

// Load returns the global snippet store bound to the config dir.
func Load(configDir string) *Store {
	if store == nil {
		store = &Store{path: filepath.Join(configDir, "snippets.json")}
		raw, err := os.ReadFile(store.path)
		if err == nil {
			_ = json.Unmarshal(raw, store)
		}
	}
	return store
}

func (s *Store) Save() error {
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Get(id string) *Snippet {
	for _, sn := range s.Snippets {
		if sn.ID == id {
			return sn
		}
	}
	return nil
}

func (s *Store) Upsert(sn *Snippet) {
	if sn.ID == "" {
		sn.ID = fmt.Sprintf("snip-%d", time.Now().UnixNano())
	}
	for i, existing := range s.Snippets {
		if existing.ID == sn.ID {
			s.Snippets[i] = sn
			_ = s.Save()
			return
		}
	}
	s.Snippets = append(s.Snippets, sn)
	_ = s.Save()
}

func (s *Store) Delete(id string) {
	for i, sn := range s.Snippets {
		if sn.ID == id {
			s.Snippets = append(s.Snippets[:i], s.Snippets[i+1:]...)
			_ = s.Save()
			return
		}
	}
}

// Search returns snippets matching the query against name, command,
// description, and tags (case-insensitive substring), sorted by name.
func (s *Store) Search(query string) []*Snippet {
	query = strings.ToLower(strings.TrimSpace(query))
	var out []*Snippet
	for _, sn := range s.Snippets {
		if query == "" ||
			strings.Contains(strings.ToLower(sn.Name), query) ||
			strings.Contains(strings.ToLower(sn.Command), query) ||
			strings.Contains(strings.ToLower(sn.Description), query) ||
			strings.Contains(strings.ToLower(strings.Join(sn.Tags, " ")), query) {
			out = append(out, sn)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

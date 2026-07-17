// Package sshmgr manages saved SSH connection profiles, tunnels, and
// the exact command lines Zet-SSH generates for them. Everything the
// app runs is expressed as a plain ssh/scp/sftp command so users can
// audit and copy it.
package sshmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AuthMethod selects how a profile authenticates.
type AuthMethod string

const (
	AuthAgent    AuthMethod = "agent"
	AuthKey      AuthMethod = "key"
	AuthPassword AuthMethod = "password"
)

// KnownHostsMode maps to OpenSSH StrictHostKeyChecking values.
type KnownHostsMode string

const (
	KnownHostsStrict KnownHostsMode = "strict" // yes
	KnownHostsPrompt KnownHostsMode = "prompt" // ask (OpenSSH default)
	KnownHostsAccept KnownHostsMode = "accept" // accept-new
)

// Profile is one saved SSH connection.
type Profile struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Host       string         `json:"host"`
	Port       int            `json:"port"`
	User       string         `json:"user"`
	Auth       AuthMethod     `json:"auth"`
	KeyPath    string         `json:"key_path,omitempty"`
	JumpHost   string         `json:"jump_host,omitempty"` // ProxyJump: [user@]host[:port]
	KnownHosts KnownHostsMode `json:"known_hosts,omitempty"`
	Keepalive  int            `json:"keepalive_sec,omitempty"` // ServerAliveInterval
	RemoteDir  string         `json:"remote_dir,omitempty"`
	Ciphers    string         `json:"ciphers,omitempty"`
	Kex        string         `json:"kex,omitempty"`
	ExtraArgs  string         `json:"extra_args,omitempty"` // raw extra ssh args, space separated
	Tags       []string       `json:"tags,omitempty"`
	LastUsed   time.Time      `json:"last_used,omitempty"`
}

// Label is the display name shown in lists.
func (p *Profile) Label() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Target()
}

// Target returns user@host.
func (p *Profile) Target() string {
	if p.User != "" {
		return p.User + "@" + p.Host
	}
	return p.Host
}

// baseOptions builds the ssh option arguments shared by ssh/scp/sftp.
// scpMode uses -P for port instead of -p.
func (p *Profile) baseOptions(scpMode bool) []string {
	var args []string
	if p.Port != 0 && p.Port != 22 {
		if scpMode {
			args = append(args, "-P", strconv.Itoa(p.Port))
		} else {
			args = append(args, "-p", strconv.Itoa(p.Port))
		}
	}
	if p.Auth == AuthKey && p.KeyPath != "" {
		args = append(args, "-i", p.KeyPath)
	}
	if p.JumpHost != "" {
		args = append(args, "-J", p.JumpHost)
	}
	switch p.KnownHosts {
	case KnownHostsStrict:
		args = append(args, "-o", "StrictHostKeyChecking=yes")
	case KnownHostsAccept:
		args = append(args, "-o", "StrictHostKeyChecking=accept-new")
	}
	if p.Keepalive > 0 {
		args = append(args, "-o", fmt.Sprintf("ServerAliveInterval=%d", p.Keepalive))
	}
	if p.Ciphers != "" {
		args = append(args, "-o", "Ciphers="+p.Ciphers)
	}
	if p.Kex != "" {
		args = append(args, "-o", "KexAlgorithms="+p.Kex)
	}
	if p.ExtraArgs != "" {
		args = append(args, strings.Fields(p.ExtraArgs)...)
	}
	return args
}

// SSHArgv returns the full argv for an interactive session.
func (p *Profile) SSHArgv() []string {
	args := []string{"ssh"}
	args = append(args, p.baseOptions(false)...)
	args = append(args, p.Target())
	if p.RemoteDir != "" {
		args = append(args, "-t", fmt.Sprintf("cd %s && exec $SHELL -l", shellQuote(p.RemoteDir)))
	}
	return args
}

// SFTPArgv returns argv for an interactive sftp session.
func (p *Profile) SFTPArgv() []string {
	args := []string{"sftp"}
	for _, a := range p.baseOptions(false) {
		args = append(args, a)
	}
	// sftp uses -P for port; strip a possible "-p N" pair and re-add.
	args = fixPortFlag(args)
	args = append(args, p.Target())
	return args
}

// ExecArgv returns argv that runs a single remote command over a shared
// ControlMaster connection (fast repeated calls for the file browser).
// When batch is true, ssh fails fast instead of prompting (used for
// agent/key auth); password-auth callers pass false and drive the prompt
// with sshpass.
func (p *Profile) ExecArgv(controlPath string, remoteCmd string, batch bool) []string {
	args := []string{"ssh"}
	args = append(args, p.baseOptions(false)...)
	args = append(args,
		"-o", "ControlMaster=auto",
		"-o", "ControlPath="+controlPath,
		"-o", "ControlPersist=120")
	if batch {
		args = append(args, "-o", "BatchMode=yes")
	}
	args = append(args, p.Target(), "--", remoteCmd)
	return args
}

// SCPArgv builds an scp command. remote paths are prefixed with target:.
func (p *Profile) SCPArgv(localPath, remotePath string, upload, recursive bool) []string {
	args := []string{"scp"}
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, p.baseOptions(true)...)
	remote := p.Target() + ":" + remotePath
	if upload {
		args = append(args, localPath, remote)
	} else {
		args = append(args, remote, localPath)
	}
	return args
}

func fixPortFlag(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			out = append(out, "-P", args[i+1])
			i++
			continue
		}
		out = append(out, args[i])
	}
	return out
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.ContainsAny(s, " \t\n'\"\\$`&;|<>(){}*?[]~#") {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}
	return s
}

// WrapSSHPass prefixes argv so the password is fed via `sshpass -e`,
// which reads it from the SSHPASS environment variable — the password
// never appears in the process arguments. Callers must set SSHPASS in
// the command's environment. If password is empty, argv is unchanged.
func WrapSSHPass(argv []string, password string) []string {
	if password == "" {
		return argv
	}
	return append([]string{"sshpass", "-e"}, argv...)
}

// CommandString renders argv into a copy-pastable shell command.
func CommandString(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = shellQuote(a)
	}
	return strings.Join(parts, " ")
}

// ParseSSHCommand parses a pasted "ssh ..." command line into a Profile.
// Supports user@host, -p/-P, -i, -J/ProxyJump, -o options, -L/-R/-D (ignored).
func ParseSSHCommand(cmdline string) (*Profile, bool) {
	fields := strings.Fields(strings.TrimSpace(cmdline))
	if len(fields) == 0 {
		return nil, false
	}
	// Require the command to actually be an ssh invocation so arbitrary
	// selected text (e.g. "ls -la") is not mistaken for a connection.
	base0 := filepath.Base(fields[0])
	if base0 != "ssh" && base0 != "sshpass" {
		return nil, false
	}
	if base0 == "ssh" || base0 == "sshpass" {
		// skip leading sshpass -p xxx / ssh
		for len(fields) > 0 {
			base := filepath.Base(fields[0])
			if base == "ssh" {
				fields = fields[1:]
				break
			}
			if base == "sshpass" {
				fields = fields[1:]
				if len(fields) >= 2 && (fields[0] == "-p" || fields[0] == "-f") {
					fields = fields[2:]
				}
				continue
			}
			break
		}
	}

	p := &Profile{Port: 22, Auth: AuthAgent, KnownHosts: KnownHostsPrompt}
	found := false
	takesValue := map[string]bool{
		"-p": true, "-P": true, "-i": true, "-J": true, "-o": true,
		"-L": true, "-R": true, "-D": true, "-l": true, "-F": true,
		"-c": true, "-m": true, "-b": true, "-e": true, "-w": true,
	}

	for i := 0; i < len(fields); i++ {
		f := fields[i]
		switch {
		case f == "-p" || f == "-P":
			if i+1 < len(fields) {
				if n, err := strconv.Atoi(fields[i+1]); err == nil {
					p.Port = n
				}
				i++
			}
		case f == "-i":
			if i+1 < len(fields) {
				p.KeyPath = fields[i+1]
				p.Auth = AuthKey
				i++
			}
		case f == "-J":
			if i+1 < len(fields) {
				p.JumpHost = fields[i+1]
				i++
			}
		case f == "-l":
			if i+1 < len(fields) {
				p.User = fields[i+1]
				i++
			}
		case f == "-o":
			if i+1 < len(fields) {
				opt := fields[i+1]
				if strings.HasPrefix(opt, "ProxyJump=") {
					p.JumpHost = strings.TrimPrefix(opt, "ProxyJump=")
				}
				if strings.HasPrefix(opt, "ServerAliveInterval=") {
					if n, err := strconv.Atoi(strings.TrimPrefix(opt, "ServerAliveInterval=")); err == nil {
						p.Keepalive = n
					}
				}
				i++
			}
		case strings.HasPrefix(f, "-"):
			if takesValue[f] {
				i++
			}
		case !found:
			target := f
			if at := strings.LastIndex(target, "@"); at != -1 {
				p.User = target[:at]
				target = target[at+1:]
			}
			// host:port form (rare but seen in pasted URIs)
			if strings.HasPrefix(target, "ssh://") {
				target = strings.TrimPrefix(target, "ssh://")
			}
			if h, prt, ok := strings.Cut(target, ":"); ok {
				if n, err := strconv.Atoi(prt); err == nil {
					p.Port = n
					target = h
				}
			}
			p.Host = target
			found = true
		}
	}

	if !found || p.Host == "" {
		return nil, false
	}
	p.Name = p.Host
	return p, true
}

// ─── Profile store ───

type ProfileStore struct {
	path     string
	Profiles []*Profile `json:"profiles"`
}

var profileStore *ProfileStore

// Profiles returns the global profile store, loading it on first use.
func Profiles(configDir string) *ProfileStore {
	if profileStore == nil {
		profileStore = &ProfileStore{path: filepath.Join(configDir, "profiles.json")}
		profileStore.load()
	}
	return profileStore
}

func (s *ProfileStore) load() {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, s)
}

func (s *ProfileStore) Save() error {
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

func (s *ProfileStore) Get(id string) *Profile {
	for _, p := range s.Profiles {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (s *ProfileStore) Upsert(p *Profile) {
	if p.ID == "" {
		p.ID = fmt.Sprintf("prof-%d", time.Now().UnixNano())
	}
	for i, existing := range s.Profiles {
		if existing.ID == p.ID {
			s.Profiles[i] = p
			_ = s.Save()
			return
		}
	}
	s.Profiles = append(s.Profiles, p)
	_ = s.Save()
}

func (s *ProfileStore) Delete(id string) {
	for i, p := range s.Profiles {
		if p.ID == id {
			s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
			_ = s.Save()
			return
		}
	}
}

// Sorted returns profiles sorted by most recently used, then name.
func (s *ProfileStore) Sorted() []*Profile {
	out := append([]*Profile(nil), s.Profiles...)
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].LastUsed.Equal(out[j].LastUsed) {
			return out[i].LastUsed.After(out[j].LastUsed)
		}
		return strings.ToLower(out[i].Label()) < strings.ToLower(out[j].Label())
	})
	return out
}

// MarkUsed stamps LastUsed and persists.
func (s *ProfileStore) MarkUsed(id string) {
	if p := s.Get(id); p != nil {
		p.LastUsed = time.Now()
		_ = s.Save()
	}
}

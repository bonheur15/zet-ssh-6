package sshmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// TunnelType selects the ssh forwarding flag.
type TunnelType string

const (
	TunnelLocal   TunnelType = "local"   // -L listen locally, forward to remote
	TunnelRemote  TunnelType = "remote"  // -R listen remotely, forward to local
	TunnelDynamic TunnelType = "dynamic" // -D SOCKS proxy
)

// TunnelSpec is a saved port-forward definition tied to a profile.
type TunnelSpec struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	ProfileID  string     `json:"profile_id"`
	Type       TunnelType `json:"type"`
	BindHost   string     `json:"bind_host,omitempty"` // listen address, default localhost
	BindPort   int        `json:"bind_port"`
	TargetHost string     `json:"target_host,omitempty"` // forward destination
	TargetPort int        `json:"target_port,omitempty"`
}

// ForwardFlag renders the -L/-R/-D argument value.
func (t *TunnelSpec) ForwardFlag() (flag, value string) {
	bind := t.BindHost
	if bind == "" {
		bind = "localhost"
	}
	switch t.Type {
	case TunnelRemote:
		return "-R", fmt.Sprintf("%s:%d:%s:%d", bind, t.BindPort, t.TargetHost, t.TargetPort)
	case TunnelDynamic:
		return "-D", fmt.Sprintf("%s:%d", bind, t.BindPort)
	default:
		return "-L", fmt.Sprintf("%s:%d:%s:%d", bind, t.BindPort, t.TargetHost, t.TargetPort)
	}
}

// Summary is a short human description, e.g. "L 8080 → db:5432".
func (t *TunnelSpec) Summary() string {
	switch t.Type {
	case TunnelRemote:
		return fmt.Sprintf("R :%d → %s:%d", t.BindPort, t.TargetHost, t.TargetPort)
	case TunnelDynamic:
		return fmt.Sprintf("D SOCKS :%d", t.BindPort)
	default:
		return fmt.Sprintf("L :%d → %s:%d", t.BindPort, t.TargetHost, t.TargetPort)
	}
}

// Argv builds the full background ssh command for this tunnel.
func (t *TunnelSpec) Argv(p *Profile) []string {
	flag, value := t.ForwardFlag()
	args := []string{"ssh", "-N", flag, value,
		"-o", "ExitOnForwardFailure=yes"}
	args = append(args, p.baseOptions(false)...)
	if p.Keepalive == 0 {
		args = append(args, "-o", "ServerAliveInterval=30")
	}
	args = append(args, p.Target())
	return args
}

// ─── Tunnel store + runtime manager ───

type TunnelStore struct {
	path    string
	Tunnels []*TunnelSpec `json:"tunnels"`

	mu       sync.Mutex
	running  map[string]*exec.Cmd
	lastErr  map[string]string
	onChange []func()
}

var tunnelStore *TunnelStore

func Tunnels(configDir string) *TunnelStore {
	if tunnelStore == nil {
		tunnelStore = &TunnelStore{
			path:    filepath.Join(configDir, "tunnels.json"),
			running: map[string]*exec.Cmd{},
			lastErr: map[string]string{},
		}
		tunnelStore.load()
	}
	return tunnelStore
}

func (s *TunnelStore) load() {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, s)
}

func (s *TunnelStore) Save() error {
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

func (s *TunnelStore) Get(id string) *TunnelSpec {
	for _, t := range s.Tunnels {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (s *TunnelStore) Upsert(t *TunnelSpec) {
	if t.ID == "" {
		t.ID = "tun-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	for i, existing := range s.Tunnels {
		if existing.ID == t.ID {
			s.Tunnels[i] = t
			_ = s.Save()
			return
		}
	}
	s.Tunnels = append(s.Tunnels, t)
	_ = s.Save()
}

func (s *TunnelStore) Delete(id string) {
	s.Stop(id)
	for i, t := range s.Tunnels {
		if t.ID == id {
			s.Tunnels = append(s.Tunnels[:i], s.Tunnels[i+1:]...)
			_ = s.Save()
			return
		}
	}
}

// OnChange registers a callback fired (from any goroutine) whenever a
// tunnel starts, stops, or dies. UI must marshal back to main loop.
func (s *TunnelStore) OnChange(fn func()) {
	s.mu.Lock()
	s.onChange = append(s.onChange, fn)
	s.mu.Unlock()
}

func (s *TunnelStore) notify() {
	s.mu.Lock()
	cbs := append([]func(){}, s.onChange...)
	s.mu.Unlock()
	for _, fn := range cbs {
		fn()
	}
}

// IsRunning reports whether the tunnel process is alive.
func (s *TunnelStore) IsRunning(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.running[id]
	return ok
}

// LastError returns the last failure output for a tunnel, if any.
func (s *TunnelStore) LastError(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr[id]
}

// Start launches the tunnel in the background. When password is non-empty
// the connection is authenticated via sshpass (fed through the SSHPASS
// env var, so it never appears in the process arguments). The generated
// command (without the password) is returned so callers can log it.
func (s *TunnelStore) Start(t *TunnelSpec, p *Profile, password string) ([]string, error) {
	s.mu.Lock()
	if _, ok := s.running[t.ID]; ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("tunnel already running")
	}
	s.mu.Unlock()

	argv := t.Argv(p)
	wrapped := WrapSSHPass(argv, password)
	cmd := exec.Command(wrapped[0], wrapped[1:]...)
	if password != "" {
		cmd.Env = append(os.Environ(), "SSHPASS="+password)
	}
	// New session so it survives independently and we can signal it cleanly.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	var stderr limitedBuffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return argv, err
	}

	s.mu.Lock()
	s.running[t.ID] = cmd
	delete(s.lastErr, t.ID)
	s.mu.Unlock()
	s.notify()

	go func() {
		err := cmd.Wait()
		s.mu.Lock()
		delete(s.running, t.ID)
		if err != nil {
			msg := stderr.String()
			if msg == "" {
				msg = err.Error()
			}
			s.lastErr[t.ID] = msg
		}
		s.mu.Unlock()
		s.notify()
	}()

	return argv, nil
}

// Stop terminates a running tunnel.
func (s *TunnelStore) Stop(id string) {
	s.mu.Lock()
	cmd := s.running[id]
	s.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}
}

// StopAll terminates every running tunnel (used at shutdown).
func (s *TunnelStore) StopAll() {
	s.mu.Lock()
	cmds := make([]*exec.Cmd, 0, len(s.running))
	for _, c := range s.running {
		cmds = append(cmds, c)
	}
	s.mu.Unlock()
	for _, c := range cmds {
		if c.Process != nil {
			_ = c.Process.Signal(syscall.SIGTERM)
		}
	}
}

// limitedBuffer keeps only the last ~4KB of output.
type limitedBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	if len(b.buf) > 4096 {
		b.buf = b.buf[len(b.buf)-4096:]
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}
